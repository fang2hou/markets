package database

import "errors"

// Connector is the interface that wraps the database accessing method.
type Connector interface {
	Get(name string) (*string, error)
	Set(name string, value *string) error
	Delete(name string) error
}

// InternalConnector is a connector that stores the values in memory.
type InternalConnector struct {
	storage map[string]string
}

func (c *InternalConnector) Set(key string, value *string) error {
	c.storage[key] = *value
	return nil
}

func (c *InternalConnector) Get(key string) (*string, error) {
	if value, ok := c.storage[key]; ok {
		return &value, nil
	}
	return nil, errors.New("KeyNotFound")
}

func (c *InternalConnector) Delete(key string) error {
	if _, ok := c.storage[key]; ok {
		delete(c.storage, key)
		return nil
	}

	return errors.New("KeyNotFound")
}
