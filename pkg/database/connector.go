package database

import (
	"context"
	"errors"
	"github.com/go-redis/redis/v8"
)

// Connector is the interface that wraps the database accessing method.
type Connector interface {
	Get(region string, name string) (*string, error)
	Set(region string, name string, value *string) error
	Delete(region string, name string) error
}

// InternalConnector is a connector that stores the values in memory.
type InternalConnector struct {
	storage map[string]map[string]string
}

func (c *InternalConnector) Set(region string, key string, valuePointer *string) error {
	if c.storage[region] == nil {
		c.storage[region] = make(map[string]string)
	}

	c.storage[region][key] = *valuePointer
	return nil
}

func (c *InternalConnector) Get(region string, key string) (*string, error) {
	if c.storage[region] == nil {
		return nil, errors.New("region not found")
	}

	if value, ok := c.storage[region][key]; ok {
		return &value, nil
	}

	return nil, errors.New("key not found")
}

func (c *InternalConnector) Delete(region string, key string) error {
	if c.storage[region] == nil {
		return errors.New("region not found")
	}

	if _, ok := c.storage[region][key]; ok {
		delete(c.storage[region], key)
		return nil
	}

	return errors.New("key not found")
}

func NewInternalConnector() *InternalConnector {
	return &InternalConnector{
		storage: make(map[string]map[string]string),
	}
}

// RedisConnector is a connector that stores the values in redis.
type RedisConnector struct {
	client  *redis.Client
	context context.Context
}

func (c *RedisConnector) Set(region string, key string, valuePointer *string) error {
	return c.client.HSet(c.context, region, key, *valuePointer).Err()
}

func (c *RedisConnector) Get(region string, key string) (*string, error) {
	if value, err := c.client.HGet(c.context, region, key).Result(); err != nil {
		return nil, err
	} else {
		return &value, nil
	}
}

func (c *RedisConnector) Delete(region string, key string) error {
	return c.client.HDel(c.context, region, key).Err()
}

func NewRedisConnector(options *redis.Options) *RedisConnector {
	return &RedisConnector{
		client:  redis.NewClient(options),
		context: context.Background(),
	}
}