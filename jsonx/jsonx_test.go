package jsonx_test

import (
	"testing"

	"github.com/acudac-com/public-go/jsonx"
)

type Payload struct {
	UserId string
	Role   string
}

var payload = &Payload{"12345", "Owner"}

func TestAll(t *testing.T) {
	for _, jx := range []*jsonx.Jsonx[*Payload]{
		jsonx.Marshaller[*Payload](),
		jsonx.B64Marshaller[*Payload](),
		jsonx.Encrypter[*Payload](),
		jsonx.B64Encrypter[*Payload](),
		jsonx.Hasher[*Payload](),
		jsonx.B64Hasher[*Payload](),
	} {
		data, err := jx.Marshal(t.Context(), payload)
		if err != nil {
			t.Errorf("Marshal failed: %v", err)
			continue
		}
		t.Logf("Marshalled data: %s", data)
		unmarshalled, err := jx.Unmarshal(t.Context(), data, &Payload{})
		if err != nil {
			t.Errorf("Unmarshal failed: %v", err)
			continue
		}
		if unmarshalled.UserId != payload.UserId || unmarshalled.Role != payload.Role {
			t.Errorf("Unmarshalled data does not match original: got %v, want %v", unmarshalled, payload)
		} else {
			t.Logf("Unmarshalled data matches original: %v", unmarshalled)
		}
	}
}

func BenchmarkMarshalling(b *testing.B) {
	// change the marshaller to the different ones to test them
	marshaller := jsonx.B64Encrypter[*Payload]()
	for i := 0; i < b.N; i++ {
		_, err := marshaller.Marshal(b.Context(), payload)
		if err != nil {
			b.Errorf("Marshal failed: %v", err)
		}
	}
}

func BenchmarkUnmarshalling(b *testing.B) {
	// change the marshaller to the different ones to test them
	marshaller := jsonx.B64Encrypter[*Payload]()
	data, err := marshaller.Marshal(b.Context(), payload)
	if err != nil {
		b.Fatalf("Marshal failed: %v", err)
	}
	for i := 0; i < b.N; i++ {
		_, err := marshaller.Unmarshal(b.Context(), data, &Payload{})
		if err != nil {
			b.Errorf("Unmarshal failed: %v", err)
		}
	}
}
