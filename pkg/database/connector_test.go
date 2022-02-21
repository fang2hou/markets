package database

import (
	"context"
	"github.com/go-redis/redis/v8"
	"testing"
)

func TestConnector_InternalConnector(t *testing.T) {
	c := InternalConnector{
		storage: make(map[string]map[string]string),
	}

	testString := "testing1234567890"

	if err := c.Set("TEST", "TEST_KEY", &testString); err != nil {
		t.Errorf("InternalConnector Set Error: %v", err)
	}

	if dataStringPointer, err := c.Get("TEST", "TEST_KEY"); err != nil {
		t.Errorf("InternalConnector Get Error: %v", err)
	} else if *dataStringPointer != testString {
		t.Errorf("InternalConnector Get Error: The value from database is not as same as the one set before.")
	}

	if err := c.Delete("TEST", "TEST_KEY"); err != nil {
		t.Errorf("InternalConnector Delete Error: %v", err)
	}
}

func TestConnector_RedisConnector_(t *testing.T) {
	c := RedisConnector{
		client: redis.NewClient(&redis.Options{
			Addr:     "localhost:6379",
			Password: "", // no password set
			DB:       0,  // use default DB
		}),
		context: context.Background(),
	}

	testString := "testing1234567890"

	if err := c.Set("TEST", "TEST_KEY", &testString); err != nil {
		t.Errorf("RedisConnector Set Error: %v", err)
	}

	if dataStringPointer, err := c.Get("TEST", "TEST_KEY"); err != nil {
		t.Errorf("RedisConnector Get Error: %v", err)
	} else if *dataStringPointer != testString {
		t.Errorf("RedisConnector Get Error: The value from database is not as same as the one set before.")
	}

	if err := c.Delete("TEST", "TEST_KEY"); err != nil {
		t.Errorf("RedisConnector Delete Error: %v", err)
	}
}
