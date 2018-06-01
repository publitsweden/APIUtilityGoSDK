// Copyright 2017 Publit Sweden AB. All rights reserved.

// Publit API Client. Handles API client information and authentication of requests to the Publit APIs.
package client

import (
	"errors"
	"fmt"
	"net/http"
	"sync"

	"github.com/publitsweden/APIUtilityGoSDK/APILog"
)

// Client is a struct that holds credential information needed to connect to the Publit API.
// This is a generic object and does not in itself contain specific information needed to access endpoints.
// To connect to the API endpoints use the API libraries together with this.
type Client struct {
	//User name of the user attempting to authorise against the Publit APIs.
	User string
	// Password of the user attempting to authorise against the Publit APIs.
	Password string
	// AccountID the id of the Publit account the client wants to connect to.
	AccountID int
	// HTTPClient an object that implement the Doer interface.
	HTTPClient Doer
	// Token is the authorisation token that can be recieved from the Publit APIs.
	Token string
	// Logger is the logger object used for logging informational and debug messages.
	Logger Logger
	// M is a mutex and is used for not causing race-conditions on the Token attribute if several goroutines simultanously is trying to update it.
	M *sync.Mutex
}

// Doer is an interface representing the ability to do a request.
type Doer interface {
	// See https://golang.org/pkg/net/http/#Client.Do for more information.
	Do(r *http.Request) (*http.Response, error)
}

// Logger is an interface representing the ability to log debug and info messages.
type Logger interface {
	Debug(message interface{})
	Info(message interface{})
}

// New creates a New API Client.
// Automatically sets HTTPClient to http.DefaultClient and Logger to APILog.APILog if not explicitly set. And also sets an empty sync.Mutex to M.
func New(configFunc ...func(c *Client)) *Client {
	c := &Client{}
	c.M = &sync.Mutex{}

	for _, v := range configFunc {
		v(c)
	}

	if c.HTTPClient == nil {
		c.HTTPClient = http.DefaultClient
	}

	if c.Logger == nil {
		c.Logger = APILog.New()
	}

	return c
}

// Call performs an authenticated request defined by http.Request.
// Call automatically sets the authentication portion of the request.
func (c *Client) Call(r *http.Request) (*http.Response, error) {
	c.setAuth(r)
	return c.CallRaw(r)
}

// CallRaw performs request directly from http.Request (without automatic authentication).
func (c *Client) CallRaw(r *http.Request) (*http.Response, error) {
	c.Logger.Info(fmt.Sprintf("Calling URL: %s %s %s %s", r.Method, r.Host, r.URL.Path, r.URL.RawQuery))
	resp, err := c.HTTPClient.Do(r)

	if err != nil {
		c.Logger.Debug(err)
	}

	c.Logger.Info(fmt.Sprintf("Request URL: [%s %s %s] responded with status: %s %d", r.Method, r.Host, r.URL.Path, resp.Status, resp.StatusCode))

	// IF token is not set attempt to set it using the response from the request
	c.M.Lock()
	t := c.Token
	c.M.Unlock()

	if t == "" {
		// No need to handle token error here since that is not the main objective of this method
		c.setTokenFromResponse(resp)
	}

	return resp, err
}

// SetNewAPIToken performs a given *http.Request and sets Client.Token.
// Does not return any other information but errors if any occured.
func (c *Client) SetNewAPIToken(r *http.Request) error {
	c.setAuth(r)

	resp, err := c.Call(r)

	if err != nil {
		c.Logger.Debug(err)
		return err
	}

	err = c.setTokenFromResponse(resp)

	if err != nil {
		c.Logger.Debug(err)
		return err
	}

	return nil
}

func (c *Client) setTokenFromResponse(r *http.Response) error {
	token := r.Header.Get("token")
	if token == "" {
		err := errors.New("No token received in header. Could not set token from response.")
		c.Logger.Debug(err)
		return err

	}

	c.M.Lock()
	c.Token = token
	c.M.Unlock()

	return nil
}

func (c *Client) setAuth(r *http.Request) {
	username := c.User + ";"
	if c.AccountID != 0 {
		username = fmt.Sprintf("%v;%v", c.User, c.AccountID)
	}

	password := c.Password
	if c.Token != "" {
		r.Header.Set("token", c.Token)
		password = ""
	}

	r.SetBasicAuth(username, password)
}

// GetAuthToken getter for authentication token.
func (c *Client) GetAuthToken() string {
	return c.Token
}

// UnsetAuthToken unsets authentication token.
// If need to re-authenticate, this can be used to force re-authentication for the next call.
func (c *Client) UnsetAuthToken() {
	c.M.Lock()
	c.Token = ""
	c.M.Unlock()
}
