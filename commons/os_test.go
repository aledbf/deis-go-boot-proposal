package commons

import (
	"testing"
)

func TestGetoptEmpty(t *testing.T) {
	value := Getopt("", "")
	if value != "" {
		t.Fatalf("Expected '' as value of empty env name", value)
	}
}

func TestGetoptValid(t *testing.T) {
	value := Getopt("valid", "value")
	if value != "value" {
		t.Fatalf("Expected 'value' as value of 'valid' but %s returned", value)
	}
}

func TestGetoptDefault(t *testing.T) {
	value := Getopt("", "default")
	if value != "default" {
		t.Fatalf("Expected 'default' as value of empty env name but %s returned", value)
	}
}
