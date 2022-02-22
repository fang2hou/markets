package exchange

import (
	"Markets/pkg/database"
	"fmt"
	"testing"
)

func TestOkx(t *testing.T) {
	okx := NewOkx(map[string]string{
		"apiKey":   "this-is-api-key",
		"secret":   "this-is-secret",
		"password": "this-is-password",
	}, database.NewInteractor(database.NewInternalConnector()))

	if okx.authData.ApiKey != "this-is-api-key" {
		t.Error("API Key is not set correctly. Expected: this-is-api-key, got: ", okx.authData.ApiKey)
	}

	if okx.authData.ApiSecret != "this-is-secret" {
		t.Error("API secret is not set correctly. Expected: this-is-secret, got: ", okx.authData.ApiSecret)
	}

	if okx.authData.Passphrase != "this-is-password" {
		t.Error("API passphrase is not set correctly. Expected: this-is-password, got: ", okx.authData.Passphrase)
	}

	if err := okx.Start(); err != nil {
		fmt.Println("Can't start okx:", err)
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

	err := okx.SendPublicMessageJSON(&parameters)
	if err != nil {
		t.Errorf("Can't send message: %s", err)
	}

	if err := okx.Stop(); err != nil {
		fmt.Println("Can't stop okx", err)
	}
}
