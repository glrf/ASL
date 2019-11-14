package main

import (
	"fmt"
	"github.com/hashicorp/vault/api"
	log "github.com/sirupsen/logrus"
)

type vault struct {
	c *api.Logical
}

func NewVaultClient(vaultAddress string) (*vault, error) {
	// Reads token from VAULT_TOKEN automatically.
	c, err := api.NewClient(&api.Config{
		Address:      vaultAddress,
	})
	if err != nil {
		log.WithError(err).Error("Failed to create Vault client")
		return nil, err
	}
	if c.Token() == "" {
		log.Error("No VAULT_TOKEN set.")
		return nil, fmt.Errorf("missing vault client token")
	}
	return &vault{c: c.Logical()}, nil
}

func (v *vault) PKIRoleExists(name string) (bool, error) {
	// TODO(bimmlerd) this might be a vuln, have not thought about it much
	s, err := v.c.Read(fmt.Sprintf("/pki/roles/%s", name))
	if err != nil {
		log.WithError(err).WithField("name", name).Error("Failed to fetch PKI role.")
		return false, err
	}
	if s == nil {
		log.WithField("name", name).Debug("No PKI Role with this name.")
		return false, nil
	}
	return true, nil
}

func (v *vault) CreatePKIUser(name string) error {
	return fmt.Errorf("not implemented")
}
