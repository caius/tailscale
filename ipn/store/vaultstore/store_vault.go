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
	kv := s.client.KVv2(s.mountPath)

	data := map[string]interface{}{"data": string(bs)}
	_, err := kv.Put(context.Background(), fmt.Sprintf("%s/%s", s.secretKey, id), data)
	if err != nil {
		return err
	}

	return nil
}

func (s *Store) ReadState(id ipn.StateKey) ([]byte, error) {
	kv := s.client.KVv2(s.mountPath)

	key := fmt.Sprintf("%s/%s", s.secretKey, id)
	data, err := kv.Get(context.Background(), key)
	if err != nil {
		// TODO: distinguish between "not found" and other errors
		return nil, ipn.ErrStateNotExist
	}

	output := data.Data["data"].(string)
	return []byte(output), nil
}
