package wsclt

import (
	"crypto/tls"
	"errors"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Options struct {
	SkipVerify     bool
	PingInterval   time.Duration
	MessageHandler func([]byte)
}

type Client struct {
	ws             *websocket.Conn
	options        *Options
	messageHandler func([]byte)

	isReading             bool
	isSending             bool
	messageWaitForSending chan []byte
	sendMux               sync.Mutex
}

func (clt *Client) readMessage() {
	clt.isReading = true
	defer func() {
		clt.isReading = false
	}()

	for {
		_, message, err := clt.ws.ReadMessage()
		if err != nil {
			close(clt.messageWaitForSending)
			return
		}

		if clt.messageHandler != nil {
			clt.messageHandler(message)
		}
	}
}

func (clt *Client) sendMessage() {
	clt.isSending = true
	defer func() {
		clt.isSending = false
	}()

	pingTicker := time.NewTicker(clt.options.PingInterval)
	defer pingTicker.Stop()

	for {
		select {
		case <-pingTicker.C:
			err := clt.ws.WriteMessage(websocket.TextMessage, []byte("ping"))
			if err != nil {
				return
			}
		case message, more := <-clt.messageWaitForSending:
			if !more {
				return
			}
			clt.sendMux.Lock()
			err := clt.ws.WriteMessage(websocket.TextMessage, message)
			clt.sendMux.Unlock()
			if err != nil {
				return
			}
		}
	}
}

func (clt *Client) IsReading() bool {
	return clt.isReading
}

func (clt *Client) IsSending() bool {
	return clt.isSending
}

func (clt *Client) RegisterMessageHandler(handler func([]byte)) {
	clt.messageHandler = handler
}

func (clt *Client) SendMessage(message []byte) error {
	if clt.isSending {
		clt.messageWaitForSending <- message
		return nil
	}

	return errors.New("client is closed")
}

func (clt *Client) Connect(url string) error {
	if clt.ws != nil {
		return errors.New("already connected")
	}

	dialer := &websocket.Dialer{
		HandshakeTimeout: 45 * time.Second,
	}

	if clt.options.SkipVerify {
		dialer.TLSClientConfig = &tls.Config{RootCAs: nil, InsecureSkipVerify: true}
	}

	ws, _, err := dialer.Dial(url, nil)
	if err != nil {
		return err
	}

	clt.ws = ws

	clt.messageWaitForSending = make(chan []byte)

	go clt.sendMessage()
	go clt.readMessage()

	return nil
}

func (clt *Client) Close() error {
	if clt.ws == nil {
		return nil
	}

	data := websocket.FormatCloseMessage(websocket.CloseNormalClosure, "")
	clt.sendMux.Lock()
	_ = clt.ws.WriteMessage(websocket.CloseMessage, data)
	clt.sendMux.Unlock()

	timeOut := time.After(time.Second * 1000)

	for {
		select {
		case <-timeOut:
			return errors.New("timeout")
		default:
			if clt.isReading || clt.isSending {
				time.Sleep(time.Millisecond * 100)
			} else {
				clt.ws = nil
				return nil
			}
		}
	}
}

func NewClient(options *Options) *Client {
	clt := &Client{
		options: options,
	}

	if clt.options.MessageHandler != nil {
		clt.messageHandler = clt.options.MessageHandler
	}

	return clt
}
