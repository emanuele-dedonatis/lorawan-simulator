package gateway

import "github.com/brocaar/lorawan"

type Gateway struct {
	EUI lorawan.EUI64

	discoveryURI string
}
