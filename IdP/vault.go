package main

import (
	"fmt"
	"github.com/hashicorp/vault/api"
	log "github.com/sirupsen/logrus"
	"regexp"
)

const alphanumeric = "[[:alnum:]]"

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
	l := log.WithField("name", name)
	if !regexp.MustCompile(alphanumeric).MatchString(name) {
		l.Error("Invalid name format.")
		return false, fmt.Errorf("invalid name format")
	}
	role, err := v.c.Read(fmt.Sprintf("/pki/roles/%s", name))
	if err != nil {
		l.WithError(err).WithField("name", name).Error("Failed to fetch PKI role.")
		return false, err
	}
	if role == nil {
		l.WithField("name", name).Debug("No PKI Role with this name.")
		return false, nil
	}
	return true, nil
}

func (v *vault) CreatePKIUser(name string) error {
	l := log.WithField("name", name)
	if !regexp.MustCompile(alphanumeric).MatchString(name) {
		l.Error("Invalid name format.")
		return fmt.Errorf("invalid name format")
	}
	// www.vaulptproject.io/api/secret/pki/index.html#create-update-role
	// Note that we depend on a lot of secure defaults here, such as key type and length.
	_, err := v.c.Write(fmt.Sprintf("/pki/roles/%s", name), map[string]interface{}{
		"allow_localhost": false,
		"allowed_domains": []string{fmt.Sprintf("%s@fadalax.tech", name)},
		"enforce_hostnames": true,
		"allow_ip_sans": false,
		"server_flag": false, // Should not be used as server certs.
		"client_flag": true, // Since we are want to generate certs which can be used for auth.
		"email_protection_flag": true, // Emails are the other core purpose.
		"organization": "imovies",
		"country": "CH",
	})
	if err != nil {
		l.WithError(err).Error("Failed to create PKI role.")
		return err
	}
	return fmt.Errorf("not implemented")
}
