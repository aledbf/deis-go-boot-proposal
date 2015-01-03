package ceph

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsOSDInitialized(t *testing.T) {
	id := "0"

	exists := IsOSDInitialized(id)
	if exists {
		t.Fatalf("Expected '%v' as value of but %v returned", false, exists)
	}

	cephFile := "/var/lib/ceph/osd/ceph-" + id + "/keyring"
	err := os.MkdirAll(filepath.Dir(cephFile), 0755)
	defer os.RemoveAll("/var/lib/ceph")
	if err != nil {
		t.Fatal(err)
	}

	f, err := os.Create(cephFile)
	if err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	exists = IsOSDInitialized(id)
	if !exists {
		t.Fatalf("Expected '%v' as value of but %v returned", false, exists)
	}
}
