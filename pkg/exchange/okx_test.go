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
		t.Error(
			"API Key is not set correctly.",
			"Expected:", os.Getenv("TEST_OKX_API_KEY"),
			"got: ", e.authData.ApiKey,
		)
	}

	if e.authData.ApiSecret != os.Getenv("TEST_OKX_SECRET") {
		t.Error(
			"API secret is not set correctly.",
			"Expected:", os.Getenv("TEST_OKX_SECRET"),
			"got: ", e.authData.ApiSecret,
		)
	}

	if e.authData.Passphrase != os.Getenv("TEST_OKX_PASSPHASE") {
		t.Error(
			"API passphrase is not set correctly.",
			"Expected:", os.Getenv("TEST_OKX_PASSPHASE"),
			"got: ", e.authData.Passphrase,
		)
	}

	if err := e.Start(); err != nil {
		t.Error("Can't start okx:", err)
	}

	<-time.After(time.Second * 10)

	if err := e.Stop(); err != nil {
		fmt.Println("Can't stop okx", err)
	}
}

func TestOkx_Redis_(t *testing.T) {
	// Okx Simulate API
	e := NewOkx(
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

	if e.authData.ApiKey != os.Getenv("TEST_OKX_API_KEY") {
		t.Error(
			"API Key is not set correctly.",
			"Expected:", os.Getenv("TEST_OKX_API_KEY"),
			"got: ", e.authData.ApiKey,
		)
	}

	if e.authData.ApiSecret != os.Getenv("TEST_OKX_SECRET") {
		t.Error(
			"API secret is not set correctly.",
			"Expected:", os.Getenv("TEST_OKX_SECRET"),
			"got: ", e.authData.ApiSecret,
		)
	}

	if e.authData.Passphrase != os.Getenv("TEST_OKX_PASSPHASE") {
		t.Error(
			"API passphrase is not set correctly.",
			"Expected:", os.Getenv("TEST_OKX_PASSPHASE"),
			"got: ", e.authData.Passphrase,
		)
	}

	if err := e.Start(); err != nil {
		fmt.Println("Can't start okx:", err)
	}

	<-time.After(time.Second * 15)

	if err := e.Stop(); err != nil {
		fmt.Println("Can't stop okx", err)
	}
}

func TestOkx_RestApi_(t *testing.T) {
	e := NewOkx(
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
		fmt.Println(string(data))
	}

	e.updateFee()
}
