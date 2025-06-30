package storage_test

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/acudac-com/public-go/storage"
)

// var Storage = storage.NewFsStorage(".storage")

var Storage = storage.NewGcsStorage(context.Background(), os.Getenv("GCS_BUCKET"), "public-go/storage")

func Test_ReadWrite(t *testing.T) {
	defer Storage.RemoveFolder(t.Context(), "")

	key := "users/123/test_file.txt"
	data := []byte("Hello world!")

	// Write
	Storage.Write(t.Context(), key, data)
	written := Storage.WriteIfMissing(t.Context(), key, data)
	if written {
		t.Fatalf("write should not have happened")
	}

	// Read
	readData, found := Storage.Read(t.Context(), key)
	if !found {
		t.Fatalf("%s not found", key)
	}
	if !reflect.DeepEqual(data, readData) {
		t.Fatalf("Read data does not match written data. Expected: %v, Got: %v", data, readData)
	}

	// Remove
	found = Storage.Remove(t.Context(), key)
	if !found {
		t.Fatalf("Remove failed: %s not found", key)
	}

	written = Storage.WriteIfMissing(t.Context(), key, data)
	if !written {
		t.Fatalf("write should have happened")
	}
}
