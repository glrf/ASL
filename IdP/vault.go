package main

import (
	"fmt"
	"github.com/hashicorp/vault/api"
	log "github.com/sirupsen/logrus"
	"regexp"
)

const alphanumeric = "[[:alnum:]]"

type vault struct {
	c   *api.Logical
	sys *api.Sys
}

func NewVaultClient(vaultAddress string) (*vault, error) {
	// Reads token from VAULT_TOKEN automatically.
	c, err := api.NewClient(&api.Config{
		Address: vaultAddress,
	})
	if err != nil {
		log.WithError(err).Error("Failed to create Vault client")
		return nil, err
	}
	if c.Token() == "" {
		log.Error("No VAULT_TOKEN set.")
		return nil, fmt.Errorf("missing vault client token")
	}
	return &vault{c: c.Logical(), sys: c.Sys()}, nil
}

func (v *vault) PKIRoleExists(name string) (bool, error) {
	l := log.WithField("name", name)
	if !regexp.MustCompile(alphanumeric).MatchString(name) {
		l.Error("Invalid name format.")
		return false, fmt.Errorf("invalid name format")
	}
	role, err := v.c.Read(fmt.Sprintf("/auth/oidc/role/%s", name))
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

	// Mount new pki
	mountPath := fmt.Sprintf("/pki-user/%s", name)
	mountInput := &api.MountInput{
		Type:        "pki",
		Description: fmt.Sprintf("PKI for user %s", name),
		Config: api.MountConfigInput{
			MaxLeaseTTL: "43800h",
		},
	}
	err := v.sys.Mount(mountPath, mountInput)
	if err != nil {
		l.WithError(err).Error("Failed to mount PKI.")
		return err
	}

	// Generate intermediate
	interCAReq, err := v.c.Write(fmt.Sprintf("%s/intermediate/generate/internal", mountPath), map[string]interface{}{
		"common_name": fmt.Sprintf("%s.fadalax.tech", name),
	})
	if err != nil {
		l.WithError(err).Error("Failed to generate intermediate.")
		return err
	}
	l.Info(interCAReq.Data["csr"])
	// Sign intermediate
	interCA, err := v.c.Write("pki/root/sign-intermediate", map[string]interface{}{
		"csr":    interCAReq.Data["csr"],
		"format": "pem_bundle",
		"ttl":    "43800h",
	})
	if err != nil {
		l.WithError(err).Error("Failed to sign intermediate.")
		return err
	}
	// Set intermediate
	_, err = v.c.Write(fmt.Sprintf("%s/intermediate/set-signed", mountPath), map[string]interface{}{
		"certificate": interCA.Data["certificate"],
	})
	if err != nil {
		l.WithError(err).Error("Failed to set intermediate.")
		return err
	}

	// www.vaulptproject.io/api/secret/pki/index.html#create-update-role
	// Note that we depend on a lot of secure defaults here, such as key type and length.
	_, err = v.c.Write(fmt.Sprintf("%s/roles/%s", mountPath, name), map[string]interface{}{
		"allow_localhost":       false,
		"allowed_domains":       []string{fmt.Sprintf("%s@fadalax.tech", name)},
		"enforce_hostnames":     true,
		"allow_bare_domains":    true,
		"allow_ip_sans":         false,
		"server_flag":           false, // Should not be used as server certs.
		"client_flag":           true,  // Since we are want to generate certs which can be used for auth.
		"email_protection_flag": true,  // Emails are the other core purpose.
		"organization":          "imovies",
		"country":               "CH",
	})
	if err != nil {
		l.WithError(err).Error("Failed to create PKI role.")
		return err
	}

	// Add policy
	_, err = v.c.Write(fmt.Sprintf("/sys/policy%s", mountPath), map[string]interface{}{
		"policy": fmt.Sprintf("path \"%s\" {capabilities = [ \"create\", \"read\", \"update\", \"delete\", \"list\", \"sudo\" ]}", mountPath),
	})
	if err != nil {
		l.WithError(err).Error("Failed to create policy.")
		return err
	}

	// Add oidc role
	_, err = v.c.Write(fmt.Sprintf("/auth/oidc/role/%s", name), map[string]interface{}{
		"bound_audiences": "vault",
		"allowed_redirect_uris": "https://vault.fadalax.tech:8200/ui/vault/auth/oidc/oidc/callback",
		"user_claim": "sub",
		"policies": fmt.Sprintf("pki-user/%s", name),
		"bound_subject": name,
	})
	if err != nil {
		l.WithError(err).Error("Failed to create oidc role.")
		return err
	}

	/*
		vault write auth/oidc/role/kv-mgr \
			bound_audiences="$AUTH0_CLIENT_ID" \
			allowed_redirect_uris="http://127.0.0.1:8200/ui/vault/auth/oidc/oidc/callback" \
			allowed_redirect_uris="http://localhost:8250/oidc/callback" \
			user_claim="sub" \
			policies="reader" \
			groups_claim="https://example.com/roles"

	*/
	return nil
}
