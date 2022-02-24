package exchange

import (
	"Markets/pkg/database"
	"Markets/pkg/wsclt"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
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
	OkxWebsocketApiProtocol          = "wss"
	OkxWebsocketApiHost              = "ws.okx.com:8443"
	OkxWebsocketPublicApiPath        = "/ws/v5/public"
	OkxWebsocketPrivateApiPath       = "/ws/v5/private"
	OkxWebsocketPrivateApiVerifyPath = "/users/self/verify"

	OkxRestApiProtocol        = "https"
	OkxRestApiHost            = "www.okx.com"
	OkxRestApiPath            = "/api/v5"
	OkxRestApiTimeStampFormat = "2006-01-02T15:04:05.999Z"
)

type restApiOption struct {
	method string
	path   string
	body   map[string]interface{}
	params map[string]string
}

type feeResult struct {
	Code int `json:"string"`
	Data []struct {
		Maker string `json:"maker"`
		Taker string `json:"taker"`
	} `json:"data"`
}

type Okx struct {
	Exchange
	wsClients struct {
		Public  *wsclt.Client
		Private *wsclt.Client
	}
	restClient *http.Client

	publicMessages  chan []byte
	privateMessages chan []byte
	loginCode       chan int

	orderCache map[string]map[string]database.Order // Used to cache order data for fetching order IDs
	authData   struct {
		ApiKey     string
		ApiSecret  string
		Passphrase string
	}
}

func (e *Okx) restApi(option *restApiOption) ([]byte, error) {
	method := strings.ToUpper(option.method)
	timeStamp := time.Now().UTC().Format(OkxRestApiTimeStampFormat)

	content := ""
	if option.body != nil {
		if contentBytes, err := json.Marshal(option.body); err != nil {
			return nil, err
		} else {
			content = string(contentBytes)
		}
	}

	queryString := ""
	if option.params != nil {
		queryString += "?"
		for key, value := range option.params {
			queryString += key + "=" + value + "&"
		}
		queryString = queryString[:len(queryString)-1]
	}

	hash := hmac.New(sha256.New, []byte(e.authData.ApiSecret))
	hash.Write([]byte(timeStamp + method + OkxRestApiPath + option.path + queryString + content))
	sign := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	restApiURL := url.URL{
		Scheme: OkxRestApiProtocol,
		Host:   OkxRestApiHost,
		Path:   OkxRestApiPath,
	}

	restApiURLString := restApiURL.String() + option.path + queryString

	if req, err := http.NewRequest(method, restApiURLString, strings.NewReader(content)); err == nil {
		req.Header.Add("OK-ACCESS-KEY", e.authData.ApiKey)
		req.Header.Add("OK-ACCESS-SIGN", sign)
		req.Header.Add("OK-ACCESS-TIMESTAMP", timeStamp)
		req.Header.Add("OK-ACCESS-PASSPHRASE", e.authData.Passphrase)
		req.Header.Add("Content-Type", "application/json")

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

func (e *Okx) updateFee() {
	if e.restClient == nil {
		panic(errors.New("the rest api client is not ready"))
	}

	for _, currency := range e.currencies {
		okxCurrency := e.convertToGeneralCurrencyString(currency)
		if data, err := e.restApi(&restApiOption{
			method: "GET",
			path:   "/account/trade-fee",
			params: map[string]string{
				"instType": "SPOT",
				"instId":   okxCurrency,
			},
		}); err != nil {
			panic(err)
		} else {
			var result feeResult
			if err := json.Unmarshal(data, &result); err != nil {
				panic(err)
			} else {
				if len(result.Data) > 0 {
					var fee database.Fee

					if value, err := strconv.ParseFloat(result.Data[0].Maker, 64); err != nil {
						panic(err)
					} else {
						fee.Maker = value
					}

					if value, err := strconv.ParseFloat(result.Data[0].Taker, 64); err != nil {
						panic(err)
					} else {
						fee.Taker = value
					}

					if err := e.database.SetFee(e.name, okxCurrency, &fee); err != nil {
						panic(err)
					}
				} else {
					panic(errors.New("the length of fee result is 0"))
				}
			}
		}
	}
}

func (e *Okx) waitForDisconnecting() {
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
			if !e.wsClients.Public.IsReading() ||
				!e.wsClients.Public.IsSending() ||
				!e.wsClients.Private.IsReading() ||
				!e.wsClients.Private.IsSending() {
				_ = e.Stop()
				return
			}
		}
	}
}

func (e *Okx) handlePublicMessage(message []byte) {
	var data map[string]interface{}

	if err := json.Unmarshal(message, &data); err != nil {
		fmt.Println("public message handler error:", err)
		return
	}

	fmt.Printf("Received message: %v+\n", data)

	if event, ok := data["event"]; ok {
		switch event {
		case "subscribe":
			if argInterface, ok := data["arg"]; ok {
				arg := argInterface.(map[string]interface{})
				if arg["channel"].(string) == "books50-l2-tbt" {
					fmt.Println("Subscribed to books50-l2-tbt with", arg["instId"].(string))
				}
			}
		}
	}
}

func (e *Okx) handlePrivateMessage(message []byte) {
	var data map[string]interface{}

	if err := json.Unmarshal(message, &data); err != nil {
		fmt.Println("private message handler error:", err)
		return
	}

	fmt.Printf("Received message: %v+\n", data)

	if event, ok := data["event"]; ok {
		switch event {
		case "login":
			if codeString, ok := data["code"]; ok {
				if code, err := strconv.Atoi(codeString.(string)); err == nil {
					select {
					case e.loginCode <- code:
					default:
						fmt.Println("already logged in")
					}
				} else {
					fmt.Println("login code conversion error:", err)
				}
			}
		}
	}
}

func (e *Okx) convertToGeneralCurrencyString(okxCurrencyString string) string {
	return strings.Replace(okxCurrencyString, "/", "-", -1)
}

func (e *Okx) convertToOkxCurrencyString(okxCurrencyString string) string {
	return strings.Replace(okxCurrencyString, "-", "/", -1)
}

func (e *Okx) subscribe() {
	var args []interface{}

	for _, currency := range e.currencies {
		okxCurrency := e.convertToGeneralCurrencyString(currency)
		args = append(args, map[string]interface{}{
			"channel": "books50-l2-tbt",
			"instId":  okxCurrency,
		})
	}

	if err := e.SendPublicMessageJSON(&map[string]interface{}{
		"op":   "subscribe",
		"args": args,
	}); err != nil {
		panic(err)
	}

	args = make([]interface{}, 0)

	args = append(args, map[string]interface{}{
		"channel": "account",
	})

	for _, currency := range e.currencies {
		okxCurrency := e.convertToGeneralCurrencyString(currency)
		args = append(args, map[string]interface{}{
			"channel":  "orders",
			"instType": "SPOT",
			"instId":   okxCurrency,
		})
	}

	if err := e.SendPrivateMessageJSON(&map[string]interface{}{
		"op":   "subscribe",
		"args": args,
	}); err != nil {
		panic(err)
	}
}

func (e *Okx) login() {
	epochTime := fmt.Sprint(time.Now().UTC().Unix())
	hash := hmac.New(sha256.New, []byte(e.authData.ApiSecret))
	hash.Write([]byte(epochTime + "GET" + OkxWebsocketPrivateApiVerifyPath))
	sign := base64.StdEncoding.EncodeToString(hash.Sum(nil))

	if err := e.SendPrivateMessageJSON(&map[string]interface{}{
		"op": "login",
		"args": []map[string]interface{}{
			{
				"apiKey":     e.authData.ApiKey,
				"passphrase": e.authData.Passphrase,
				"timestamp":  epochTime,
				"sign":       sign,
			},
		},
	}); err != nil {
		panic(err)
	}

	if code := <-e.loginCode; code != 0 {
		panic(errors.New("login failed"))
	} else {
		close(e.loginCode)
		fmt.Println("login!")
	}
}

func (e *Okx) sendMessageRawBytes(clt *wsclt.Client, dataBytes []byte) error {
	if err := clt.SendMessage(dataBytes); err != nil {
		return err
	} else {
		return nil
	}
}

func (e *Okx) SendPublicMessageRawBytes(dataBytes []byte) error {
	return e.sendMessageRawBytes(e.wsClients.Public, dataBytes)
}

func (e *Okx) SendPrivateMessageRawBytes(dataBytes []byte) error {
	return e.sendMessageRawBytes(e.wsClients.Private, dataBytes)
}

func (e *Okx) sendMessageJSON(clt *wsclt.Client, data *map[string]interface{}) error {
	if dataBytes, err := json.Marshal(data); err != nil {
		return err
	} else {
		return e.sendMessageRawBytes(clt, dataBytes)
	}
}

func (e *Okx) SendPublicMessageJSON(data *map[string]interface{}) error {
	return e.sendMessageJSON(e.wsClients.Public, data)
}

func (e *Okx) SendPrivateMessageJSON(data *map[string]interface{}) error {
	return e.sendMessageJSON(e.wsClients.Private, data)
}

func (e *Okx) Start() error {
	if e.running {
		return errors.New("exchange is already running")
	} else {
		e.running = true
	}

	go e.waitForDisconnecting()

	okxWebsocketPublicApiURL := url.URL{
		Scheme: OkxWebsocketApiProtocol,
		Host:   OkxWebsocketApiHost,
		Path:   OkxWebsocketPublicApiPath,
	}

	e.wsClients.Public = wsclt.NewClient(&wsclt.Options{
		SkipVerify:     false,
		PingInterval:   e.aliveSignalInterval,
		MessageHandler: e.handlePublicMessage,
	})

	if err := e.wsClients.Public.Connect(okxWebsocketPublicApiURL.String()); err != nil {
		return err
	}

	okxWebsocketPrivateApiURL := url.URL{
		Scheme: OkxWebsocketApiProtocol,
		Host:   OkxWebsocketApiHost,
		Path:   OkxWebsocketPrivateApiPath,
	}

	e.wsClients.Private = wsclt.NewClient(&wsclt.Options{
		SkipVerify:     false,
		PingInterval:   e.aliveSignalInterval,
		MessageHandler: e.handlePrivateMessage,
	})

	if err := e.wsClients.Private.Connect(okxWebsocketPrivateApiURL.String()); err != nil {
		return err
	}

	e.login()
	e.subscribe()

	e.restClient = &http.Client{}

	return nil
}

func (e *Okx) Stop() error {
	if !e.running {
		return nil
	}

	if err := e.wsClients.Public.Close(); err != nil {
		return err
	}

	if err := e.wsClients.Private.Close(); err != nil {
		return err
	}

	e.running = false
	return nil
}

func NewOkx(config map[string]string, currencies []string, interactor *database.Interactor) *Okx {
	okx := &Okx{
		Exchange: Exchange{
			name:                "okx",
			database:            interactor,
			running:             false,
			aliveSignalInterval: 25 * time.Second,
			currencies:          currencies,
		},

		publicMessages:  make(chan []byte),
		privateMessages: make(chan []byte),
		loginCode:       make(chan int),

		orderCache: make(map[string]map[string]database.Order),
	}

	if apiKey, ok := config["apiKey"]; ok {
		okx.authData.ApiKey = apiKey
	} else {
		panic("No API key provided for OKX")
	}

	if apiSecret, ok := config["secret"]; ok {
		okx.authData.ApiSecret = apiSecret
	} else {
		panic("No API secret provided for OKX")
	}

	if passphrase, ok := config["password"]; ok {
		okx.authData.Passphrase = passphrase
	} else {
		panic("No API Passphrase provided for OKX")
	}

	return okx
}
