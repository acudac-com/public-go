package gobx_test

import (
	"testing"

	"github.com/acudac-com/public-go/encoding/gobx"
)

type Payload struct {
	UserId string
	Role   string
}

var payload = &Payload{"12345", "Owner"}

func Test_Marshal(t *testing.T) {
	marshalled := gobx.Marshal(payload)
	unmarshalled, err := gobx.Unmarshal(marshalled, &Payload{})
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if unmarshalled.UserId != payload.UserId || unmarshalled.Role != payload.Role {
		t.Errorf("Unmarshalled payload does not match original: got %v, want %v", unmarshalled, payload)
	}
}

func Test_B64Marshal(t *testing.T) {
	marshalled := gobx.MarshalB64(payload)
	unmarshalled, err := gobx.UnmarshalB64(marshalled, &Payload{})
	if err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}
	if unmarshalled.UserId != payload.UserId || unmarshalled.Role != payload.Role {
		t.Errorf("Unmarshalled payload does not match original: got %v, want %v", unmarshalled, payload)
	}
}

func Benchmark_Marshal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = gobx.Marshal(payload)
	}
}

func Benchmark_Unmarshal(b *testing.B) {
	marshalled := gobx.Marshal(payload)
	for i := 0; i < b.N; i++ {
		_, _ = gobx.Unmarshal(marshalled, &Payload{})
	}
}

func BenchmarkB64Marshal(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = gobx.MarshalB64(payload)
	}
}

func BenchmarkB64Unmarshal(b *testing.B) {
	marshalled := gobx.MarshalB64(payload)
	for i := 0; i < b.N; i++ {
		_, _ = gobx.UnmarshalB64(marshalled, &Payload{})
	}
}
