package wsclt

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"
)

func TestClient(t *testing.T) {
	// Use OKX Public WebSocket API to test the client
	var result []byte

	messageHandler := func(msg []byte) {
		if result == nil {
			result = msg
		}
	}

	clt := NewClient(&Options{
		SkipVerify:     false,
		PingInterval:   600 * time.Second,
		MessageHandler: messageHandler,
	})

	err := clt.Connect("wss://ws.okx.com:8443/ws/v5/public")
	if err != nil {
		t.Errorf("Connect error: %v", err)
	}

	err = clt.Connect("wss://ws.okx.com:8443/ws/v5/public")
	if err.Error() != "already connected" {
		t.Errorf("Connection state check error: %v", err)
	}

	parameters := map[string]interface{}{
		"op": "subscribe",
		"args": []map[string]interface{}{
			{
				"channel": "books50-l2-tbt",
				"instId":  "BTC-USDT",
			},
		},
	}

	if dataBytes, err := json.Marshal(parameters); err != nil {
		t.Errorf("Marshal error: %v", err)
	} else {
		if err := clt.SendMessage(dataBytes); err != nil {
			t.Errorf("SendMessage error: %v", err)
		}
	}

	for {
		if result != nil {
			break
		}
		<-time.After(time.Second)
	}

	var decodedMsg map[string]interface{}
	err = json.Unmarshal(result, &decodedMsg)

	if err != nil {
		t.Errorf("Unmarshal error: %v", err)
	}

	expectedMsg := map[string]interface{}{
		"event": "subscribe",
		"arg": map[string]interface{}{
			"channel": "books50-l2-tbt",
			"instId":  "BTC-USDT",
		},
	}

	if !reflect.DeepEqual(expectedMsg, decodedMsg) {
		t.Errorf("Expected message is\n\t%v, however the message is\n\t%v", expectedMsg, decodedMsg)
	}

	if err := clt.Close(); err != nil {
		t.Errorf("Close failed")
	}
}
