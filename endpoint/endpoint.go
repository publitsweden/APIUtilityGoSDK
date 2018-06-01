// Copyright 2018 Publit Sweden AB. All rights reserved.

// Package endpoint is a common package can be used to construct common endpoints that fulfill the AdminAPI.endpointer interface
package endpoint

import (
	"errors"
	"fmt"
	"strings"
)

// Resource struct
type Resource struct {
	Endpoint   Endpoint
	Qualifiers []interface{}
	Endpoints  map[Endpoint]string
}

// Endpoint enumeration type
type Endpoint int

// Endpoints is a map of endpoints
var Endpoints map[Endpoint]string

// GetEndpoint is a method connected to Resource that fullfils the Enpointer interface as stated in publishing.
func (r Resource) GetEndpoint() (string, error) {
	e := r.Endpoints[r.Endpoint]

	end := e

	noOfQualifiers := strings.Count(e, "%v")
	if noOfQualifiers != len(r.Qualifiers) {
		return "", errors.New(fmt.Sprintf("Amount of qualifiers did not match expected. Got %v, expected %v", len(r.Qualifiers), noOfQualifiers))
	}

	if noOfQualifiers > 0 {
		end = fmt.Sprintf(e, r.Qualifiers...)
	}

	return end, nil
}
