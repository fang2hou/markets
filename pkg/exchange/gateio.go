package exchange

import (
	"Markets/pkg/database"
	"Markets/pkg/wsclt"
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"
)

const (
	GateioWebsocketApiProtocol = "wss"
	GateioWebsocketApiHost     = "api.gateio.ws"
	GateioWebsocketApiPath     = "/ws/v4/"

	GateioRestApiProtocol = "https"
	GateioRestApiHost     = "api.gateio.ws"
	GateioRestApiPath     = "/api/v4"
)

type gateioFeeResult struct {
	TakerFeeRate string `json:"taker_fee"`
	MakerFeeRate string `json:"maker_fee"`
}

type gateioBalanceResult struct {
	Currency  string `json:"currency"`
	Available string `json:"available"`
	Locked    string `json:"locked"`
}

type Gateio struct {
	Exchange

	restClient *http.Client
	wsClient   *wsclt.Client

	messages chan []byte

	authData struct {
		ApiKey    string
		ApiSecret string
	}
}

func (e *Gateio) updateFee() {
	if data, err := e.RestApi(&RestApiOption{
		method: "GET",
		path:   "/wallet/fee",
	}); err != nil {
		panic(err)
	} else {
		var result gateioFeeResult
		if err := json.Unmarshal(data, &result); err != nil {
			panic(err)
		} else {

			makerFee, _ := strconv.ParseFloat(result.MakerFeeRate, 64)
			takerFee, _ := strconv.ParseFloat(result.TakerFeeRate, 64)

			fee := &database.Fee{
				Maker: makerFee,
				Taker: takerFee,
			}

			for _, currency := range e.currencies {
				gateioCurrency := e.convertToGateioCurrencyString(currency)
				err := e.database.SetFee(e.name, gateioCurrency, fee)
				if err != nil {
					return
				}
			}
		}
	}
}

func (e *Gateio) updateBalance() error {
	if e.restClient == nil {
		panic(errors.New("the rest api client is not ready"))
	}

	if data, err := e.RestApi(&RestApiOption{
		method: "GET",
		path:   "/spot/accounts",
	}); err != nil {
		panic(err)
	} else {
		var result []gateioBalanceResult
		if err := json.Unmarshal(data, &result); err != nil {
			return err
		} else {
			for _, b := range result {
				balance := &database.Balance{}

				balance.Free, _ = strconv.ParseFloat(b.Available, 64)
				balance.Used, _ = strconv.ParseFloat(b.Locked, 64)
				balance.Total = balance.Free + balance.Used

				if err := e.database.SetBalance(e.name, b.Currency, balance); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (e *Gateio) handleMessage(message []byte) {
	if !jsonCheck.MatchString(string(message)) {
		return
	}

}

func (e *Gateio) convertToGeneralCurrencyString(okxCurrencyString string) string {
	return strings.Replace(okxCurrencyString, "/", "_", -1)
}

func (e *Gateio) convertToGateioCurrencyString(okxCurrencyString string) string {
	return strings.Replace(okxCurrencyString, "_", "/", -1)
}

func (e *Gateio) SendMessageRawBytes(dataBytes []byte) error {
	if err := e.wsClient.SendMessage(dataBytes); err != nil {
		return err
	} else {
		return nil
	}
}

func (e *Gateio) SendMessageJSON(data map[string]interface{}) error {
	if dataBytes, err := json.Marshal(data); err != nil {
		return err
	} else {
		return e.SendMessageRawBytes(dataBytes)
	}
}

func (e *Gateio) subscribe() {
	// Order Book
	for _, currency := range e.currencies {
		gateioCurrency := e.convertToGateioCurrencyString(currency)
		params := map[string]interface{}{
			"time":    time.Now().Unix(),
			"channel": "spot.order_book",
			"event":   "subscribe",
			"instId":  []string{gateioCurrency, "10", "100ms"},
		}

		if err := e.SendMessageJSON(params); err != nil {
			panic(err)
		}
	}

	// Order
	{
		currencies := make([]string, 0)

		for _, currency := range e.currencies {
			currencies = append(currencies, e.convertToGateioCurrencyString(currency))
		}

		params := map[string]interface{}{
			"time":    time.Now().Unix(),
			"channel": "spot.orders",
			"event":   "subscribe",
			"payload": currencies,
		}

		if err := e.SendMessageJSON(params); err != nil {
			panic(err)
		}
	}

	// balance
	{
		params := map[string]interface{}{
			"time":    time.Now().Unix(),
			"channel": "spot.balances",
			"event":   "subscribe",
		}

		if err := e.SendMessageJSON(params); err != nil {
			panic(err)
		}
	}
}

func (e *Gateio) waitForDisconnecting() {
	// Handle SIGINT and SIGTERM.
	interruptSignal := make(chan os.Signal, 1)
	signal.Notify(interruptSignal, os.Interrupt)
	defer close(interruptSignal)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-interruptSignal:
			_ = e.Stop()
			return
		case <-ticker.C:
			if !e.wsClient.IsReading() || !e.wsClient.IsSending() {
				_ = e.Stop()
				return
			}
		}
	}
}

func (e *Gateio) RestApi(option *RestApiOption) ([]byte, error) {
	method := strings.ToUpper(option.method)
	timeStamp := strconv.FormatInt(time.Now().Unix(), 10)

	hash := sha512.New()
	var content string
	if option.body != nil {
		if contentBytes, err := json.Marshal(option.body); err != nil {
			return nil, err
		} else {
			content = string(contentBytes)
			hash.Write(contentBytes)
		}
	}

	hashedContent := hex.EncodeToString(hash.Sum(nil))

	queryString := ""
	if option.params != nil {
		for key, value := range option.params {
			queryString += key + "=" + value + "&"
		}
		queryString = queryString[:len(queryString)-1]
	}

	hash = hmac.New(sha512.New, []byte(e.authData.ApiSecret))

	hash.Write([]byte(fmt.Sprintf("%s\n%s\n%s\n%s\n%s",
		method,
		GateioRestApiPath+option.path,
		queryString,
		hashedContent,
		timeStamp,
	)))

	sign := hex.EncodeToString(hash.Sum(nil))

	restApiURL := url.URL{
		Scheme: GateioRestApiProtocol,
		Host:   GateioRestApiHost,
		Path:   GateioRestApiPath,
	}

	restApiURLString := restApiURL.String() + option.path
	if queryString != "" {
		restApiURLString += "?" + queryString
	}

	if req, err := http.NewRequest(method, restApiURLString, strings.NewReader(content)); err == nil {
		req.Header.Add("KEY", e.authData.ApiKey)
		req.Header.Add("SIGN", sign)
		req.Header.Add("Timestamp", timeStamp)
		req.Header.Add("Content-Type", "application/json")
		req.Header.Add("Accept", "application/json")

		if resp, err := e.restClient.Do(req); err == nil {
			defer func(Body io.ReadCloser) {
				err := Body.Close()
				if err != nil {
					panic(err)
				}
			}(resp.Body)

			if resp.StatusCode == http.StatusOK ||
				resp.StatusCode == http.StatusCreated ||
				resp.StatusCode == http.StatusAccepted {

				if bodyBytes, err := io.ReadAll(resp.Body); err != nil {
					return nil, err
				} else {
					return bodyBytes, nil
				}
			} else {
				return nil, errors.New(resp.Status)
			}
		} else {
			return nil, err
		}
	} else {
		return nil, err
	}
}

func (e *Gateio) Start() error {
	if e.running {
		return errors.New("exchange is already running")
	} else {
		e.running = true
	}

	go e.waitForDisconnecting()

	okxWebsocketPublicApiURL := url.URL{
		Scheme: GateioWebsocketApiProtocol,
		Host:   GateioWebsocketApiHost,
		Path:   GateioWebsocketApiPath,
	}

	e.wsClient = wsclt.NewClient(&wsclt.Options{
		SkipVerify:     true,
		PingInterval:   e.aliveSignalInterval,
		MessageHandler: e.handleMessage,
	})

	if err := e.wsClient.Connect(okxWebsocketPublicApiURL.String()); err != nil {
		return err
	}

	e.subscribe()

	e.restClient = &http.Client{}
	e.updateFee()
	err := e.updateBalance()
	if err != nil {
		return err
	}

	return nil
}

func (e *Gateio) Stop() error {
	if !e.running {
		return nil
	}

	if err := e.wsClient.Close(); err != nil {
		return err
	}

	e.running = false
	return nil
}

func NewGateio(config map[string]string, currencies []string, interactor *database.Interactor) *Gateio {
	gateio := &Gateio{
		Exchange: Exchange{
			name:                "gateio",
			database:            interactor,
			running:             false,
			aliveSignalInterval: 25 * time.Second,
			currencies:          currencies,
		},

		messages: make(chan []byte, 100),
	}

	if apiKey, ok := config["apiKey"]; ok {
		gateio.authData.ApiKey = apiKey
	} else {
		panic("No API key provided for Gateio")
	}

	if apiSecret, ok := config["secret"]; ok {
		gateio.authData.ApiSecret = apiSecret
	} else {
		panic("No API secret provided for Gateio")
	}

	return gateio
}
