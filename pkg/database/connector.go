package database

import (
	"context"
	"errors"
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

func (c *InternalConnector) Set(region string, key string, value *string) error {
	if c.storage[region] == nil {
		c.storage[region] = make(map[string]string)
	}

	c.storage[region][key] = *value
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
