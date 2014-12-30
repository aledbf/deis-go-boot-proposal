package commons

import (
	"testing"

	"github.com/deis/go-boot-proposal/commons"
)

func TestGetoptEmpty(t *testing.T) {
	value := commons.Getopt("", "")
	if value != "" {
		t.Fatalf("Expected '' as value of empty env name", value)
	}
}

func TestGetoptValid(t *testing.T) {
	value := commons.Getopt("valid", "value")
	if value != "value" {
		t.Fatalf("Expected 'value' as value of 'valid' but %s returned", value)
	}
}

func TestGetoptDefault(t *testing.T) {
	value := commons.Getopt("", "default")
	if value != "default" {
		t.Fatalf("Expected 'default' as value of empty env name but %s returned", value)
	}
}
