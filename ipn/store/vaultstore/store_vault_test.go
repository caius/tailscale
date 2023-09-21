package vaultstore

import (
	"testing"

	"tailscale.com/tstest"
)

func TestNewVaultStore(t *testing.T) {
	tstest.PanicOnLog()

	s, err := New(nil, "secret/tailscale")
	if err != nil {
		t.Fatal(err)
	}

	s.WriteState("foo", []byte("bar"))

	data, err := s.ReadState("foo")
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != "bar" {
		t.Fatal("data does not match")
	}
}
