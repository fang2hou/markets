package config

import (
	"reflect"
	"testing"
)

func TestConfig(t *testing.T) {
	testConfig := Config{}
	testConfigFile := []byte(`
exchange:
  okx:
    apiKey: 123456
    secret: 123456
    password: 123456
  gateio:
    apiKey: 123456
    secret: 123456
currency:
  - STARL/USDT
  - BTC/USDT
log:
  enable: true
  path: logs
`)

	err := testConfig.Load(testConfigFile)
	if err != nil {
		t.Errorf("Config Load Error: '%s'", err)
	}

	if currencies, err := testConfig.GetCurrenciesSetting(); err != nil {
		t.Errorf("Config GetCurrencies Error: '%s'", err)
	} else if reflect.DeepEqual(currencies, []string{"STARL/USDT", "BTC/USDT"}) == false {
		t.Errorf("Config GetCurrencies Error: Expected '%v' Got '%v':", []string{"STARL/USDT", "BTC/USDT"}, currencies)
	}

	if exchange, err := testConfig.GetExchangeSetting("okx"); err != nil {
		t.Errorf("Config GetExchanges Error: '%s'", err)
	} else if reflect.DeepEqual(exchange, map[string]string{"apiKey": "123456", "secret": "123456", "password": "123456"}) == false {
		t.Errorf("Config GetExchanges Error: Expected '%v' Got '%v':", map[string]string{"apiKey": "123456", "secret": "123456", "password": "123456"}, exchange)
	}
}
