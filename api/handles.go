package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

// bisq api
// https://mrosseel.github.io/bisq-api-examples/

var (
	BisqAPIURL = "http://localhost:8080"

	PaymentAccountsURL = "/api/v1/payment-accounts"
	OfferURL           = "/api/v1/offers"
	TakeOfferURL       = "/api/v1/offers/%s/take"
	PaymentStartedURL  = "/api/v1/trades/%s/payment-started"
	PaymentReceivedURL = "/api/v1/trades/%s/payment-received"

	EthplorerAPI    = "https://api.ethplorer.io"
	EthplorerAPIKey = "freekey"
	GetTxURL      = "/getTxInfo/%s"
)

func InitClient() *http.Client {
	client := http.Client{

	}
	return &client
}

type PaymentAccount struct {
	Name                  string   `json:"accountName"`
	TradeCurrencies       []string `json:"tradeCurrencies"`
	PaymentMethod         string   `json:"paymentMethod"`
	ID                    string   `json:"id"`
	Details               string   `json:"paymentDetails"`
	SelectedTradeCurrency string   `json:"selectedTradeCurrency"`
}

func RegisterPaymentAccounts(logger *zap.Logger, client *http.Client, account *PaymentAccount) (*PaymentAccount, error) {
	logger.Info("api.handles.RegisterPaymentAccounts: received new request.")
	apiURL := BisqAPIURL + PaymentAccountsURL

	reqBody, err := json.Marshal(*account)
	if err != nil {
		logger.Error("api.handles.RegisterPaymentAccounts: json marshal failure.", zap.Error(err))
		return nil, err
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(reqBody))
	if err != nil {
		logger.Error("api.handles.RegisterPaymentAccounts: creating request failure.", zap.Error(err))
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	logger.Info("api.handles.RegisterPaymentAccounts: sending request to bisq API.")
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("api.handles.RegisterPaymentAccounts: sending request failure.", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	var p PaymentAccount
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error(
			"api.handles.RegisterPaymentAccounts: response failure",
			zap.Int("status", resp.StatusCode),
			zap.String("body", string(body)),
		)
		return nil, errors.New("response failure, status = " + strconv.Itoa(resp.StatusCode))
	}

	logger.Info("api.handles.RegisterPaymentAccounts: received request successfully.")

	err = json.Unmarshal(body, &p)
	if err != nil {
		logger.Error("api.handles.RegisterPaymentAccounts: json unmarshal failure.", zap.Error(err))
		return nil, err
	}

	return &p, nil
}

type OfferToCreate struct {
	FundUsingBisqWallet       bool   `json:"fundUsingBisqWallet"`
	OfferID                   string `json:"offerId"`
	AccountID                 string `json:"accountId"`
	Direction                 string `json:"direction"`
	PriceType                 string `json:"priceType"`
	MarketPair                string `json:"marketPair"`
	PercentageFromMarketPrice int64  `json:"percentageFromMarketPrice"`
	FixedPrice                int64  `json:"fixedPrice"`
	Amount                    int64  `json:"amount"`
	MinAmount                 int64  `json:"minAmount"`
	BuyerSecurityDeposit      int64  `json:"buyerSecurityDeposit"`
}

type OfferDetail struct {
	Date                       time.Time `json:"date"`
	MinAmount                  int64     `json:"minAmount"`
	SellerSecurityDeposit      int64     `json:"sellerSecurityDeposit"`
	IsPrivateOffer             bool      `json:"isPrivateOffer"`
	IsCurrencyForMakerFeeBtc   bool      `json:"isCurrencyForMakerFeeBtc"`
	MakerPaymentAccountID      string    `json:"makerPaymentAccountId"`
	HashOfChallenge            string    `json:"hashOfChallenge"`
	BuyerSecurityDeposit       int64     `json:"buyerSecurityDeposit"`
	OfferFeePaymentTxID        string    `json:"offerFeePaymentTxId"`
	UseAutoClose               bool      `json:"useAutoClose"`
	BlockHeightAtOfferCreation int64     `json:"blockHeightAtOfferCreation"`
	UseMarketBasedPrice        bool      `json:"useMarketBasedPrice"`
	CounterCurrencyCode        string    `json:"counterCurrencyCode"`
	MakerFee                   int64     `json:"makerFee"`
	CountryCode                string    `json:"countryCode"`
	PaymentMethodID            string    `json:"paymentMethodId"`
	Price                      int64     `json:"price"`
	ProtocolVersion            int32     `json:"protocolVersion"`
	ID                         string    `json:"id"`
	MaxTradePeriod             int64     `json:"maxTradePeriod"`
	State                      string    `json:"state"`
	UseReOpenAfterAutoClose    bool      `json:"useReOpenAfterAutoClose"`
	VersionNr                  string    `json:"versionNr"`
	UpperClosePrice            int64     `json:"upperClosePrice"`
	Direction                  string    `json:"direction"`
	OwnerNodeAddress           string    `json:"ownerNodeAddress"`
	AcceptedBankIds            []string  `json:"acceptedBankIds"`
	MaxTradeLimit              int64     `json:"maxTradeLimit"`
	Amount                     int64     `json:"amount"`
	AcceptedCountryCodes       []string  `json:"acceptedCountryCodes"`
	MarketPriceMargin          float64   `json:"marketPriceMargin"`
	BankID                     string    `json:"bankId"`
	BaseCurrencyCode           string    `json:"baseCurrencyCode"`
	ArbitratorNodeAddresses    []string  `json:"arbitratorNodeAddresses"`
	TxFee                      int64     `json:"txFee"`
	CurrencyCode               string    `json:"currencyCode"`
	LowerClosePrice            int64     `json:"lowerClosePrice"`
}

func PublishOffer(logger *zap.Logger, client *http.Client, offer *OfferToCreate) (*OfferDetail, error) {
	logger.Info("api.handles.PublishOffer: received new request.")
	apiURL := BisqAPIURL + OfferURL

	reqBody, err := json.Marshal(*offer)
	if err != nil {
		logger.Error("api.handles.PublishOffer: json marshal failure.", zap.Error(err))
		return nil, err
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(reqBody))
	if err != nil {
		logger.Error("api.handles.PublishOffer: creating request failure.", zap.Error(err))
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	logger.Info("api.handles.PublishOffer: sending request to bisq API.")
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("api.handles.PublishOffer: sending request failure.", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	var d OfferDetail
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error(
			"api.handles.PublishOffer: response failure",
			zap.Int("status", resp.StatusCode),
			zap.String("body", string(body)),
		)
		return nil, errors.New("response failure, status = " + strconv.Itoa(resp.StatusCode))
	}

	logger.Info("api.handles.PublishOffer: received request successfully.")

	err = json.Unmarshal(body, &d)
	if err != nil {
		logger.Error("api.handles.PublishOffer: json unmarshal failure.", zap.Error(err))
		return nil, err
	}

	return &d, nil
}

type OfferToTake struct {
	PaymentAccountID string `json:"paymentAccountId"`
	Amount           int64  `json:"amount"`
}

type TradeDetails struct {
	BuyerPaymentAccount      PaymentAccount `json:"buyerPaymentAccount"`
	SellerPaymentAccount     PaymentAccount `json:"sellerPaymentAccount"`
	ID                       string         `json:"id"`
	Offer                    OfferDetail    `json:"offer"`
	IsCurrencyForTakerFeeBtc bool           `json:"isCurrencyForTakerFeeBtc"`
	TxFee                    int64          `json:"txFee"`
	TakerFee                 int64          `json:"takerFee"`
	TakeOfferDate            int64          `json:"takeOfferDate"`
	TakerFeeTxID             string         `json:"takerFeeTxId"`
	DepositTxID              string         `json:"depositTxId"`
	PayoutTxID               string         `json:"payoutTxId"`
	TradeAmount              int64          `json:"tradeAmount"`
	TradePrice               int64          `json:"tradePrice"`
	State                    string         `json:"state"`
	DisputeState             string         `json:"disputeState"`
	TradePeriodState         string         `json:"tradePeriodState"`
	ArbitratorBtcPubKey      []string       `json:"arbitratorBtcPubKey"`
	ContractHash             []string       `json:"contractHash"`
	MediatorNodeAddress      string         `json:"mediatorNodeAddress"`
	TakerContractSignature   string         `json:"takerContractSignature"`
	MakerContractSignature   string         `json:"makerContractSignature"`
	ArbitratorNodeAddress    string         `json:"arbitratorNodeAddress"`
	TradingPeerNodeAddress   string         `json:"tradingPeerNodeAddress"`
	TakerPaymentAccountID    string         `json:"takerPaymentAccountId"`
	ErrorMessage             string         `json:"errorMessage"`
	CounterCurrencyTxID      string         `json:"counterCurrencyTxId"`
}

func TakeOffer(logger *zap.Logger, client *http.Client, offer *OfferToTake) (*TradeDetails, error) {
	logger.Info("api.handles.TakeOffer: received new request.")
	apiURL := BisqAPIURL + fmt.Sprintf(TakeOfferURL, offer.PaymentAccountID)

	reqBody, err := json.Marshal(*offer)
	if err != nil {
		logger.Error("api.handles.TakeOffer: json marshal failure.", zap.Error(err))
		return nil, err
	}

	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(reqBody))
	if err != nil {
		logger.Error("api.handles.TakeOffer: creating request failure.", zap.Error(err))
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	logger.Info("api.handles.TakeOffer: sending request to bisq API.")
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("api.handles.TakeOffer: sending request failure.", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	var d TradeDetails
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error(
			"api.handles.TakeOffer: response failure",
			zap.Int("status", resp.StatusCode),
			zap.String("body", string(body)),
		)
		return nil, errors.New("response failure, status = " + strconv.Itoa(resp.StatusCode))
	}

	logger.Info("api.handles.TakeOffer: received request successfully.")

	err = json.Unmarshal(body, &d)
	if err != nil {
		logger.Error("api.handles.TakeOffer: json unmarshal failure.", zap.Error(err))
		return nil, err
	}

	return &d, nil
}

func PaymentStarted(logger *zap.Logger, client *http.Client, trade *TradeDetails) error {
	logger.Info("api.handles.PaymentStarted: received new request.")
	apiURL := BisqAPIURL + fmt.Sprintf(PaymentStartedURL, trade.ID)

	req, err := http.NewRequest("POST", apiURL, nil)
	if err != nil {
		logger.Error("api.handles.PaymentStarted: creating request failure.", zap.Error(err))
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	logger.Info("api.handles.PaymentStarted: sending request to bisq API.")
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("api.handles.PaymentStarted: sending request failure.", zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error(
			"api.handles.PaymentStarted: response failure",
			zap.Int("status", resp.StatusCode),
			zap.String("body", string(body)),
		)
		return errors.New("response failure, status = " + strconv.Itoa(resp.StatusCode))
	}

	logger.Info("api.handles.PaymentStarted: received request successfully.")

	return nil
}

func PaymentReceived(logger *zap.Logger, client *http.Client, trade *TradeDetails) error {
	logger.Info("api.handles.PaymentReceived: received new request.")
	apiURL := BisqAPIURL + fmt.Sprintf(PaymentReceivedURL, trade.ID)

	req, err := http.NewRequest("POST", apiURL, nil)
	if err != nil {
		logger.Error("api.handles.PaymentReceived: creating request failure.", zap.Error(err))
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	logger.Info("api.handles.PaymentReceived: sending request to bisq API.")
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("api.handles.PaymentReceived: sending request failure.", zap.Error(err))
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error(
			"api.handles.PaymentReceived: response failure",
			zap.Int("status", resp.StatusCode),
			zap.String("body", string(body)),
		)
		return errors.New("response failure, status = " + strconv.Itoa(resp.StatusCode))
	}

	logger.Info("api.handles.PaymentReceived: received request successfully.")

	return nil
}

type TransactionLogs struct {
	Address string `json:"address"`
	Topics  string `json:"topics"`
	Data    string `json:"data"`
}

type TransactionOperations struct {
	Timestamp       int64   `json:"timestamp"`
	TransactionHash string  `json:"transactionHash"`
	TokenInfo       string  `json:"tokenInfo"`
	Type            string  `json:"type"`
	Address         string  `json:"address"`
	From            string  `json:"from"`
	To              string  `json:"to"`
	Value           float64 `json:"value"`
}

type TransactionInfo struct {
	Hash          string                  `json:"hash"`
	Timestamp     int64                   `json:"timestamp"`
	BlockNumber   int64                   `json:"blockNumber"`
	Confirmations int                     `json:"confirmations"`
	Success       bool                    `json:"success"`
	From          string                  `json:"from"`
	To            string                  `json:"to"`
	Value         float64                 `json:"value"`
	Input         string                  `json:"input"`
	GasLimit      int64                   `json:"gasLimit"`
	GasUsed       int64                   `json:"gasUsed"`
	Logs          []TransactionLogs       `json:"logs"`
	Operations    []TransactionOperations `json:"operations"`
}

func GetTxInfo(logger *zap.Logger, client *http.Client, transactionID string) (*TransactionInfo, error) {
	logger.Info("api.handles.GetTxInfo: received new request.")
	apiURL := EthplorerAPI + fmt.Sprintf(GetTxURL, transactionID)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		logger.Error("api.handles.GetTxInfo: creating request failure.", zap.Error(err))
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	logger.Info("api.handles.GetTxInfo: sending request to Ethplorer API.")
	resp, err := client.Do(req)
	if err != nil {
		logger.Error("api.handles.GetTxInfo: sending request failure.", zap.Error(err))
		return nil, err
	}
	defer resp.Body.Close()

	var t TransactionInfo
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		logger.Error(
			"api.handles.GetTxInfo: response failure",
			zap.Int("status", resp.StatusCode),
			zap.String("body", string(body)),
		)
		return nil, errors.New("response failure, status = " + strconv.Itoa(resp.StatusCode))
	}

	logger.Info("api.handles.GetTxInfo: received request successfully.")

	err = json.Unmarshal(body, &t)
	if err != nil {
		logger.Error("api.handles.GetTxInfo: json unmarshal failure.", zap.Error(err))
		return nil, err
	}

	return &t, nil
}
