package main

import (
	"context"
	"fmt"
	"github.com/hashicorp/vault/api"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
)

const alphanumeric = "[[:alnum:]]"

type vault struct {
	c   *api.Logical
	sys *api.Sys
}

func NewVaultClient(vaultAddress string, token string) (*vault, error) {
	// Reads token from VAULT_TOKEN automatically.
	c, err := api.NewClient(&api.Config{
		Address: vaultAddress,
	})
	if token != "" {
		c.SetToken(token)
	}
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

func NewVaultUserClient(vaultAddress string, name string, jwtoken string) (*vault, error) {
	// Reads token from VAULT_TOKEN automatically.
	c, err := api.NewClient(&api.Config{
		Address: vaultAddress,
	})
	m := regexp.MustCompile(bearerToken).FindStringSubmatch(jwtoken)
	if len(m) != 2 {
		return nil, fmt.Errorf("malformed Authorization header")
	}
	tok, err := c.Logical().Write("auth/jwt/login", map[string]interface{}{
		"role": name,
		"jwt":  m[1],
	})
	if err != nil {
		log.WithError(err).Error("Failed to login.")
		return nil, err
	}

	c.SetToken(tok.Data["client_token"].(string))
	if c.Token() == "" {
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
		"policy": fmt.Sprintf("path \"%s/*\" {capabilities = [ \"create\", \"read\", \"update\", \"delete\", \"list\", \"sudo\" ]}", mountPath),
	})
	if err != nil {
		l.WithError(err).Error("Failed to create policy.")
		return err
	}

	// Add oidc role
	_, err = v.c.Write(fmt.Sprintf("/auth/oidc/role/%s", name), map[string]interface{}{
		"bound_audiences":       "vault",
		"allowed_redirect_uris": "https://vault.fadalax.tech:8200/ui/vault/auth/oidc/oidc/callback",
		"user_claim":            "sub",
		"policies":              fmt.Sprintf("pki-user/%s", name),
		"bound_subject":         name,
	})
	if err != nil {
		l.WithError(err).Error("Failed to create oidc role.")
		return err
	}

	// Add jwt role
	_, err = v.c.Write(fmt.Sprintf("/auth/jwt/role/%s", name), map[string]interface{}{
		"bound_audiences": "vault",
		"user_claim":      "sub",
		"policies":        fmt.Sprintf("pki-user/%s", name),
		"bound_subject":   name,
		"role_type":       "jwt",
	})
	if err != nil {
		l.WithError(err).Error("Failed to create jwt role.")
		return err
	}

	return nil
}

func (v *vault) GetCert(ctx context.Context, name string) ([]byte, error) {
	l := log.WithField("name", name)
	if !regexp.MustCompile(alphanumeric).MatchString(name) {
		l.Error("Invalid name format.")
		return nil, fmt.Errorf("invalid name format")
	}
	mountPath := fmt.Sprintf("/pki-user/%s", name)
	cert, err := v.c.Write(fmt.Sprintf("%s/issue/%s", mountPath, name), map[string]interface{}{
		"common_name": fmt.Sprintf("%s@fadalax.tech", name),
		"ttl":         "336h",
	})
	if err != nil {
		l.WithError(err).Error("Failed to issue cert.")
		return nil, err
	}

	dir, err := ioutil.TempDir("", "cert")
	if err != nil {
		l.WithError(err).Error("Failed to create tmp dir.")
		return nil, err
	}

	defer os.RemoveAll(dir) // clean up

	priv := filepath.Join(dir, "private.key")
	if err := ioutil.WriteFile(priv, []byte(cert.Data["private_key"].(string)), 0666); err != nil {
		l.WithError(err).Error("Failed to write key.")
		return nil, err
	}
	certf := filepath.Join(dir, "cert.pem")
	if err := ioutil.WriteFile(certf, []byte(cert.Data["certificate"].(string)), 0666); err != nil {
		l.WithError(err).Error("Failed to write cert.")
		return nil, err
	}

	res, err := exec.CommandContext( ctx, "/usr/bin/openssl", "pkcs12", "-export", "-inkey", fmt.Sprintf("%s/%s", dir, "private.key"),
		"-in", fmt.Sprintf("%s/%s", dir, "cert.pem"), "-password", "pass:").Output()
	if err != nil {
		l.WithError(err).Error("Failed to convert file.")
		return nil, err
	}
	l.Info("Issued Certificate")

	return res, nil
}
