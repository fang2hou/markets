package exchange

import (
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

	"markets/pkg/database"
	"markets/pkg/wsclt"
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

type gateioBalanceRestApiResult struct {
	Currency  string `json:"currency"`
	Available string `json:"available"`
	Locked    string `json:"locked"`
}

type gateioBalanceWebSocketApiResult struct {
	Result []struct {
		Currency  string `json:"currency"`
		Available string `json:"available"`
		Total     string `json:"total"`
	} `json:"result"`
}

type gateioOrderBookRestApiResult struct {
	Id   int64      `json:"id"`
	Asks [][]string `json:"asks"`
	Bids [][]string `json:"bids"`
}

type gateioOrderBookWebSocketApiResult struct {
	Result struct {
		GateioCurrency string     `json:"s"`
		FirstUpdate    int64      `json:"U"`
		LastUpdate     int64      `json:"u"`
		Asks           [][]string `json:"a"`
		Bids           [][]string `json:"b"`
	} `json:"result"`
}

type gateioOrderResult struct {
	Data []struct {
		Id               string `json:"id"`
		CreateTime       string `json:"create_time"`
		UpdateTime       string `json:"update_time"`
		Price            string `json:"price"`
		Amount           string `json:"amount"`
		Side             string `json:"side"`
		Type             string `json:"type"`
		Left             string `json:"left"`
		FilledTotalPrice string `json:"filled_total"`
		Fee              string `json:"fee"`
		FeeCurrency      string `json:"fee_currency"`
		Event            string `json:"event"`
		GateioCurrency   string `json:"currency_pair"`
	} `json:"result"`
}

type gateioCacheOrderBook struct {
	Id   int64
	Data *database.OrderBook
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

	orderBookCache map[string]*gateioCacheOrderBook
}

func (e *Gateio) updateFee() error {
	restApiOption := &RestApiOption{
		method: "GET",
		path:   "/wallet/fee",
	}

	if data, err := e.RestApi(restApiOption); err != nil {
		return err
	} else {
		var result gateioFeeResult
		if err := json.Unmarshal(data, &result); err != nil {
			return err
		} else {
			fee := &database.Fee{}
			fee.Maker, _ = strconv.ParseFloat(result.MakerFeeRate, 64)
			fee.Taker, _ = strconv.ParseFloat(result.TakerFeeRate, 64)

			for _, currency := range e.currencies {
				if err := e.database.SetFee(e.name, currency, fee); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func (e *Gateio) initializeOrderBook(currency string) error {
	gateioCurrency := e.convertToGateioCurrencyString(currency)

	restApiOption := &RestApiOption{
		method: "GET",
		path:   "/spot/order_book",
		params: map[string]string{
			"currency_pair": gateioCurrency,
			"limit":         "100",
			"with_id":       "true",
		},
	}

	if data, err := e.RestApi(restApiOption); err != nil {
		return err
	} else {
		e.orderBookCache[currency] = &gateioCacheOrderBook{
			Id:   0,
			Data: &database.OrderBook{},
		}

		var result gateioOrderBookRestApiResult
		if err := json.Unmarshal(data, &result); err != nil {
			return err
		} else {
			e.orderBookCache[currency].Id = result.Id

			updateOrderBook(true, e.orderBookCache[currency].Data, result.Asks, result.Bids)

			if err := e.database.SetOrderBook(e.name, currency, e.orderBookCache[currency].Data); err != nil {
				return err
			}
		}
	}

	return nil
}

func (e *Gateio) updateOrderBook(message []byte) error {
	var result gateioOrderBookWebSocketApiResult
	err := json.Unmarshal(message, &result)
	if err != nil {
		fmt.Println(err)
		return err
	}

	currency := e.convertToGeneralCurrencyString(result.Result.GateioCurrency)

	if orderBook, ok := e.orderBookCache[currency]; ok {
		if orderBook.Id+1 >= result.Result.FirstUpdate && orderBook.Id+1 <= result.Result.LastUpdate {
			updateOrderBook(false, orderBook.Data, result.Result.Asks, result.Result.Bids)
			orderBook.Id = result.Result.LastUpdate
			if err := e.database.SetOrderBook(e.name, currency, orderBook.Data); err != nil {
				return err
			}
		} else if orderBook.Id+1 > result.Result.LastUpdate {
			return nil
		} else if orderBook.Id+1 < result.Result.FirstUpdate {
			err := e.initializeOrderBook(currency)
			if err != nil {
				return err
			}
		}
	} else {
		return errors.New("gateio: order book not found" + currency)
	}

	return nil
}

func (e *Gateio) initializeBalance() error {
	if e.restClient == nil {
		panic(errors.New("the rest api client is not ready"))
	}

	if data, err := e.RestApi(&RestApiOption{
		method: "GET",
		path:   "/spot/accounts",
	}); err != nil {
		panic(err)
	} else {
		var result []gateioBalanceRestApiResult
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

func (e *Gateio) updateBalance(message []byte) error {
	var result gateioBalanceWebSocketApiResult
	if err := json.Unmarshal(message, &result); err != nil {
		return err
	}

	for _, data := range result.Result {
		balance := &database.Balance{}
		balance.Total, _ = strconv.ParseFloat(data.Total, 64)
		balance.Free, _ = strconv.ParseFloat(data.Available, 64)
		balance.Used = balance.Total - balance.Free

		if err := e.database.SetBalance(e.name, data.Currency, balance); err != nil {
			return err
		}
	}

	return nil
}

func (e *Gateio) updateOrder(message []byte) error {
	var result gateioOrderResult
	if err := json.Unmarshal(message, &result); err != nil {
		return err
	}

	for _, o := range result.Data {
		currency := e.convertToGeneralCurrencyString(o.GateioCurrency)
		order := &database.Order{
			Id:           o.Id,
			Type:         o.Type,
			Side:         o.Side,
			CreateTime:   o.CreateTime,
			UpdateTime:   o.UpdateTime,
			Price:        0,
			FilledPrice:  0,
			Amount:       0,
			FilledAmount: 0,
			LeftAmount:   0,
			Status:       "",
			Fee:          0,
			FeeCurrency:  "",
		}

		order.Price, _ = strconv.ParseFloat(o.Price, 64)
		order.Amount, _ = strconv.ParseFloat(o.Amount, 64)
		order.LeftAmount, _ = strconv.ParseFloat(o.Left, 64)
		order.FilledAmount = order.Amount - order.LeftAmount

		switch o.Event {
		case "put":
			order.Status = "normal"
		case "update":
			order.Status = "normal"
		case "finish":
			order.FeeCurrency = o.FeeCurrency
			order.Fee, _ = strconv.ParseFloat(o.Fee, 64)

			if o.Left == "0" {
				order.Status = "finished"
				filledTotalPrice, _ := strconv.ParseFloat(o.FilledTotalPrice, 64)
				order.FilledPrice = filledTotalPrice / order.FilledAmount
			} else if o.Left == o.Amount {
				order.Status = "canceled"
			} else {
				order.Status = "partial canceled"
				filledTotalPrice, _ := strconv.ParseFloat(o.FilledTotalPrice, 64)
				order.FilledPrice = filledTotalPrice / order.FilledAmount
			}
		}

		if err := e.database.SetOrder(e.name, currency, o.Id, order); err != nil {
			return err
		}
	}

	return nil
}

func (e *Gateio) handleMessage(message []byte) {
	var data map[string]interface{}
	if err := json.Unmarshal(message, &data); err != nil {
		return
	}

	if channel, ok := data["channel"]; ok {
		switch channel {
		case "spot.order_book_update":
			if event, ok := data["event"]; ok {
				switch event {
				case "subscribe":
					fmt.Println("Subscribe to order book")
				case "update":
					if err := e.updateOrderBook(message); err != nil {
						panic(err)
					}
				}
			}
		case "spot.orders":
			if event, ok := data["event"]; ok {
				switch event {
				case "subscribe":
					fmt.Println("Subscribe to order")
				case "update":
					if err := e.updateOrder(message); err != nil {
						panic(err)
					}
				}
			}
		case "spot.balances":
			if event, ok := data["event"]; ok {
				switch event {
				case "subscribe":
					fmt.Println("Subscribe to balance")
				case "update":
					if err := e.updateBalance(message); err != nil {
						panic(err)
					}
				}
			}
		}
	}
}

func (e *Gateio) convertToGeneralCurrencyString(gateioCurrencyString string) string {
	return strings.Replace(gateioCurrencyString, "_", "/", -1)
}

func (e *Gateio) convertToGateioCurrencyString(generalCurrencyString string) string {
	return strings.Replace(generalCurrencyString, "/", "_", -1)
}

func (e *Gateio) SendMessageRawBytes(dataBytes []byte) error {
	if err := e.wsClient.SendMessage(dataBytes); err != nil {
		return err
	} else {
		return nil
	}
}

func (e *Gateio) SendMessageJSON(data map[string]interface{}) error {
	var channel, event string

	if value, ok := data["channel"]; ok {
		channel = value.(string)
	}

	if value, ok := data["event"]; ok {
		event = value.(string)
	}

	timeStamp := time.Now().Unix()
	data["time"] = timeStamp

	hash := hmac.New(sha512.New, []byte(e.authData.ApiSecret))
	hash.Write([]byte(fmt.Sprintf("channel=%s&event=%s&time=%d", channel, event, timeStamp)))
	sign := hex.EncodeToString(hash.Sum(nil))

	data["auth"] = map[string]interface{}{
		"method": "api_key",
		"KEY":    e.authData.ApiKey,
		"SIGN":   sign,
	}

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
			"channel": "spot.order_book_update",
			"event":   "subscribe",
			"payload": []string{gateioCurrency, "100ms"},
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

	e.restClient = &http.Client{}

	go e.waitForDisconnecting()

	gateioWebsocketPublicApiURL := url.URL{
		Scheme: GateioWebsocketApiProtocol,
		Host:   GateioWebsocketApiHost,
		Path:   GateioWebsocketApiPath,
	}

	e.wsClient = wsclt.NewClient(&wsclt.Options{
		SkipVerify:     false,
		PingInterval:   e.aliveSignalInterval,
		MessageHandler: e.handleMessage,
	})

	if err := e.wsClient.Connect(gateioWebsocketPublicApiURL.String()); err != nil {
		return err
	}

	e.subscribe()

	if err := e.updateFee(); err != nil {
		return err
	}

	if err := e.initializeBalance(); err != nil {
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

	gateio.orderBookCache = make(map[string]*gateioCacheOrderBook)

	for _, currency := range currencies {
		gateio.orderBookCache[currency] = &gateioCacheOrderBook{
			Id:   0,
			Data: &database.OrderBook{},
		}
	}

	return gateio
}
