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

func (i *Interactor) GetString(key string) (*string, error) {
	dataStringPointer, err := i.connector.Get(key)

	if err != nil {
		return nil, err
	}

	return dataStringPointer, nil
}

func (i *Interactor) SetString(key string, value *string) error {
	return i.connector.Set(key, value)
}

func (i *Interactor) GetMap(key string) (*map[string]interface{}, error) {
	dataStringPointer, err := i.connector.Get(key)

	if err != nil {
		return nil, err
	}

	var data map[string]interface{}

	if err := json.Unmarshal([]byte(*dataStringPointer), &data); err != nil {
		return nil, err
	}

	return &data, nil
}

func (i *Interactor) SetMap(key string, value *map[string]interface{}) error {
	if dataBytes, err := json.Marshal(value); err != nil {
		return err
	} else {
		dataString := string(dataBytes)
		return i.connector.Set(key, &dataString)
	}
}

func (i *Interactor) Delete(key string) error {
	return i.connector.Delete(key)
}

func (i *Interactor) GetBalance(exchangeName string, currency string) (*Balance, error) {
	key := i.GenerateKeyWithPath([]string{"BALANCE", exchangeName, currency})

	dataString, err := i.connector.Get(key)

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
	key := i.GenerateKeyWithPath([]string{"BALANCE", exchangeName, currency})
	if dataBytes, err := json.Marshal(balance); err != nil {
		return err
	} else {
		dataString := string(dataBytes)
		return i.connector.Set(key, &dataString)
	}
}

func (i *Interactor) GetFee(exchangeName string, currency string) (*Fee, error) {
	key := i.GenerateKeyWithPath([]string{"FEE", exchangeName, currency})

	dataStringPointer, err := i.connector.Get(key)

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
	key := i.GenerateKeyWithPath([]string{"FEE", exchangeName, currency})
	if dataBytes, err := json.Marshal(fee); err != nil {
		return err
	} else {
		dataString := string(dataBytes)
		return i.connector.Set(key, &dataString)
	}
}

func (i *Interactor) GetOrder(exchangeName string, currency string, orderId string) (*Order, error) {
	key := i.GenerateKeyWithPath([]string{"ORDER", exchangeName, currency, orderId})

	dataStringPointer, err := i.connector.Get(key)

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
	key := i.GenerateKeyWithPath([]string{"ORDER", exchangeName, currency, orderId})
	if dataBytes, err := json.Marshal(order); err != nil {
		return err
	} else {
		dataString := string(dataBytes)
		return i.connector.Set(key, &dataString)
	}
}

func (i *Interactor) GetOrderBook(exchangeName string, currency string) (*OrderBook, error) {
	key := i.GenerateKeyWithPath([]string{"ORDER-BOOK", exchangeName, currency})

	dataStringPointer, err := i.connector.Get(key)

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
	key := i.GenerateKeyWithPath([]string{"ORDER-BOOK", exchangeName, currency})
	if dataBytes, err := json.Marshal(orderBook); err != nil {
		return err
	} else {
		dataString := string(dataBytes)
		return i.connector.Set(key, &dataString)
	}
}
