package exchange

import (
	"Markets/pkg/database"
	"Markets/pkg/wsclt"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"time"
)

const (
	CacheTimeout               = 1000
	OkxWebsocketProtocol       = "wss"
	OkxWebsocketApiHost        = "ws.okx.com:8443"
	OkxWebsocketPublicApiPath  = "ws/v5/public"
	OkxWebsocketPrivateApiPath = "ws/v5/private"
)

type Okx struct {
	Exchange
	wsClients struct {
		Public  *wsclt.Client
		Private *wsclt.Client
	}

	running         bool
	publicMessages  chan []byte
	privateMessages chan []byte
	stopErrors      chan error

	orderCache                map[string]map[string]database.Order // Used to cache order data for fetching order IDs
	numInitializedConnections chan int                             // Used to keep track of how many connections have been initialized
	authData                  struct {
		ApiKey     string
		ApiSecret  string
		Passphrase string
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
}

func (e *Okx) SendPublicMessageRawBytes(dataBytes []byte) error {
	if err := e.wsClients.Public.SendMessage(dataBytes); err != nil {
		return err
	}

	return nil
}

func (e *Okx) SendPublicMessageJSON(data *map[string]interface{}) error {
	if dataBytes, err := json.Marshal(data); err != nil {
		return err
	} else {
		return e.SendPublicMessageRawBytes(dataBytes)
	}
}

func (e *Okx) Start() error {
	if e.running {
		return errors.New("exchange is already running")
	} else {
		e.running = true
	}

	go e.waitForDisconnecting()

	okxWebsocketPublicApiURL := url.URL{
		Scheme: OkxWebsocketProtocol,
		Host:   OkxWebsocketApiHost,
		Path:   OkxWebsocketPublicApiPath,
	}

	e.wsClients.Public = wsclt.NewClient(&wsclt.Options{
		SkipVerify:     false,
		PingInterval:   25 * time.Second,
		MessageHandler: e.handlePublicMessage,
	})

	if err := e.wsClients.Public.Connect(okxWebsocketPublicApiURL.String()); err != nil {
		return err
	}

	okxWebsocketPrivateApiURL := url.URL{
		Scheme: OkxWebsocketProtocol,
		Host:   OkxWebsocketApiHost,
		Path:   OkxWebsocketPrivateApiPath,
	}

	e.wsClients.Private = wsclt.NewClient(&wsclt.Options{
		SkipVerify:     false,
		PingInterval:   25 * time.Second,
		MessageHandler: e.handlePublicMessage,
	})

	if err := e.wsClients.Private.Connect(okxWebsocketPrivateApiURL.String()); err != nil {
		return err
	}

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

func NewOkx(config map[string]string, interactor *database.Interactor) *Okx {
	okx := &Okx{
		Exchange: Exchange{
			name:     "OKX",
			database: interactor,
			running:  false,
		},

		running:         false,
		publicMessages:  make(chan []byte),
		privateMessages: make(chan []byte),

		numInitializedConnections: make(chan int),
		orderCache:                make(map[string]map[string]database.Order),
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
