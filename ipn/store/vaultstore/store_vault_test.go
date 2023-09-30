package vaultstore

import (
	"context"
	"testing"

	hclog "github.com/hashicorp/go-hclog"
	kv "github.com/hashicorp/vault-plugin-secrets-kv"
	"github.com/hashicorp/vault/api"
	vaulthttp "github.com/hashicorp/vault/http"
	"github.com/hashicorp/vault/sdk/logical"
	hashivault "github.com/hashicorp/vault/vault"
	"tailscale.com/ipn"
	"tailscale.com/tstest"
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
		Logger:      hclog.NewNullLogger(),
	})
	cluster.Start()

	// Create KV V2 mount
	if err := cluster.Cores[0].Client.Sys().Mount("kv", &api.MountInput{
		Type: "kv",
		Options: map[string]string{
			"version": "1",
		},
	}); err != nil {
		t.Fatal(err)
	}

	return cluster
}

func TestVaultReadStateMissing(t *testing.T) {
	tstest.PanicOnLog()

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

func TestVaultReadStatePresent(t *testing.T) {
	tstest.PanicOnLog()

	cluster := createVaultTestCluster(t)
	defer cluster.Cleanup()

	vaultClient := cluster.Cores[0].Client

	s := Store{
		client:    vaultClient,
		mountPath: "kv",
		secretKey: "tailscale",
	}

	kvClient := vaultClient.KVv1("kv")
	data := map[string]interface{}{
		"machineId": "74889988-C2D2-4858-92EB-B51489F31E0E",
	}
	err := kvClient.Put(context.Background(), s.secretKey, data)
	if err != nil {
		t.Fatal(err)
	}

	value, err := s.ReadState("machineId")
	if err != nil {
		t.Fatal(err)
	}

	if string(value) != "74889988-C2D2-4858-92EB-B51489F31E0E" {
		t.Fatal("value does not match")
	}
}

func TestVaultStoreRoundtrip(t *testing.T) {
	tstest.PanicOnLog()

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

func TestVaultUpdateExistingData(t *testing.T) {
	tstest.PanicOnLog()

	cluster := createVaultTestCluster(t)
	defer cluster.Cleanup()

	vaultClient := cluster.Cores[0].Client

	s := Store{
		client:    vaultClient,
		mountPath: "kv",
		secretKey: "tailscale",
	}

	err := s.WriteState("one", []byte("1"))
	if err != nil {
		t.Fatal(err)
	}

	err = s.WriteState("two", []byte("2"))
	if err != nil {
		t.Fatal(err)
	}

	value, err := s.ReadState("one")
	if err != nil {
		t.Fatal(err)
	}
	if string(value) != "1" {
		t.Fatal("One did not match expected value")
	}

	value2, err := s.ReadState("two")
	if err != nil {
		t.Fatal(err)
	}
	if string(value2) != "2" {
		t.Fatal("One did not match expected value")
	}
}
