package client

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	HTTP_GET  = "GET"
	HTTP_POST = "POST"
)

func TestNew(t *testing.T) {
	t.Parallel()
	aId := 94
	c := New(
		func(c *Client) {
			c.ClientId = aId
		},
	)

	if c.ClientId != aId {
		t.Errorf(`Client account id did not match expected. Expected "%v", but got "%v"`, aId, c.ClientId)
	}

	if c.HTTPClient == nil {
		t.Error("Expected http client to be set but was not.")
	}

	if c.Logger == nil {
		t.Error("Expected logger to be set but was not.")
	}

}

func TestGetResponseFromCall(t *testing.T) {
	t.Parallel()

	loginfocounter := 0
	dbginfocounter := 0
	c := New(
		func(c *Client) {
			c.HTTPClient = MockClient{
				SetTokenHeader: true,
			}
			c.Logger = &MockLogger{
				InfoCallback: func(message interface{}) {
					loginfocounter++
				},
				DebugCallback: func(message interface{}) {
					dbginfocounter++
				},
			}
		},
	)

	r := httptest.NewRequest(HTTP_GET, "http://someurl.test", nil)
	r.RequestURI = ""

	_, err := c.Call(r)

	if err != nil {
		t.Errorf("Received an error but did not expect one: %v", err.Error())
	}

	if dbginfocounter != 0 {
		t.Errorf("Expected no debug logs to be written but got %d.", dbginfocounter)
	}

	if loginfocounter != 2 {
		t.Errorf("Expected exactly two info log to be written but got: %d.", loginfocounter)
	}
}

func TestCallSetsBasicAuthHeaders(t *testing.T) {
	t.Parallel()
	c := New(
		func(c *Client) {
			c.User = "someuser"
			c.Password = "somepassword"
			c.HTTPClient = MockClient{}
			c.Logger = &MockLogger{}
		},
	)

	t.Run(
		"Without account id",
		func(t *testing.T) {
			r := httptest.NewRequest(HTTP_GET, "http://someurl.test", nil)
			r.RequestURI = ""

			_, err := c.Call(r)

			if err != nil {
				t.Errorf("Received an error but did not expect one: %v", err.Error())
			}

			//No account id
			authString := fmt.Sprintf("%v;:%v", c.User, c.Password)
			expectedBasic := "Basic " + b64enc(authString)

			if r.Header.Get("Authorization") != expectedBasic {
				t.Error("Basic auth header was not set, but was expected to be.")
			}
		},
	)

	t.Run(
		"With account id",
		func(t *testing.T) {
			c.ClientId = 1

			r := httptest.NewRequest(HTTP_GET, "http://someurl.test", nil)
			r.RequestURI = ""

			_, err := c.Call(r)

			if err != nil {
				t.Errorf("Received an error but did not expect one: %v", err.Error())
			}

			// With account id
			authString := fmt.Sprintf("%v;%v:%v", c.User, c.ClientId, c.Password)
			expectedBasic := "Basic " + b64enc(authString)

			if r.Header.Get("Authorization") != expectedBasic {
				t.Error("Basic auth header was not set, but was expected to be.")
			}
		},
	)
}

func TestCallSetsTokenHeaders(t *testing.T) {
	t.Parallel()
	token := "sometokenhash"
	c := New(
		func(c *Client) {
			c.User = "someuser"
			c.Password = "somepassword"
			c.ClientId = 1
			c.Token = "sometokenhash"
			c.HTTPClient = MockClient{}
			c.Logger = &MockLogger{}
		},
	)

	r := httptest.NewRequest(HTTP_GET, "http://someurl.test", nil)
	r.RequestURI = ""

	_, err := c.Call(r)

	if err != nil {
		t.Errorf("Received an error but did not expect one: %v", err.Error())
	}

	// Note that password is no longer part of the authString
	authString := fmt.Sprintf("%v;%v:", c.User, c.ClientId)
	expectedBasic := "Basic " + b64enc(authString)

	if r.Header.Get("Authorization") != expectedBasic {
		t.Error("Basic auth header was not set, but was expected to be.")
	}

	if r.Header.Get("token") != token {
		t.Error("Token header was not set but was expected to be.")
	}
}

func TestCallGetsResponse(t *testing.T) {
	t.Parallel()
	c := New(
		func(c *Client) {
			c.User = "someuser"
			c.Password = "somepassword"
			c.Token = "sometokenhash"
			c.Logger = &MockLogger{}
		},
	)

	message := []byte(`Received request`)

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Write(message)
	}

	// create test server with handler
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	req := httptest.NewRequest(HTTP_GET, ts.URL, nil)
	req.RequestURI = ""

	resp, err := c.Call(req)

	if err != nil {
		t.Error("Received an error but did not expect one.")
	}

	responseBody, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()

	if string(responseBody) != string(message) {
		t.Errorf(`Response body did not match expected. Received: "%v", expected "%v"`, string(responseBody), string(message))
	}
}

func TestCanSetAPIToken(t *testing.T) {
	t.Parallel()
	c := New(
		func(c *Client) {
			c.User = "someuser"
			c.Password = "somepassword"
			c.Logger = &MockLogger{}
		},
	)

	tokenHash := "somerandomtokenstring"

	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Token", tokenHash)
	}

	// create test server with handler
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	req := httptest.NewRequest(HTTP_POST, ts.URL, nil)
	req.RequestURI = ""

	err := c.SetNewAPIToken(req)

	if err != nil {
		t.Error("Received an error but did not expect one.", err.Error())
	}

	if c.Token != tokenHash {
		t.Error("Set token did not match expected.")
	}
}

func TestNewAPITokenCanHandleErrors(t *testing.T) {
	t.Parallel()
	c := New(
		func(c *Client) {
			c.User = "someuser"
			c.Password = "somepassword"
			c.Logger = &MockLogger{}
		},
	)

	t.Run(
		"If no header is returned",
		func(t *testing.T) {
			// Do not write a header
			handler := func(w http.ResponseWriter, r *http.Request) {}

			// create test server with handler
			ts := httptest.NewServer(http.HandlerFunc(handler))
			defer ts.Close()

			req := httptest.NewRequest(HTTP_POST, ts.URL, nil)
			req.RequestURI = ""

			err := c.SetNewAPIToken(req)

			if err == nil {
				t.Error("Did not receive an error but expected one.")
			}
		},
	)

	t.Run(
		"If http client returns error",
		func(t *testing.T) {
			c.HTTPClient = MockClient{false, true}

			r := httptest.NewRequest(HTTP_POST, "http://someurl.com", nil)
			err := c.SetNewAPIToken(r)

			if err == nil {
				t.Error("Did not receive an error but expected one.")
			}
		},
	)

}

func TestCanGetToken(t *testing.T) {
	t.Parallel()
	token := "sometoken"
	c := New(
		func(c *Client) {
			c.Token = token
		},
	)

	authT := c.GetAuthToken()

	if token != authT {
		t.Error("Token did not match expected.")
	}
}

func TestCanUnsetToken(t *testing.T) {
	t.Parallel()
	c := New(
		func(c *Client) {
			c.Token = "sometoken"
		},
	)

	c.UnsetAuthToken()

	if c.Token != "" {
		t.Error("Expected token to be unset but was not")
	}
}

func b64enc(str string) string {
	return base64.StdEncoding.EncodeToString([]byte(str))
}

type MockClient struct {
	SetTokenHeader bool
	ReturnError    bool
}

func (m MockClient) Do(r *http.Request) (*http.Response, error) {

	if m.ReturnError {
		return &http.Response{
			Status:     "BadRequest",
			StatusCode: http.StatusBadRequest,
		}, errors.New("Some error")
	}

	h := http.Header{}
	if m.SetTokenHeader {
		h.Set("token", "sometoken")
	}
	return &http.Response{
		Header:     h,
		Status:     "ok",
		StatusCode: http.StatusOK,
	}, nil
}

type MockLogger struct {
	DebugCallback func(message interface{})
	InfoCallback  func(message interface{})
}

func (l *MockLogger) Debug(message interface{}) {
	if l.DebugCallback != nil {
		l.DebugCallback(message)
	}
}

func (l *MockLogger) Info(message interface{}) {
	if l.InfoCallback != nil {
		l.InfoCallback(message)
	}
}

// EXAMPLES

func Example() {

	// Create new Client
	c := New()

	// Create new status_check request.
	r, err := http.NewRequest(HTTP_GET, "https://url.to.publit/v2.0/status_check", nil)
	if err != nil {
		log.Fatal(err)
	}

	// CallRaw (non authenticatied call)
	resp, err := c.CallRaw(r)
	if err != nil {
		log.Fatal(err)
	}

	// If status 200
	if resp.StatusCode == http.StatusOK {
		log.Println("Service is up!")
	}

	log.Println("Service is down.")
}

func ExampleNew() {
	// New accepts variadic functions for setting additional attributes to the created client.
	// This approach gives the implementer options for programmatically changing config options based on internal logic.

	// tokens slice. For illustrative example only. In real world this would more likely be a client config.
	tokens := []string{
		"",
		"sometokenhash",
	}

	// Range tokens.
	for _, v := range tokens {
		// Create a slice of func(c *CLient)
		clientSettings := []func(c *Client){
			func(c *Client) {
				c.User = "MyUserName"
			},
		}

		// Check if token is set for client.
		if v != "" {
			clientSettings = append(clientSettings, func(c *Client) {
				c.Token = v
			})
		} else {
			clientSettings = append(clientSettings, func(c *Client) {
				c.Password = "MyPassword"
			})
		}

		// Call new to create a client.
		c := New(clientSettings...)

		fmt.Printf("%s - %s\n", c.User, c.Password)
		// Output:
		// MyUserName - MyPassword
		// MyUserName -
	}
}

func ExampleClient_SetNewAPIToken() {
	c := New(
		func(c *Client) {
			c.User = "MyUserName"
			c.Password = "MyPassword"
			c.ClientId = 12345
		},
	)

	// Token should be empty as this stage.
	token := c.GetAuthToken()

	// If empty attempt to set it via the token endpoint.
	if token == "" {
		// Create new request.
		r, err := http.NewRequest(HTTP_POST, "https://url.to.publit/<api>/v2.0/token", nil)
		if err != nil {
			log.Fatal(err)
		}

		// Call SetNewAPIToken.
		err = c.SetNewAPIToken(r)
		if err != nil {
			log.Fatal(err)
		}
	}

	// Token should be set here.
	log.Println(c.GetAuthToken())
}