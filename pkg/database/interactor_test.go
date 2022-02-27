package database

import (
	"reflect"
	"testing"
)

func TestInteractor_Base(t *testing.T) {
	testString := "testing_string_1001"
	testMap := map[string]interface{}{
		"test_key_1001": "test_value_1001",
		"test_key_1002": "test_value_1002",
	}

	interactor := NewInteractor(NewInternalConnector())

	key := interactor.GenerateKeyWithPath([]string{"TEST", "GO", "a", "b", "c"})
	if key != "TEST.GO.a.b.c" {
		t.Errorf("Interactor GenerateKeyWithPath Error: Expected key to be 'a.b.c', got '%s'", key)
	}

	if err := interactor.SetString("Test", key, &testString); err != nil {
		t.Errorf("Interactor SetString Error: '%s'", err)
	}

	if dataStringPointer, err := interactor.GetString("Test", key); err != nil {
		t.Errorf("Interactor GetString Error: '%s'", err)
	} else if *dataStringPointer != testString {
		t.Errorf("Interactor GetString Error: Expected '%s', got '%s'", testString, *dataStringPointer)
	}

	if err := interactor.SetMap("Test", key, &testMap); err != nil {
		t.Errorf("Interactor SetMap Error: '%s'", err)
	}

	if dataMapPointer, err := interactor.GetMap("Test", key); err != nil {
		t.Errorf("Interactor GetMap Error: '%s'", err)
	} else if !reflect.DeepEqual(*dataMapPointer, testMap) {
		t.Errorf("Interactor GetMap Error: Expected '%v', got '%v'", testMap, *dataMapPointer)
	}

	if err := interactor.Delete("Test", key); err != nil {
		t.Errorf("Interactor Delete Error: '%s'", err)
	}
}

func TestInteractor_Balance(t *testing.T) {
	testBalance := Balance{
		Free:  100000,
		Used:  20000,
		Total: 120000,
	}

	interactor := NewInteractor(NewInternalConnector())

	if err := interactor.SetBalance("TestExchange", "TEST_CURRENCY", &testBalance); err != nil {
		t.Errorf("Interactor SetBalance Error: '%s'", err)
	}

	if dataPointer, err := interactor.GetBalance("TestExchange", "TEST_CURRENCY"); err != nil {
		t.Errorf("Interactor GetBalance Error: '%s'", err)
	} else if !reflect.DeepEqual(*dataPointer, testBalance) {
		t.Errorf("Interactor GetBalance Error: Expected '%v', got '%v'", testBalance, *dataPointer)
	}

	if err := interactor.Delete("Balance", "TestExchange.TEST_CURRENCY"); err != nil {
		t.Errorf("Interactor Delete Error: '%s'", err)
	}
}

func TestInteractor_Fee(t *testing.T) {
	testFee := Fee{
		Maker: 0.1,
		Taker: 0.2,
	}

	interactor := NewInteractor(NewInternalConnector())

	if err := interactor.SetFee("TestExchange", "TEST_CURRENCY", &testFee); err != nil {
		t.Errorf("Interactor SetFee Error: '%s'", err)
	}

	if dataPointer, err := interactor.GetFee("TestExchange", "TEST_CURRENCY"); err != nil {
		t.Errorf("Interactor GetFee Error: '%s'", err)
	} else if !reflect.DeepEqual(*dataPointer, testFee) {
		t.Errorf("Interactor GetFee Error: Expected '%v', got '%v'", testFee, *dataPointer)
	}

	if err := interactor.Delete("Fee", "TestExchange.TEST_CURRENCY"); err != nil {
		t.Errorf("Interactor Delete Error: '%s'", err)
	}
}

func TestInteractor_Order(t *testing.T) {
	testOrder := Order{
		Id:          "123456789",
		Type:        "limit",
		Side:        "buy",
		CreateTime:  "2022-01-01T00:00:00Z",
		UpdateTime:  "2022-01-01T00:00:00Z",
		Price:       0.0000026400,
		Amount:      1000000,
		LeftAmount:  20000,
		Status:      "partial_filled",
		Fee:         0.000001,
		FeeCurrency: "TEST_CURRENCY",
	}

	interactor := NewInteractor(NewInternalConnector())

	if err := interactor.SetOrder("TestExchange", "TEST_CURRENCY", testOrder.Id, &testOrder); err != nil {
		t.Errorf("Interactor SetOrder Error: '%s'", err)
	}

	if dataPointer, err := interactor.GetOrder("TestExchange", "TEST_CURRENCY", testOrder.Id); err != nil {
		t.Errorf("Interactor GetOrder Error: '%s'", err)
	} else if !reflect.DeepEqual(*dataPointer, testOrder) {
		t.Errorf("Interactor GetOrder Error: Expected '%v', got '%v'", testOrder, *dataPointer)
	}

	if err := interactor.Delete("Order", "TestExchange.TEST_CURRENCY."+testOrder.Id); err != nil {
		t.Errorf("Interactor Delete Error: '%s'", err)
	}
}

func TestInteractor_OrderBook(t *testing.T) {
	testOrderBook := OrderBook{
		Asks: map[string]string{
			"0.0000026400": "1000000",
			"0.0000026500": "20000",
		},
		Bids: map[string]string{
			"0.0000026200": "1000000",
			"0.0000026000": "20000",
			"0.0000025000": "10000",
		},
	}

	interactor := NewInteractor(NewInternalConnector())

	if err := interactor.SetOrderBook("TestExchange", "TEST_CURRENCY", &testOrderBook); err != nil {
		t.Errorf("Interactor SetOrderBook Error: '%s'", err)
	}

	if dataPointer, err := interactor.GetOrderBook("TestExchange", "TEST_CURRENCY"); err != nil {
		t.Errorf("Interactor GetOrderBook Error: '%s'", err)
	} else if !reflect.DeepEqual(*dataPointer, testOrderBook) {
		t.Errorf("Interactor GetOrderBook Error: Expected '%v', got '%v'", testOrderBook, *dataPointer)
	}

	if err := interactor.Delete("OrderBook", "TestExchange.TEST_CURRENCY"); err != nil {
		t.Errorf("Interactor Delete Error: '%s'", err)
	}
}
