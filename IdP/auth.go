package main

import (
	"context"
	"fmt"
	"github.com/coreos/go-oidc"
	log "github.com/sirupsen/logrus"
	"regexp"
)

const bearerToken = "(?i)^bearer (.*)" // case insensitive match for "Bearer someTokenHere"

type validator struct {
	provider       *oidc.Provider
	tokenExtractor *regexp.Regexp
	clientID       string
}

func NewValidator(issuer string, cid string) (*validator, error) {
	log.WithField("issuer-url", issuer).Info("Contacting OIDC Issuer...")
	p, err := oidc.NewProvider(context.Background(), issuer)
	if err != nil {
		log.WithError(err).Error("Failed to create OIDC Provider.")
		return nil, err
	}
	log.WithField("issuer", issuer).Info("Successfully created OIDC Provider for Issuer.")
	v := validator{provider: p, tokenExtractor: regexp.MustCompile(bearerToken), clientID: cid}
	return &v, nil
}

// Validate takes the whole authorization header and if it is a JWT, validates it.
func (v *validator) Validate(ctx context.Context, authHeader string) (string, error) {
	m := v.tokenExtractor.FindStringSubmatch(authHeader)
	if len(m) != 2 {
		return "", fmt.Errorf("malformed Authorization header")
	}
	verifier := v.provider.Verifier(&oidc.Config{ClientID: v.clientID})
	tok, err := verifier.Verify(context.Background(), m[1])
	if err != nil {
		return "", err
	}
	return tok.Subject, nil
}
