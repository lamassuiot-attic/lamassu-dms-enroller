package certs

import "math/big"

type CRT struct {
	ID             int
	Status         string
	Serial         *big.Int
	ExpirationDate string
	RevocationDate string
	CertPath       string
	DN             string
}

type CRTs struct {
	CRTs []CRT `json:"-"`
}
