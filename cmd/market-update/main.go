package main

import (
	"Markets/internal/pkg/config"
	"fmt"
	"io/ioutil"
)

func main() {
	dataBytes, err := ioutil.ReadFile("configs/config.yaml.example")
	if err != nil {
		return
	}

	testConfig := config.Config{}

	if err := testConfig.Load(dataBytes); err != nil {
		return
	}

	fmt.Printf("%+v\n", testConfig)
}
