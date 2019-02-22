// Copyright 2018 Publit Sweden AB. All rights reserved.

// Package APIClient is an implementation of client.Client containing the most used Publit API-interface operations (GET, PUT, POST, DELETE)
package APIClient

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/publitsweden/APIUtilityGoSDK/common"
)

// General API constants
const (
	// Supported version of the Publit APIs
	API_VERSION = "v2.0"

	// Status check resource
	RESOURCE_STATUSCHECK = "status_check"

	// Token resource
	RESOURCE_TOKEN = "token"
)

// Endpointer interface declares how an endpoint should be defined
type Endpointer interface {
	GetEndpoint() (string, error)
}

// APICaller is an interface that defines how a client should use the Publit APIs.
// The github.com/publitsweden/APIUtilityGoSDK/client.Client fulfills this interface.
type APICaller interface {
	Call(r *http.Request) (*http.Response, error)
	CallRaw(r *http.Request) (*http.Response, error)
	SetNewAPIToken(r *http.Request) error
	UnsetAuthToken()
}

// APIClient hold Client information for connecting to the Publit APIs and base URLs.
type APIClient struct {
	Client  APICaller
	BaseURL string
	API     string
}

// StatusCheck checks if the Publit service is up.
func (c *APIClient) StatusCheck() (bool, error) {
	url, err := c.compileStatusCheckURL()

	if err != nil {
		return false, err
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)

	if err != nil {
		return false, err
	}

	// Use CallRaw since no authentication is needed for status check.
	r, err := c.Client.CallRaw(req)

	if err != nil {
		return false, err
	}

	if r.StatusCode != http.StatusOK {
		return false, nil
	}

	return true, nil
}

// Compiles statuscheck URL against the admin API.
func (c APIClient) compileStatusCheckURL() (string, error) {
	if c.BaseURL == "" {
		return "", errors.New("Could not compile status check URL. Missing APIClient.BaseURL")
	}

	return fmt.Sprintf("%s/%s/%s", c.BaseURL, API_VERSION, RESOURCE_STATUSCHECK), nil
}

// SetNewAPIToken creates and sets new token to client.
func (c APIClient) SetNewAPIToken() error {
	url, err := c.compileTokenURL()
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	return c.Client.SetNewAPIToken(req)
}

func (c APIClient) compileTokenURL() (string, error) {
	if c.BaseURL == "" || c.API == "" {
		return "", errors.New("Could not compile Token URL, missing one or both of APIClient.BaseURL or APIClient.API")
	}

	return fmt.Sprintf("%s/%s/%s/%s", c.BaseURL, c.API, API_VERSION, RESOURCE_TOKEN), nil
}

// Get Performs a GET method action against the Publit admin API.
func (c APIClient) Get(endpoint Endpointer, model interface{}, queryParams ...func(q url.Values)) error {
	epoint, err := endpoint.GetEndpoint()
	if err != nil {
		return err
	}

	endUrl := c.CompileEndpointURL(epoint)
	req, _ := http.NewRequest(http.MethodGet, endUrl, nil)

	q := req.URL.Query()
	for _, v := range queryParams {
		v(q)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := c.Client.Call(req)
	defer resp.Body.Close()

	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return MakeResponseError(resp)
	}

	err = json.NewDecoder(resp.Body).Decode(model)

	if err != nil {
		return err
	}

	return nil
}

// Post performs a POST method action against the Publit API.
func (c APIClient) Post(endpoint Endpointer, payload interface{}, result interface{}, headers ...func(h *http.Header)) error {
	return c.postPut(http.MethodPost, endpoint, payload, result, headers...)
}

// Put performs a PUT method action against the Publit API.
func (c APIClient) Put(endpoint Endpointer, payload interface{}, result interface{}, headers ...func(h *http.Header)) error {
	return c.postPut(http.MethodPut, endpoint, payload, result, headers...)
}

// postPut performs a post or put method action against the Publit admin API.
func (c APIClient) postPut(method string, endpoint Endpointer, payload interface{}, result interface{}, headers ...func(h *http.Header)) error {
	epoint, err := endpoint.GetEndpoint()
	if err != nil {
		return err
	}
	endUrl := c.CompileEndpointURL(epoint)

	body, err := json.Marshal(payload)

	if err != nil {
		return err
	}

	req, _ := http.NewRequest(method, endUrl, bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	h := &req.Header
	for _, v := range headers {
		v(h)
	}

	resp, err := c.Client.Call(req)
	if err != nil {
		return err
	}

	if resp.Body != nil {
		defer resp.Body.Close()
	}

	if resp.StatusCode != http.StatusOK {
		return MakeResponseError(resp)
	}

	err = json.NewDecoder(resp.Body).Decode(result)

	if err != nil {
		return err
	}

	return nil
}

// Delete performs a DELETE http call against the Publit API.
func (c APIClient) Delete(endpoint Endpointer, result interface{}, headers ...func(h *http.Header)) error {
	epoint, err := endpoint.GetEndpoint()
	if err != nil {
		return err
	}
	endUrl := c.CompileEndpointURL(epoint)
	req, _ := http.NewRequest(http.MethodDelete, endUrl, nil)

	h := &req.Header
	for _, v := range headers {
		v(h)
	}

	resp, err := c.Client.Call(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return MakeResponseError(resp)
	}

	err = json.NewDecoder(resp.Body).Decode(result)

	if err != nil {
		return err
	}

	return nil
}

// CompileEndpointURL compiles regular endpoints URL.
// Endpoints are defined in format baseurl / api / version / endpoint
func (c APIClient) CompileEndpointURL(endpoint string) string {
	return fmt.Sprintf("%v/%v/%v/%v", c.BaseURL, c.API, API_VERSION, endpoint)
}

// MakeResponseError attempts to make a better response error from response.
func MakeResponseError(resp *http.Response) error {
	if resp.Header.Get("Content-Type") == "application/json" {
		APIErr := &common.APIErrorResponse{}
		err := json.NewDecoder(resp.Body).Decode(APIErr)
		if err == nil && APIErr.HasInformation() { // Only return this error message if APIErr has information.
			return APIErr.GetAsError()
		}
	}

	// Special check for unauthorized reponse.
	if resp.StatusCode == http.StatusUnauthorized {
		return fmt.Errorf(`Unauthorized. Code: "%v"`, resp.StatusCode)
	}

	// Default
	return fmt.Errorf(`Response not ok. No information given. Code: "%v"`, resp.StatusCode)
}
