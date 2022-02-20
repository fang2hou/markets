package database

import "testing"

func TestConnector_InternalConnector(t *testing.T) {
	c := InternalConnector{
		storage: make(map[string]string),
	}

	testString := "testing1234567890"

	if err := c.Set("TEST.TEST_KEY", &testString); err != nil {
		t.Errorf("InternalConnector Set Error: %v", err)
	}

	if dataStringPointer, err := c.Get("TEST.TEST_KEY"); err != nil {
		t.Errorf("InternalConnector Get Error: %v", err)
	} else if *dataStringPointer != testString {
		t.Errorf("InternalConnector Get Error: The value from database is not as same as the one set before.")
	}

	if err := c.Delete("TEST.TEST_KEY"); err != nil {
		t.Errorf("InternalConnector Delete Error: %v", err)
	}
}
