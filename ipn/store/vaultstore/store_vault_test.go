package vaultstore

import (
	"testing"

	kv "github.com/hashicorp/vault-plugin-secrets-kv"
	"github.com/hashicorp/vault/api"
	vaulthttp "github.com/hashicorp/vault/http"
	"github.com/hashicorp/vault/sdk/logical"
	hashivault "github.com/hashicorp/vault/vault"
	"tailscale.com/ipn"
)

func createVaultTestCluster(t *testing.T) *hashivault.TestCluster {
	t.Helper()

	coreConfig := &hashivault.CoreConfig{
		LogicalBackends: map[string]logical.Factory{
			"kv": kv.Factory,
		},
	}
	cluster := hashivault.NewTestCluster(t, coreConfig, &hashivault.TestClusterOptions{
		HandlerFunc: vaulthttp.Handler,
	})
	cluster.Start()

	// Create KV V2 mount
	if err := cluster.Cores[0].Client.Sys().Mount("kv", &api.MountInput{
		Type: "kv",
		Options: map[string]string{
			"version": "2",
		},
	}); err != nil {
		t.Fatal(err)
	}

	return cluster
}

// func TestNewVaultStore(t *testing.T) {
// 	tstest.PanicOnLog()

// 	s, err := New(nil, "secret:tailscale")
// 	if err != nil {
// 		t.Fatal(err)
// 	}

// 	t.Fatal(s)
// }

func TestVaultStoreRetrieve(t *testing.T) {
	// tstest.PanicOnLog()

	cluster := createVaultTestCluster(t)
	defer cluster.Cleanup()

	vaultClient := cluster.Cores[0].Client

	s := Store{
		client:    vaultClient,
		mountPath: "kv",
		secretKey: "tailscale",
	}
	s.WriteState("machineId", []byte("74889988-C2D2-4858-92EB-B51489F31E0E"))

	data, err := s.ReadState("machineId")
	if err != nil {
		t.Fatal(err)
	}

	if string(data) != "74889988-C2D2-4858-92EB-B51489F31E0E" {
		t.Fatal("data does not match")
	}
}

func TestVaultReadStateMissing(t *testing.T) {
	cluster := createVaultTestCluster(t)
	defer cluster.Cleanup()

	vaultClient := cluster.Cores[0].Client

	s := Store{
		client:    vaultClient,
		mountPath: "kv",
		secretKey: "tailscale",
	}

	_, err := s.ReadState("machineId")
	if err != ipn.ErrStateNotExist {
		t.Fatal("expected ErrStateNotExist")
	}
}
