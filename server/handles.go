package server

import (
	"bisq-add-on/api"
	"encoding/json"
	"go.uber.org/zap"
	"net/http"
	"sync"
)

type Service struct {
	logger *zap.Logger
	client *http.Client
	mu     *sync.Mutex

	buyOffers  map[string]*UserOffer
	sellOffers map[string]*UserOffer
	accounts   map[string]*api.PaymentAccount

	buyTrades  map[string]*api.TradeDetails
	sellTrades map[string]*api.TradeDetails

	matchedBuyAccounts  map[string]string
	matchedSellAccounts map[string]string

	ethereumWallets map[string]string
	transactionIDs  map[string]string
}

func initLogger() *zap.Logger {
	logger, _ := zap.NewDevelopment()
	return logger
}

func InitService() *Service {
	s := Service{
		logger: initLogger(),
		client: api.InitClient(),
		mu:     &sync.Mutex{},

		buyOffers:  make(map[string]*UserOffer),
		sellOffers: make(map[string]*UserOffer),
		accounts:   make(map[string]*api.PaymentAccount),

		buyTrades:  make(map[string]*api.TradeDetails),
		sellTrades: make(map[string]*api.TradeDetails),

		matchedBuyAccounts:  make(map[string]string),
		matchedSellAccounts: make(map[string]string),

		ethereumWallets: make(map[string]string),
		transactionIDs:  make(map[string]string),
	}

	return &s
}

type UserOffer struct {
	AccountName string `json:"accountName"`
	Token       string `json:"token"`

	Price     int64  `json:"price"`
	Amount    int64  `json:"amount"`
	Direction string `json:"direction"`

	EthereumWallet string `json:"ethereumWallet"`
}

func (s *Service) BuyHandle(w http.ResponseWriter, r *http.Request) {
	s.logger.Info("server.handles.BuyHandle: received new request.")

	var offer UserOffer
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&offer)
	if err != nil {
		s.logger.Error("server.handles.BuyHandle: json decoder failure.", zap.Error(err))
		handleSimpleResponse(w, http.StatusInternalServerError, "json decoder failure.")
		return
	}

	matched, err := s.matchOffers(&offer)
	if err != nil {
		s.logger.Error("server.handles.BuyHandle: server.matchOffers failure.")
		handleSimpleResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	if matched {
		handleSimpleResponse(w, http.StatusOK, "Your offer was matched successfully.")
		return
	}

	s.mu.Lock()
	s.buyOffers[offer.AccountName] = &offer
	s.mu.Unlock()

	handleSimpleResponse(w, http.StatusOK, "Your offer was saved successfully.")
}

func (s *Service) SellHandle(w http.ResponseWriter, r *http.Request) {
	s.logger.Info("server.handles.SellHandle: received new request.")

	var offer UserOffer
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&offer)
	if err != nil {
		s.logger.Error("server.handles.SellHandle: json decoder failure.", zap.Error(err))
		handleSimpleResponse(w, http.StatusInternalServerError, "json decoder failure.")
		return
	}

	matched, err := s.matchOffers(&offer)
	if err != nil {
		s.logger.Error("server.handles.SellHandle: server.matchOffers failure.")
		handleSimpleResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	if matched {
		handleSimpleResponse(w, http.StatusOK, "Your offer was matched successfully.")
		return
	}

	s.mu.Lock()
	s.sellOffers[offer.AccountName] = &offer
	s.mu.Unlock()

	handleSimpleResponse(w, http.StatusOK, "Your offer was saved successfully.")
}

func (s *Service) CheckOfferHandle(w http.ResponseWriter, r *http.Request) {
	s.logger.Info("server.handles.CheckOfferHandle: received new request.")

	accountNames, ok := r.URL.Query()["account"]
	if !ok || len(accountNames) == 0 {
		s.logger.Info("server.handles.CheckOfferHandle: 'account' parameter is missing.")
		handleSimpleResponse(w, http.StatusBadRequest, "'account' parameter is missing.")
		return
	}
	accountName := accountNames[0]
	counterAccountName := s.matchedSellAccounts[accountName]
	ethereumWallet := s.ethereumWallets[counterAccountName]

	handleSimpleResponse(w, http.StatusOK, ethereumWallet)
}

type MoneySentRequest struct {
	TransactionID string `json:"transactionID"`
}

func (s *Service) MoneySentHandle(w http.ResponseWriter, r *http.Request) {
	s.logger.Info("server.handles.MoneySentHandle: received new request.")

	accountNames, ok := r.URL.Query()["account"]
	if !ok || len(accountNames) == 0 {
		s.logger.Info("server.handles.MoneySentHandle: 'account' parameter is missing.")
		handleSimpleResponse(w, http.StatusBadRequest, "'account' parameter is missing.")
		return
	}
	accountName := accountNames[0]
	counterAccountName := s.matchedSellAccounts[accountName]

	var req MoneySentRequest
	dec := json.NewDecoder(r.Body)
	err := dec.Decode(&req)
	if err != nil {
		s.logger.Error("server.handles.MoneySentHandle: json decoder failure.", zap.Error(err))
		handleSimpleResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	ok, err = s.checkTransaction(req.TransactionID, accountName, counterAccountName)
	if ok && err != nil {
		s.logger.Error("server.handles.MoneySentHandle: server.checkTransaction failure.")
		handleSimpleResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	if !ok && err != nil {
		s.logger.Error("server.handles.MoneySentHandle: invalid transaction.", zap.Error(err))
		handleSimpleResponse(w, http.StatusBadRequest, err.Error())
		return
	}

	s.logger.Info("server.handles.MoneySentHandle: transaction is valid, proceed to finishing trade...")

	err = s.handleSuccessfulTransaction(s.buyTrades[accountName])
	if err != nil {
		s.logger.Error("server.handles.MoneySentHandle: server.handleSuccessfulTransaction failure.", zap.Error(err))
		handleSimpleResponse(w, http.StatusInternalServerError, err.Error())
		return
	}

	s.logger.Info("server.handles.MoneySentHandle: trade completed successfully.")
	handleSimpleResponse(w, http.StatusOK, "trade completed successfully")
	return
}
