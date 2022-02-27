package exchange

import (
	"Markets/pkg/database"
	"encoding/json"
	"net/http"
	"os"
	"testing"
)

func TestOkx(t *testing.T) {
	e := NewOkx(
		map[string]string{
			"apiKey":   os.Getenv("TEST_OKX_API_KEY"),
			"secret":   os.Getenv("TEST_OKX_SECRET"),
			"password": os.Getenv("TEST_OKX_PASSPHASE"),
		},
		[]string{"BTC/USDT"},
		database.NewInteractor(database.NewInternalConnector()),
	)

	if e.authData.ApiKey != os.Getenv("TEST_OKX_API_KEY") {
		t.Errorf("API Key is not set correctly.\nExpected:\n\t%s\nActual:\n\t%s",
			os.Getenv("TEST_OKX_API_KEY"), e.authData.ApiKey)
	}

	if e.authData.ApiSecret != os.Getenv("TEST_OKX_SECRET") {
		t.Errorf("API Secret is not set correctly.\nExpected:\n\t%s\nActual:\n\t%s",
			os.Getenv("TEST_OKX_SECRET"), e.authData.ApiSecret)
	}

	if e.authData.Passphrase != os.Getenv("TEST_OKX_PASSPHASE") {
		t.Errorf("API Passphrase is not set correctly.\nExpected:\n\t%s\nActual:\n\t%s",
			os.Getenv("TEST_OKX_PASSPHASE"), e.authData.Passphrase)
	}

	if err := e.Start(); err != nil {
		t.Error("Can't start okx:", err)
	}

	if err := e.Stop(); err != nil {
		t.Error("Can't stop okx", err)
	}
}

func TestOkx_RestApi_(t *testing.T) {
	e := NewOkx(
		map[string]string{
			"apiKey":   os.Getenv("TEST_OKX_API_KEY"),
			"secret":   os.Getenv("TEST_OKX_SECRET"),
			"password": os.Getenv("TEST_OKX_PASSPHASE"),
		},
		[]string{"BTC/USDT"},
		database.NewInteractor(database.NewInternalConnector()),
	)

	e.restClient = &http.Client{}

	if data, err := e.RestApi(&RestApiOption{
		method: "GET",
		path:   "/account/trade-fee",
		params: map[string]string{
			"instType": "SPOT",
			"instId":   "STARL-USDT",
		},
	}); err != nil {
		t.Error(err)
	} else {
		var dataMap map[string]interface{}

		if err := json.Unmarshal(data, &dataMap); err != nil {
			t.Error(err)
		} else {
			if dataMap["code"] != "0" {
				t.Errorf("Time not get correctly: %v", dataMap)
			}
		}
	}

	if err := e.updateFee(); err != nil {
		t.Error(err)
	}
}
