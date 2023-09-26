package exchange

import (
	"encoding/json"
	"os"
	"testing"

	"markets/pkg/database"
)

func TestGateio(t *testing.T) {
	e := NewGateio(
		map[string]string{
			"apiKey": os.Getenv("TEST_GATEIO_API_KEY"),
			"secret": os.Getenv("TEST_GATEIO_SECRET"),
		},
		[]string{"BTC/USDT"},
		database.NewInteractor(database.NewInternalConnector()),
	)

	if e.authData.ApiKey != os.Getenv("TEST_GATEIO_API_KEY") {
		t.Errorf("API Key is not set correctly.\nExpected:\n\t%s\nActual:\n\t%s",
			os.Getenv("TEST_GATEIO_API_KEY"), e.authData.ApiKey)
	}

	if e.authData.ApiSecret != os.Getenv("TEST_GATEIO_SECRET") {
		t.Errorf("API Secret is not set correctly.\nExpected:\n\t%s\nActual:\n\t%s",
			os.Getenv("TEST_GATEIO_SECRET"), e.authData.ApiSecret)
	}

	if err := e.Start(); err != nil {
		t.Error("Can't start gateio:", err)
	}

	if dataBytes, err := e.RestApi(&RestApiOption{
		method: "GET",
		path:   "/wallet/fee",
	}); err != nil {
		t.Error("Can't get fee:", err)
	} else {
		var data map[string]interface{}
		if err := json.Unmarshal(dataBytes, &data); err != nil {
			t.Error("Can't unmarshal fee:", err)
		} else {
			if _, ok := data["taker_fee"]; !ok {
				t.Error("Can't get taker fee:", data)
			}
		}
	}

	if err := e.Stop(); err != nil {
		t.Error("Can't stop okx", err)
	}
}
