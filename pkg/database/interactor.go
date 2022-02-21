package database

import (
	"encoding/json"
	"strings"
)

//Interactor is the interface for interacting with the database
type Interactor struct {
	connector Connector
}

func (_ *Interactor) GenerateKeyWithPath(path []string) string {
	return strings.Join(path, ".")
}

func (i *Interactor) GetString(region string, key string) (*string, error) {
	dataStringPointer, err := i.connector.Get(region, key)

	if err != nil {
		return nil, err
	}

	return dataStringPointer, nil
}

func (i *Interactor) SetString(region string, key string, value *string) error {
	return i.connector.Set(region, key, value)
}

func (i *Interactor) GetMap(region string, key string) (*map[string]interface{}, error) {
	dataStringPointer, err := i.connector.Get(region, key)

	if err != nil {
		return nil, err
	}

	var data map[string]interface{}

	if err := json.Unmarshal([]byte(*dataStringPointer), &data); err != nil {
		return nil, err
	}

	return &data, nil
}

func (i *Interactor) SetMap(region string, key string, value *map[string]interface{}) error {
	if dataBytes, err := json.Marshal(value); err != nil {
		return err
	} else {
		dataString := string(dataBytes)
		return i.connector.Set(region, key, &dataString)
	}
}

func (i *Interactor) Delete(region string, key string) error {
	return i.connector.Delete(region, key)
}

func (i *Interactor) GetBalance(exchangeName string, currency string) (*Balance, error) {
	key := i.GenerateKeyWithPath([]string{exchangeName, currency})

	dataString, err := i.connector.Get("Balance", key)

	if err != nil {
		return nil, err
	}

	var data Balance

	if err := json.Unmarshal([]byte(*dataString), &data); err != nil {
		return nil, err
	}

	return &data, nil
}

func (i *Interactor) SetBalance(exchangeName string, currency string, balance *Balance) error {
	key := i.GenerateKeyWithPath([]string{exchangeName, currency})
	if dataBytes, err := json.Marshal(balance); err != nil {
		return err
	} else {
		dataString := string(dataBytes)
		return i.connector.Set("Balance", key, &dataString)
	}
}

func (i *Interactor) GetFee(exchangeName string, currency string) (*Fee, error) {
	key := i.GenerateKeyWithPath([]string{exchangeName, currency})

	dataStringPointer, err := i.connector.Get("Fee", key)

	if err != nil {
		return nil, err
	}

	var data Fee

	if err := json.Unmarshal([]byte(*dataStringPointer), &data); err != nil {
		return nil, err
	}

	return &data, nil
}

func (i *Interactor) SetFee(exchangeName string, currency string, fee *Fee) error {
	key := i.GenerateKeyWithPath([]string{exchangeName, currency})
	if dataBytes, err := json.Marshal(fee); err != nil {
		return err
	} else {
		dataString := string(dataBytes)
		return i.connector.Set("Fee", key, &dataString)
	}
}

func (i *Interactor) GetOrder(exchangeName string, currency string, orderId string) (*Order, error) {
	key := i.GenerateKeyWithPath([]string{exchangeName, currency, orderId})

	dataStringPointer, err := i.connector.Get("Order", key)

	if err != nil {
		return nil, err
	}

	var data Order

	if err := json.Unmarshal([]byte(*dataStringPointer), &data); err != nil {
		return nil, err
	}

	return &data, nil
}

func (i *Interactor) SetOrder(exchangeName string, currency string, orderId string, order *Order) error {
	key := i.GenerateKeyWithPath([]string{exchangeName, currency, orderId})
	if dataBytes, err := json.Marshal(order); err != nil {
		return err
	} else {
		dataString := string(dataBytes)
		return i.connector.Set("Order", key, &dataString)
	}
}

func (i *Interactor) GetOrderBook(exchangeName string, currency string) (*OrderBook, error) {
	key := i.GenerateKeyWithPath([]string{exchangeName, currency})

	dataStringPointer, err := i.connector.Get("OrderBook", key)

	if err != nil {
		return nil, err
	}

	var data OrderBook

	if err := json.Unmarshal([]byte(*dataStringPointer), &data); err != nil {
		return nil, err
	}

	return &data, nil
}

func (i *Interactor) SetOrderBook(exchangeName string, currency string, orderBook *OrderBook) error {
	key := i.GenerateKeyWithPath([]string{exchangeName, currency})
	if dataBytes, err := json.Marshal(orderBook); err != nil {
		return err
	} else {
		dataString := string(dataBytes)
		return i.connector.Set("OrderBook", key, &dataString)
	}
}

func NewInteractor(connector Connector) *Interactor {
	return &Interactor{
		connector: connector,
	}
}
