package main

import (
	"fmt"
	"github.com/hashicorp/vault/api"
	log "github.com/sirupsen/logrus"
)

type vault struct {
	c *api.Client
}

func NewVaultClient(vaultAddress string) (*vault, error) {
	// Reads token from VAULT_TOKEN automatically.
	c, err := api.NewClient(&api.Config{
		Address:      vaultAddress,
		AgentAddress: "https://idp.fadalax.tech",
	})
	if err != nil {
		log.WithError(err).Error("Failed to create Vault client")
		return nil, err
	}
	return &vault{c: c}, nil
}

func (v *vault) PKIRoleExists(name string) (bool, error) {
	return false, fmt.Errorf("not implemented")
}

func (v *vault) CreatePKIUser(name string) error {
	return fmt.Errorf("not implemented")
}
