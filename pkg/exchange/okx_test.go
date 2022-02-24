package exchange

import (
	"Markets/pkg/database"
	"fmt"
	"github.com/go-redis/redis/v8"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestOkx(t *testing.T) {
	// Okx Simulate API
	okx := NewOkx(
		map[string]string{
			"apiKey":   os.Getenv("TEST_OKX_API_KEY"),
			"secret":   os.Getenv("TEST_OKX_SECRET"),
			"password": os.Getenv("TEST_OKX_PASSPHASE"),
		},
		[]string{"STARL/USDT"},
		database.NewInteractor(database.NewInternalConnector()),
	)

	if okx.authData.ApiKey != os.Getenv("TEST_OKX_API_KEY") {
		t.Error(
			"API Key is not set correctly.",
			"Expected:", os.Getenv("TEST_OKX_API_KEY"),
			"got: ", okx.authData.ApiKey,
		)
	}

	if okx.authData.ApiSecret != os.Getenv("TEST_OKX_SECRET") {
		t.Error(
			"API secret is not set correctly.",
			"Expected:", os.Getenv("TEST_OKX_SECRET"),
			"got: ", okx.authData.ApiSecret,
		)
	}

	if okx.authData.Passphrase != os.Getenv("TEST_OKX_PASSPHASE") {
		t.Error(
			"API passphrase is not set correctly.",
			"Expected:", os.Getenv("TEST_OKX_PASSPHASE"),
			"got: ", okx.authData.Passphrase,
		)
	}

	if err := okx.Start(); err != nil {
		fmt.Println("Can't start okx:", err)
	}

	<-time.After(time.Second * 10)

	if err := okx.Stop(); err != nil {
		fmt.Println("Can't stop okx", err)
	}
}

func TestOkx_RestApi_(t *testing.T) {
	okx := NewOkx(
		map[string]string{
			"apiKey":   os.Getenv("TEST_OKX_API_KEY"),
			"secret":   os.Getenv("TEST_OKX_SECRET"),
			"password": os.Getenv("TEST_OKX_PASSPHASE"),
		},
		[]string{"STARL/USDT"},
		database.NewInteractor(database.NewRedisConnector(&redis.Options{
			Addr:     "localhost:6379",
			Password: "", // no password set
			DB:       0,  // use default DB
		})),
	)

	okx.restClient = &http.Client{}

	if data, err := okx.restApi(&restApiOption{
		method: "GET",
		path:   "/account/trade-fee",
		params: map[string]string{
			"instType": "SPOT",
			"instId":   "STARL-USDT",
		},
	}); err != nil {
		t.Error(err)
	} else {
		fmt.Println(string(data))
	}

	okx.updateFee()
}
