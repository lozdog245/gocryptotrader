package exmo

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/thrasher-corp/gocryptotrader/common"
	"github.com/thrasher-corp/gocryptotrader/config"
	"github.com/thrasher-corp/gocryptotrader/currency"
	exchange "github.com/thrasher-corp/gocryptotrader/exchanges"
	"github.com/thrasher-corp/gocryptotrader/exchanges/request"
	"github.com/thrasher-corp/gocryptotrader/exchanges/ticker"
	"github.com/thrasher-corp/gocryptotrader/exchanges/websocket/wshandler"
	log "github.com/thrasher-corp/gocryptotrader/logger"
)

const (
	exmoAPIURL     = "https://api.exmo.com"
	exmoAPIVersion = "1"

	exmoTrades          = "trades"
	exmoOrderbook       = "order_book"
	exmoTicker          = "ticker"
	exmoPairSettings    = "pair_settings"
	exmoCurrency        = "currency"
	exmoUserInfo        = "user_info"
	exmoOrderCreate     = "order_create"
	exmoOrderCancel     = "order_cancel"
	exmoOpenOrders      = "user_open_orders"
	exmoUserTrades      = "user_trades"
	exmoCancelledOrders = "user_cancelled_orders"
	exmoOrderTrades     = "order_trades"
	exmoRequiredAmount  = "required_amount"
	exmoDepositAddress  = "deposit_address"
	exmoWithdrawCrypt   = "withdraw_crypt"
	exmoGetWithdrawTXID = "withdraw_get_txid"
	exmoExcodeCreate    = "excode_create"
	exmoExcodeLoad      = "excode_load"
	exmoWalletHistory   = "wallet_history"

	// Rate limit: 180 per/minute
	exmoAuthRate   = 180
	exmoUnauthRate = 180
)

// EXMO exchange struct
type EXMO struct {
	exchange.Base
}

// SetDefaults sets the basic defaults for exmo
func (e *EXMO) SetDefaults() {
	e.Name = "EXMO"
	e.Enabled = false
	e.Verbose = false
	e.RESTPollingDelay = 10
	e.APIWithdrawPermissions = exchange.AutoWithdrawCryptoWithSetup | exchange.NoFiatWithdrawals
	e.RequestCurrencyPairFormat.Delimiter = "_"
	e.RequestCurrencyPairFormat.Uppercase = true
	e.RequestCurrencyPairFormat.Separator = ","
	e.ConfigCurrencyPairFormat.Delimiter = "_"
	e.ConfigCurrencyPairFormat.Uppercase = true
	e.AssetTypes = []string{ticker.Spot}
	e.SupportsAutoPairUpdating = true
	e.SupportsRESTTickerBatching = true
	e.Requester = request.New(e.Name,
		request.NewRateLimit(time.Minute, exmoAuthRate),
		request.NewRateLimit(time.Minute, exmoUnauthRate),
		common.NewHTTPClientWithTimeout(exchange.DefaultHTTPTimeout))
	e.APIUrlDefault = exmoAPIURL
	e.APIUrl = e.APIUrlDefault
	e.Websocket = wshandler.New()
}

// Setup takes in the supplied exchange configuration details and sets params
func (e *EXMO) Setup(exch *config.ExchangeConfig) {
	if !exch.Enabled {
		e.SetEnabled(false)
	} else {
		e.Enabled = true
		e.HTTPDebugging = exch.HTTPDebugging
		e.AuthenticatedAPISupport = exch.AuthenticatedAPISupport
		e.SetAPIKeys(exch.APIKey, exch.APISecret, "", false)
		e.SetHTTPClientTimeout(exch.HTTPTimeout)
		e.SetHTTPClientUserAgent(exch.HTTPUserAgent)
		e.RESTPollingDelay = exch.RESTPollingDelay
		e.Verbose = exch.Verbose
		e.BaseCurrencies = exch.BaseCurrencies
		e.AvailablePairs = exch.AvailablePairs
		e.EnabledPairs = exch.EnabledPairs
		err := e.SetCurrencyPairFormat()
		if err != nil {
			log.Fatal(err)
		}
		err = e.SetAssetTypes()
		if err != nil {
			log.Fatal(err)
		}
		err = e.SetAutoPairDefaults()
		if err != nil {
			log.Fatal(err)
		}
		err = e.SetAPIURL(exch)
		if err != nil {
			log.Fatal(err)
		}
		err = e.SetClientProxyAddress(exch.ProxyAddress)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// GetTrades returns the trades for a symbol or symbols
func (e *EXMO) GetTrades(symbol string) (map[string][]Trades, error) {
	v := url.Values{}
	v.Set("pair", symbol)
	result := make(map[string][]Trades)
	urlPath := fmt.Sprintf("%s/v%s/%s", e.APIUrl, exmoAPIVersion, exmoTrades)

	return result, e.SendHTTPRequest(common.EncodeURLValues(urlPath, v), &result)
}

// GetOrderbook returns the orderbook for a symbol or symbols
func (e *EXMO) GetOrderbook(symbol string) (map[string]Orderbook, error) {
	v := url.Values{}
	v.Set("pair", symbol)
	result := make(map[string]Orderbook)
	urlPath := fmt.Sprintf("%s/v%s/%s", e.APIUrl, exmoAPIVersion, exmoOrderbook)

	return result, e.SendHTTPRequest(common.EncodeURLValues(urlPath, v), &result)
}

// GetTicker returns the ticker for a symbol or symbols
func (e *EXMO) GetTicker(symbol string) (map[string]Ticker, error) {
	v := url.Values{}
	v.Set("pair", symbol)
	result := make(map[string]Ticker)
	urlPath := fmt.Sprintf("%s/v%s/%s", e.APIUrl, exmoAPIVersion, exmoTicker)

	return result, e.SendHTTPRequest(common.EncodeURLValues(urlPath, v), &result)
}

// GetPairSettings returns the pair settings for a symbol or symbols
func (e *EXMO) GetPairSettings() (map[string]PairSettings, error) {
	result := make(map[string]PairSettings)
	urlPath := fmt.Sprintf("%s/v%s/%s", e.APIUrl, exmoAPIVersion, exmoPairSettings)

	return result, e.SendHTTPRequest(urlPath, &result)
}

// GetCurrency returns a list of currencies
func (e *EXMO) GetCurrency() ([]string, error) {
	var result []string
	urlPath := fmt.Sprintf("%s/v%s/%s", e.APIUrl, exmoAPIVersion, exmoCurrency)
	return result, e.SendHTTPRequest(urlPath, &result)
}

// GetUserInfo returns the user info
func (e *EXMO) GetUserInfo() (UserInfo, error) {
	var result UserInfo
	err := e.SendAuthenticatedHTTPRequest(http.MethodPost, exmoUserInfo, url.Values{}, &result)
	return result, err
}

// CreateOrder creates an order
// Params: pair, quantity, price and type
// Type can be buy, sell, market_buy, market_sell, market_buy_total and market_sell_total
func (e *EXMO) CreateOrder(pair, orderType string, price, amount float64) (int64, error) {
	type response struct {
		OrderID int64  `json:"order_id"`
		Result  bool   `json:"result"`
		Error   string `json:"error"`
	}

	v := url.Values{}
	v.Set("pair", pair)
	v.Set("type", orderType)
	v.Set("price", strconv.FormatFloat(price, 'f', -1, 64))
	v.Set("quantity", strconv.FormatFloat(amount, 'f', -1, 64))

	var resp response
	err := e.SendAuthenticatedHTTPRequest(http.MethodPost, exmoOrderCreate, v, &resp)
	if !resp.Result {
		return -1, errors.New(resp.Error)
	}
	return resp.OrderID, err
}

// CancelExistingOrder cancels an order by the orderID
func (e *EXMO) CancelExistingOrder(orderID int64) error {
	v := url.Values{}
	v.Set("order_id", strconv.FormatInt(orderID, 10))
	type response struct {
		Result bool   `json:"result"`
		Error  string `json:"error"`
	}
	var resp response
	err := e.SendAuthenticatedHTTPRequest(http.MethodPost, exmoOrderCancel, v, &resp)
	if !resp.Result {
		return errors.New(resp.Error)
	}
	return err
}

// GetOpenOrders returns the users open orders
func (e *EXMO) GetOpenOrders() (map[string]OpenOrders, error) {
	result := make(map[string]OpenOrders)
	err := e.SendAuthenticatedHTTPRequest(http.MethodPost, exmoOpenOrders, url.Values{}, &result)
	return result, err
}

// GetUserTrades returns the user trades
func (e *EXMO) GetUserTrades(pair, offset, limit string) (map[string][]UserTrades, error) {
	result := make(map[string][]UserTrades)
	v := url.Values{}
	v.Set("pair", pair)

	if offset != "" {
		v.Set("offset", offset)
	}

	if limit != "" {
		v.Set("limit", limit)
	}

	err := e.SendAuthenticatedHTTPRequest(http.MethodPost, exmoUserTrades, v, &result)
	return result, err
}

// GetCancelledOrders returns a list of cancelled orders
func (e *EXMO) GetCancelledOrders(offset, limit string) ([]CancelledOrder, error) {
	var result []CancelledOrder
	v := url.Values{}

	if offset != "" {
		v.Set("offset", offset)
	}

	if limit != "" {
		v.Set("limit", limit)
	}

	err := e.SendAuthenticatedHTTPRequest(http.MethodPost, exmoCancelledOrders, v, &result)
	return result, err
}

// GetOrderTrades returns a history of order trade details for the specific orderID
func (e *EXMO) GetOrderTrades(orderID int64) (OrderTrades, error) {
	var result OrderTrades
	v := url.Values{}
	v.Set("order_id", strconv.FormatInt(orderID, 10))

	err := e.SendAuthenticatedHTTPRequest(http.MethodPost, exmoOrderTrades, v, &result)
	return result, err
}

// GetRequiredAmount calculates the sum of buying a certain amount of currency
// for the particular currency pair
func (e *EXMO) GetRequiredAmount(pair string, amount float64) (RequiredAmount, error) {
	v := url.Values{}
	v.Set("pair", pair)
	v.Set("quantity", strconv.FormatFloat(amount, 'f', -1, 64))
	var result RequiredAmount
	err := e.SendAuthenticatedHTTPRequest(http.MethodPost, exmoRequiredAmount, v, &result)
	return result, err
}

// GetCryptoDepositAddress returns a list of addresses for cryptocurrency deposits
func (e *EXMO) GetCryptoDepositAddress() (map[string]string, error) {
	var result interface{}
	err := e.SendAuthenticatedHTTPRequest(http.MethodPost, exmoDepositAddress, url.Values{}, &result)
	if err != nil {
		return nil, err
	}

	switch r := result.(type) {
	case map[string]interface{}:
		mapString := make(map[string]string)

		for key, value := range r {
			strValue := fmt.Sprintf("%v", value)
			mapString[key] = strValue
		}

		return mapString, nil

	default:
		return nil, errors.New("no addresses found, generate required addresses via site")
	}
}

// WithdrawCryptocurrency withdraws a cryptocurrency from the exchange to the desired address
// NOTE: This API function is available only after request to their tech support team
func (e *EXMO) WithdrawCryptocurrency(currency, address, invoice string, amount float64) (int64, error) {
	type response struct {
		TaskID  int64  `json:"task_id,string"`
		Result  bool   `json:"result"`
		Error   string `json:"error"`
		Success int64  `json:"success"`
	}

	v := url.Values{}
	v.Set("currency", currency)
	v.Set("address", address)

	if common.StringToUpper(currency) == "XRP" {
		v.Set(invoice, invoice)
	}

	v.Set("amount", strconv.FormatFloat(amount, 'f', -1, 64))
	var resp response
	err := e.SendAuthenticatedHTTPRequest(http.MethodPost, exmoWithdrawCrypt, v, &resp)
	if err != nil {
		return -1, err
	}
	if resp.Success == 0 || !resp.Result {
		return -1, errors.New(resp.Error)
	}
	return resp.TaskID, err
}

// GetWithdrawTXID gets the result of a withdrawal request
func (e *EXMO) GetWithdrawTXID(taskID int64) (string, error) {
	type response struct {
		Status bool   `json:"status"`
		TXID   string `json:"txid"`
	}

	v := url.Values{}
	v.Set("task_id", strconv.FormatInt(taskID, 10))

	var result response
	err := e.SendAuthenticatedHTTPRequest(http.MethodPost, exmoGetWithdrawTXID, v, &result)
	return result.TXID, err
}

// ExcodeCreate creates an EXMO coupon
func (e *EXMO) ExcodeCreate(currency string, amount float64) (ExcodeCreate, error) {
	v := url.Values{}
	v.Set("currency", currency)
	v.Set("amount", strconv.FormatFloat(amount, 'f', -1, 64))

	var result ExcodeCreate
	err := e.SendAuthenticatedHTTPRequest(http.MethodPost, exmoExcodeCreate, v, &result)
	return result, err
}

// ExcodeLoad loads an EXMO coupon
func (e *EXMO) ExcodeLoad(excode string) (ExcodeLoad, error) {
	v := url.Values{}
	v.Set("code", excode)

	var result ExcodeLoad
	err := e.SendAuthenticatedHTTPRequest(http.MethodPost, exmoExcodeLoad, v, &result)
	return result, err
}

// GetWalletHistory returns the users deposit/withdrawal history
func (e *EXMO) GetWalletHistory(date int64) (WalletHistory, error) {
	v := url.Values{}
	v.Set("date", strconv.FormatInt(date, 10))

	var result WalletHistory
	err := e.SendAuthenticatedHTTPRequest(http.MethodPost, exmoWalletHistory, v, &result)
	return result, err
}

// SendHTTPRequest sends an unauthenticated HTTP request
func (e *EXMO) SendHTTPRequest(path string, result interface{}) error {
	return e.SendPayload(http.MethodGet,
		path,
		nil,
		nil,
		result,
		false,
		false,
		e.Verbose,
		e.HTTPDebugging,
		e.HTTPRecording)
}

// SendAuthenticatedHTTPRequest sends an authenticated HTTP request
func (e *EXMO) SendAuthenticatedHTTPRequest(method, endpoint string, vals url.Values, result interface{}) error {
	if !e.AuthenticatedAPISupport {
		return fmt.Errorf(exchange.WarningAuthenticatedRequestWithoutCredentialsSet,
			e.Name)
	}

	n := e.Requester.GetNonce(true).String()
	vals.Set("nonce", n)

	payload := vals.Encode()
	hash := common.GetHMAC(common.HashSHA512,
		[]byte(payload),
		[]byte(e.APISecret))

	if e.Verbose {
		log.Debugf("Sending %s request to %s with params %s\n",
			method,
			endpoint,
			payload)
	}

	headers := make(map[string]string)
	headers["Key"] = e.APIKey
	headers["Sign"] = common.HexEncodeToString(hash)
	headers["Content-Type"] = "application/x-www-form-urlencoded"

	path := fmt.Sprintf("%s/v%s/%s", e.APIUrl, exmoAPIVersion, endpoint)

	return e.SendPayload(method,
		path,
		headers,
		strings.NewReader(payload),
		result,
		true,
		true,
		e.Verbose,
		e.HTTPDebugging,
		e.HTTPRecording)
}

// GetFee returns an estimate of fee based on type of transaction
func (e *EXMO) GetFee(feeBuilder *exchange.FeeBuilder) (float64, error) {
	var fee float64
	switch feeBuilder.FeeType {
	case exchange.CryptocurrencyTradeFee:
		fee = calculateTradingFee(feeBuilder.PurchasePrice, feeBuilder.Amount)
	case exchange.CryptocurrencyWithdrawalFee:
		fee = getCryptocurrencyWithdrawalFee(feeBuilder.Pair.Base)
	case exchange.InternationalBankWithdrawalFee:
		fee = getInternationalBankWithdrawalFee(feeBuilder.FiatCurrency,
			feeBuilder.Amount,
			feeBuilder.BankTransactionType)
	case exchange.InternationalBankDepositFee:
		fee = getInternationalBankDepositFee(feeBuilder.FiatCurrency,
			feeBuilder.Amount,
			feeBuilder.BankTransactionType)
	case exchange.OfflineTradeFee:
		fee = calculateTradingFee(feeBuilder.PurchasePrice, feeBuilder.Amount)
	}

	if fee < 0 {
		fee = 0
	}

	return fee, nil
}

func getCryptocurrencyWithdrawalFee(c currency.Code) float64 {
	return WithdrawalFees[c]
}

func calculateTradingFee(price, amount float64) float64 {
	return 0.002 * price * amount
}

func getInternationalBankWithdrawalFee(c currency.Code, amount float64, bankTransactionType exchange.InternationalBankTransactionType) float64 {
	var fee float64

	switch bankTransactionType {
	case exchange.WireTransfer:
		switch c {
		case currency.RUB:
			fee = 3200
		case currency.PLN:
			fee = 125
		case currency.TRY:
			fee = 0
		}
	case exchange.PerfectMoney:
		switch c {
		case currency.USD:
			fee = 0.01 * amount
		case currency.EUR:
			fee = 0.0195 * amount
		}
	case exchange.Neteller:
		switch c {
		case currency.USD:
			fee = 0.0195 * amount
		case currency.EUR:
			fee = 0.0195 * amount
		}
	case exchange.AdvCash:
		switch c {
		case currency.USD:
			fee = 0.0295 * amount
		case currency.EUR:
			fee = 0.03 * amount
		case currency.RUB:
			fee = 0.0195 * amount
		case currency.UAH:
			fee = 0.0495 * amount
		}
	case exchange.Payeer:
		switch c {
		case currency.USD:
			fee = 0.0395 * amount
		case currency.EUR:
			fee = 0.01 * amount
		case currency.RUB:
			fee = 0.0595 * amount
		}
	case exchange.Skrill:
		switch c {
		case currency.USD:
			fee = 0.0145 * amount
		case currency.EUR:
			fee = 0.03 * amount
		case currency.TRY:
			fee = 0
		}
	case exchange.VisaMastercard:
		switch c {
		case currency.USD:
			fee = 0.06 * amount
		case currency.EUR:
			fee = 0.06 * amount
		case currency.PLN:
			fee = 0.06 * amount
		}
	}

	return fee
}

func getInternationalBankDepositFee(c currency.Code, amount float64, bankTransactionType exchange.InternationalBankTransactionType) float64 {
	var fee float64
	switch bankTransactionType {
	case exchange.WireTransfer:
		switch c {
		case currency.RUB:
			fee = 1600
		case currency.PLN:
			fee = 30
		case currency.TRY:
			fee = 0
		}
	case exchange.Neteller:
		switch c {
		case currency.USD:
			fee = (0.035 * amount) + 0.29
		case currency.EUR:
			fee = (0.035 * amount) + 0.25
		}
	case exchange.AdvCash:
		switch c {
		case currency.USD:
			fee = 0.0295 * amount
		case currency.EUR:
			fee = 0.01 * amount
		case currency.RUB:
			fee = 0.0495 * amount
		case currency.UAH:
			fee = 0.01 * amount
		}
	case exchange.Payeer:
		switch c {
		case currency.USD:
			fee = 0.0195 * amount
		case currency.EUR:
			fee = 0.0295 * amount
		case currency.RUB:
			fee = 0.0345 * amount
		}
	case exchange.Skrill:
		switch c {
		case currency.USD:
			fee = (0.0495 * amount) + 0.36
		case currency.EUR:
			fee = (0.0295 * amount) + 0.29
		case currency.PLN:
			fee = (0.035 * amount) + 1.21
		case currency.TRY:
			fee = 0
		}
	}

	return fee
}
