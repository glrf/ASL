package main

import "fmt"

type vault struct {

}

func NewVaultClient() (*vault, error) {
	return nil, fmt.Errorf("not implemented")
}

func (v *vault) PKIRoleExists(name string) (bool, error) {
	return false, fmt.Errorf("not implemented")
}

func (v *vault) CreatePKIUser(name string) error {
	return fmt.Errorf("not implemented")
}