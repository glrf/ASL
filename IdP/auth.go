package main

import (
	"context"
	"fmt"
	"github.com/coreos/go-oidc"
	log "github.com/sirupsen/logrus"
	"regexp"
	"time"
)

const bearerToken = "(?i)^bearer (.*)" // case insensitive match for "Bearer someTokenHere"

type validator struct {
	provider       *oidc.Provider
	tokenExtractor *regexp.Regexp
	hydra          HydraClient
}

func NewValidator(issuer string, hydra HydraClient) (*validator, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	log.WithField("issuer-url", issuer).Info("Contacting OIDC Issuer...")
	p, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		log.WithError(err).Error("Failed to create OIDC Provider.")
		return nil, err
	}
	log.WithField("issuer", issuer).Info("Successfully created OIDC Provider for Issuer.")
	v := validator{provider: p, tokenExtractor: regexp.MustCompile(bearerToken), hydra:hydra}
	return &v, nil
}

// Validate takes the whole authorization header and if it is a JWT, validates it.
func (v *validator) Validate(authHeader string) (string, error) {
	m := v.tokenExtractor.FindStringSubmatch(authHeader)
	if len(m) != 2 {
		return "", fmt.Errorf("malformed Authorization header")
	}
	return v.validate(m[1])
}

// validate returns the uid of a valid token, an error otherwise.
func (v *validator) validate(jwtToken string) (string, error) {
	return v.hydra.IntrospectToken(jwtToken)
}
