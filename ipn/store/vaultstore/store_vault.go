// Package vaultstore contains an ipn.StateStore implementation using Hashicorp Vault.
package vaultstore

import (
	"context"
	"fmt"
	"strings"

	vault "github.com/hashicorp/vault/api"

	"tailscale.com/ipn"
	"tailscale.com/types/logger"
)

type Store struct {
	client    *vault.Client
	mountPath string
	secretKey string
}

// keyPath should be in the format "mountPath:secretPath"
func New(_ logger.Logf, keyPath string) (*Store, error) {
	client, err := vault.NewClient(nil) // nil means pick up defaults from environment
	if err != nil {
		return nil, err
	}

	bits := strings.SplitN(keyPath, ":", 2)
	if len(bits) != 2 {
		return nil, fmt.Errorf("invalid secret key: %s", keyPath)
	}

	return &Store{
		client:    client,
		mountPath: bits[0],
		secretKey: bits[1],
	}, nil
}

func (s *Store) WriteState(id ipn.StateKey, bs []byte) error {
	// TODO: somehow figure out how we handle v1 vs v2
	kv := s.client.KVv1(s.mountPath)

	// Read existing data out before updating our key
	var data map[string]interface{}
	value, err := kv.Get(context.Background(), s.secretKey)
	if err != nil || value == nil {
		data = map[string]interface{}{}
	} else {
		data = value.Data
	}
	data[string(id)] = string(bs)

	err = kv.Put(context.Background(), s.secretKey, data)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) ReadState(id ipn.StateKey) ([]byte, error) {
	kv := s.client.KVv1(s.mountPath)

	data, err := kv.Get(context.Background(), s.secretKey)
	if err != nil {
		// TODO: distinguish between "not found" and other errors
		return nil, ipn.ErrStateNotExist
	}

	val, ok := data.Data[string(id)]
	if ok {
		return []byte(val.(string)), nil
	} else {
		return nil, ipn.ErrStateNotExist
	}
}
