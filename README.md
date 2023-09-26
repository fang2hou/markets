# markets

## Introduction

This project is used to build a real-time cross-exchange price database. In addition to the prices, the program also automatically stores the commission information obtained through the exchange API. Based on the information in the database, you can set up strategies to trade robotically.

It's a free work for everyone, but if you use this library, the code you write need to be open source.

```
LICENSE: LGPL version 3
```

## Features

1. Support WebSocket API of various exchanges with very high performance.
2. Support for using both Rest API and WebSocket API in a single thread.
3. Multiple exchanges can be polled at the same time.
4. It is easy to extend the program to support more exchanges.

## Storage

The program only provides two types of storage:
1. Simply save in memory.
2. Redis

The other databases are also supported, but you need to write a connector for them in golang.  
Check the files in `pkg/database` if you want to know how to create a connector.

## Usage

Here is the sample code, just set your API token in the `config.yaml` file, and then run the program.

```go
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

```