package exchange

import (
	"Markets/pkg/database"
	"github.com/go-redis/redis/v8"
	"os"
	"testing"
)

func TestGateio_RestApi_(t *testing.T) {
	e := NewGateio(
		map[string]string{
			"apiKey": os.Getenv("TEST_GATEIO_API_KEY"),
			"secret": os.Getenv("TEST_GATEIO_SECRET"),
		},
		[]string{"STARL/USDT"},
		database.NewInteractor(database.NewRedisConnector(&redis.Options{
			Addr:     "localhost:6379",
			Password: "", // no password set
			DB:       0,  // use default DB
		})),
	)

	if e.authData.ApiKey != os.Getenv("TEST_GATEIO_API_KEY") {
		t.Error(
			"API Key is not set correctly.",
			"Expected:", os.Getenv("TEST_GATEIO_SECRET"),
			"got: ", e.authData.ApiKey,
		)
	}

	if e.authData.ApiSecret != os.Getenv("TEST_GATEIO_SECRET") {
		t.Error(
			"API secret is not set correctly.",
			"Expected:", os.Getenv("TEST_GATEIO_SECRET"),
			"got: ", e.authData.ApiSecret,
		)
	}

	if err := e.Start(); err != nil {
		t.Error("Can't start gateio:", err)
	}

	if err := e.updateBalance(); err != nil {
		t.Error("Can't update balance:", err)
	}
}
