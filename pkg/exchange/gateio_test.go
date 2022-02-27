package exchange

import (
	"Markets/pkg/database"
	"github.com/go-redis/redis/v8"
	"os"
	"testing"
	"time"
)

func TestGateio_Redis_(t *testing.T) {
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

	//if dataBytes, err := e.RestApi(&RestApiOption{
	//	method: "GET",
	//	path:   "/wallet/fee",
	//}); err != nil {
	//	t.Error("Can't get fee:", err)
	//} else {
	//	var data map[string]interface{}
	//	if err := json.Unmarshal(dataBytes, &data); err != nil {
	//		t.Error("Can't unmarshal fee:", err)
	//	} else {
	//		if _, ok := data["taker_fee"]; !ok {
	//			t.Error("Can't get taker fee:", data)
	//		}
	//	}
	//}

	//if err := e.SendMessageJSON(map[string]interface{}{
	//	"channel": "spot.order_book_update",
	//	"event":   "subscribe",
	//	"payload": []string{"STARL_USDT", "100ms"},
	//}); err != nil {
	//	t.Error("Can't subscribe order book update:", err)
	//}

	<-time.After(time.Second * 100)
}
