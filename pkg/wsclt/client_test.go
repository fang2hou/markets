package wsclt

import (
	"encoding/json"
	"fmt"
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
		t.Errorf("Connect error: %v", err)
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

	dataBytes, err := json.Marshal(parameters)
	if err != nil {
		t.Errorf("Marshal error: %v", err)
	}

	if err := clt.SendMessage(dataBytes); err != nil {
		t.Errorf("SendMessage error: %v", err)
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

	if reflect.DeepEqual(decodedMsg, expectedMsg) == false {
		t.Errorf("Unexpected message: %v", decodedMsg)
	}

	fmt.Println("try disconnect")
	clt.Close()
}
