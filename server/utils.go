package server

import (
	"bisq-add-on/api"
	"errors"
	"go.uber.org/zap"
	"net/http"
)

func handleSimpleResponse(w http.ResponseWriter, status int, msg string) {
	w.WriteHeader(status)
	_, _ = w.Write([]byte(msg))
}

func (s *Service) matchOffers(offer *UserOffer) (bool, error) {
	s.logger.Info("server.utils.matchOffers: searching for match offer...")

	offers := s.buyOffers

	if offer.Direction == "BUY" {
		offers = s.sellOffers
	}

	for account, savedOffer := range offers {
		if account != offer.AccountName {
			if savedOffer.Token == offer.Token && savedOffer.Amount == offer.Amount && savedOffer.Price == offer.Price {
				s.logger.Info("server.utils.matchOffers: found offer to match.")

				buyOffer := savedOffer
				sellOffer := offer
				if offer.Direction == "BUY" {
					buyOffer, sellOffer = sellOffer, buyOffer
				}

				err := s.handleMatchedOffers(buyOffer, sellOffer)
				if err != nil {
					s.logger.Error("server.utils.matchOffer: server.handleMatchedOffers failure.")
					return false, err
				}

				s.logger.Info("server.utils.matchOffers: matched offers successfully.")

				return true, nil
			}
		}
	}

	s.logger.Info("server.utils.matchOffers: no offers to match was found.")
	return false, nil
}

func (s *Service) handleMatchedOffers(buyOffer *UserOffer, sellOffer *UserOffer) error {
	s.logger.Info("server.utils.handleMatchedOffers: new incoming offers...")

	s.mu.Lock()
	s.matchedBuyAccounts[sellOffer.AccountName] = buyOffer.AccountName
	s.matchedSellAccounts[buyOffer.AccountName] = sellOffer.AccountName
	s.mu.Unlock()

	buyAccount := api.PaymentAccount{
		Name:                  buyOffer.AccountName,
		TradeCurrencies:       []string{"BTC", "ETH"},
		PaymentMethod:         "BLOCK_CHAINS",
		ID:                    "",
		Details:               buyOffer.EthereumWallet,
		SelectedTradeCurrency: "ETH",
	}

	respBuyAcc, err := api.RegisterPaymentAccounts(s.logger, s.client, &buyAccount)
	if err != nil {
		s.logger.Error("server.utils.handleMatchedOffers: api.RegisterPaymentAccounts failure.")
		return err
	}

	s.logger.Info("server.utils.handleMatchedOffers: buy account parsed successfully.")

	s.mu.Lock()
	s.accounts[respBuyAcc.Name] = respBuyAcc
	s.ethereumWallets[respBuyAcc.Name] = respBuyAcc.Details
	s.mu.Unlock()

	offerToCreate := api.OfferToCreate{
		FundUsingBisqWallet:       true,
		OfferID:                   "",
		AccountID:                 respBuyAcc.ID,
		Direction:                 "BUY",
		PriceType:                 "",
		MarketPair:                "btc_eth",
		PercentageFromMarketPrice: 0,
		FixedPrice:                buyOffer.Price,
		Amount:                    buyOffer.Amount,
		MinAmount:                 buyOffer.Amount,
		BuyerSecurityDeposit:      1,
	}

	offerDetails, err := api.PublishOffer(s.logger, s.client, &offerToCreate)
	if err != nil {
		s.logger.Error("server.utils.handleMatchedOffers: api.PublishOffer failure.")
		return err
	}

	s.logger.Info("server.utils.handleMatchedOffers: published buy offer successfully.")

	sellAccount := api.PaymentAccount{
		Name:                  sellOffer.AccountName,
		TradeCurrencies:       []string{"BTC", "ETH"},
		PaymentMethod:         "BLOCK_CHAINS",
		ID:                    "",
		Details:               sellOffer.EthereumWallet,
		SelectedTradeCurrency: "ETH",
	}

	respSellAcc, err := api.RegisterPaymentAccounts(s.logger, s.client, &sellAccount)
	if err != nil {
		s.logger.Error("server.utils.handleMatchedOffers: api.RegisterPaymentAccounts failure.")
		return err
	}

	s.logger.Info("server.utils.handleMatchedOffers: sell account parsed successfully.")

	s.mu.Lock()
	s.accounts[respSellAcc.Name] = respSellAcc
	s.ethereumWallets[respSellAcc.Name] = respSellAcc.Details
	s.mu.Unlock()

	offerToTake := api.OfferToTake{
		PaymentAccountID: respSellAcc.ID,
		Amount:           offerToCreate.Amount,
	}

	tradeDetails, err := api.TakeOffer(s.logger, s.client, &offerToTake)
	if err != nil {
		s.logger.Error("server.utils.handleMatchedOffers: api.TakeOffer failure.")
		return err
	}

	s.logger.Info("server.utils.handleMatchedOffers: took buy order successfully.")

	s.mu.Lock()
	s.buyTrades[respBuyAcc.Name] = tradeDetails
	s.sellTrades[respSellAcc.Name] = tradeDetails
	s.mu.Unlock()

	return nil
}

func (s *Service) checkTransaction(transactionID string, buySide string, sellSide string) (bool, error) {
	s.logger.Info("server.utils.checkTransaction: new incoming transaction...")

	transactionInfo, err := api.GetTxInfo(s.logger, s.client, transactionID)
	if err != nil {
		return true, err
	}

	if !transactionInfo.Success {
		return false, errors.New("transaction did not complete successfully")
	}

	fromAddr := s.ethereumWallets[buySide]
	toAddr := s.ethereumWallets[sellSide]

	if transactionInfo.From != fromAddr {
		return false, errors.New("transaction sender address is incorrect")
	}

	if transactionInfo.To != toAddr {
		return false, errors.New("transaction receiver address is incorrect")
	}

	return true, nil
}

func (s *Service) handleSuccessfulTransaction(trade *api.TradeDetails) error {
	s.logger.Info("server.utils.handleSuccessfulTransaction: new incoming trade...")
	err := api.PaymentStarted(s.logger, s.client, trade)
	if err != nil {
		s.logger.Error("server.utils.handleSuccessfulTransaction: api.PaymentStarted failure.")
		return err
	}

	err = api.PaymentReceived(s.logger, s.client, trade)
	if err != nil {
		s.logger.Error("server.utils.handleSuccessfulTransaction: api.PaymentReceived failure.")
		return err
	}

	return nil
}