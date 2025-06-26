package jsonx_test

import (
	"testing"

	"github.com/acudac-com/public-go/encoding/jsonx"
)

type Payload struct {
	UserId string
	Role   string
}

var payload = &Payload{"12345", "Owner"}

func Test_Marshal(t *testing.T) {
	marshalled := jsonx.Marshal(t.Context(), payload)
	unmarshalled, err := jsonx.Unmarshal(t.Context(), marshalled, &Payload{})
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if unmarshalled.UserId != payload.UserId || unmarshalled.Role != payload.Role {
		t.Errorf("Unmarshalled payload does not match original: got %v, want %v", unmarshalled, payload)
	}
}

func Test_B64Marshal(t *testing.T) {
	marshalled := jsonx.MarshalB64(t.Context(), payload)
	unmarshalled, err := jsonx.UnmarshalB64(t.Context(), marshalled, &Payload{})
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if unmarshalled.UserId != payload.UserId || unmarshalled.Role != payload.Role {
		t.Errorf("Unmarshalled payload does not match original: got %v, want %v", unmarshalled, payload)
	}
}

func Benchmark_Marshal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = jsonx.Marshal(b.Context(), payload)
	}
}

func Benchmark_Unmarshal(b *testing.B) {
	marshalled := jsonx.Marshal(b.Context(), payload)
	for i := 0; i < b.N; i++ {
		_, _ = jsonx.Unmarshal(b.Context(), marshalled, &Payload{})
	}
}

func BenchmarkB64Marshal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = jsonx.MarshalB64(b.Context(), payload)
	}
}

func BenchmarkB64Unmarshal(b *testing.B) {
	marshalled := jsonx.MarshalB64(b.Context(), payload)
	for i := 0; i < b.N; i++ {
		_, _ = jsonx.UnmarshalB64(b.Context(), marshalled, &Payload{})
	}
}
