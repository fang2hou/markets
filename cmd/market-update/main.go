package main

import (
	"os"

	"github.com/go-redis/redis/v8"

	"markets/internal/pkg/config"
	"markets/pkg/database"
	"markets/pkg/exchange"
)

func pollExchange(e exchange.Exchanger) {
	forever := make(chan bool)

	if err := e.Start(); err != nil {
		panic(err)
	}

	defer func() {
		if err := e.Stop(); err != nil {
			panic(err)
		}
	}()

	<-forever
}

func main() {
	dataBytes, err := os.ReadFile("configs/config.yaml")
	if err != nil {
		panic(err)
	}

	cfg := config.Config{}

	if err := cfg.Load(dataBytes); err != nil {
		panic(err)
	}

	var currencies []string

	if value, err := cfg.GetCurrenciesSetting(); err != nil {
		panic(err)
	} else {
		currencies = value
	}

	if setting, err := cfg.GetExchangeSetting("okx"); err != nil {
		panic(err)
	} else {
		e := exchange.NewOkx(
			setting,
			currencies,
			database.NewInteractor(database.NewRedisConnector(&redis.Options{
				Addr: "localhost:6379",
			})),
		)

		go pollExchange(e)
	}

	if setting, err := cfg.GetExchangeSetting("gateio"); err != nil {
		panic(err)
	} else {
		e := exchange.NewGateio(
			setting,
			currencies,
			database.NewInteractor(database.NewRedisConnector(&redis.Options{
				Addr: "localhost:6379",
			})),
		)

		go pollExchange(e)
	}

	forever := make(chan bool)
	<-forever
}
