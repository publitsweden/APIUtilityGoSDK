package APIClient_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"testing"

	. "github.com/publitsweden/APIUtilityGoSDK/APIClient"
	"github.com/publitsweden/APIUtilityGoSDK/client"
	"github.com/publitsweden/APIUtilityGoSDK/endpoint"
)

var TestAPI string = "someapi"

func TestCanCheckStatus(t *testing.T) {
	t.Parallel()

	t.Run(
		"If status is ok",
		func(t *testing.T) {
			caller := &MockAPICaller{}

			caller.Response = createCallerResponse(http.StatusOK, "")

			baseurl := "somebaseurl"

			c := &APIClient{Client: caller, BaseURL: baseurl, API: TestAPI}

			ok, _ := c.StatusCheck()

			if !ok {
				t.Error("Expected status check to pass, but received false.")
			}

			if c.GetLastResponseCode() != http.StatusOK {
				t.Errorf("Unexpected response code. Expected %d, got %d", http.StatusOK, c.GetLastResponseCode())
			}
		},
	)

	t.Run(
		"If status is not ok",
		func(t *testing.T) {
			caller := &MockAPICaller{}

			caller.Response = createCallerResponse(http.StatusBadRequest, "")

			baseurl := "somebaseurl"

			c := &APIClient{Client: caller, BaseURL: baseurl, API: TestAPI}

			ok, _ := c.StatusCheck()

			if ok {
				t.Error("Expected status check to fail, but received true.")
			}

			if c.GetLastResponseCode() != http.StatusBadRequest {
				t.Errorf("Unexpected response code. Expected %d, got %d", http.StatusBadRequest, c.GetLastResponseCode())
			}
		},
	)

	t.Run(
		"If call returns error",
		func(t *testing.T) {
			caller := &MockAPICaller{}

			caller.Response = createCallerResponse(http.StatusBadRequest, "")
			caller.ReturnErrors = true

			baseurl := "somebaseurl"

			c := &APIClient{Client: caller, BaseURL: baseurl, API: TestAPI}

			ok, err := c.StatusCheck()

			if ok {
				t.Error("Expected status check to fail, but received true.")
			}

			if err == nil {
				t.Error("Expected an error to be set but did not receive one.")
			}

			if c.GetLastResponseCode() != http.StatusBadRequest {
				t.Errorf("Unexpected response code. Expected %d, got %d", http.StatusBadRequest, c.GetLastResponseCode())
			}
		},
	)

	t.Run(
		"If baseURL is not set",
		func(t *testing.T) {
			caller := &MockAPICaller{}

			caller.Response = createCallerResponse(http.StatusBadRequest, "")
			caller.ReturnErrors = true

			c := &APIClient{
				Client: caller,
				API:    TestAPI,
			}

			ok, err := c.StatusCheck()

			if ok {
				t.Error("Expected status check to fail, but received true.")
			}

			if err == nil {
				t.Error("Expected an error to be set but did not receive one.")
			}
		},
	)
}

func TestCanSetNewAPIToken(t *testing.T) {
	t.Parallel()

	t.Run(
		"If call is ok",
		func(t *testing.T) {
			caller := &MockAPICaller{}
			caller.Response = createCallerResponse(http.StatusOK, ``)

			baseurl := "somebaseurl"

			c := &APIClient{Client: caller, BaseURL: baseurl, API: TestAPI}

			err := c.SetNewAPIToken()

			if err != nil {
				t.Error("Expected to be able to set new API token but got an error.")
			}
		},
	)

	t.Run(
		"If call returns error",
		func(t *testing.T) {
			caller := &MockAPICaller{}
			caller.Response = createCallerResponse(http.StatusOK, ``)
			caller.ReturnErrors = true

			baseurl := "somebaseurl"

			c := &APIClient{Client: caller, BaseURL: baseurl, API: TestAPI}

			err := c.SetNewAPIToken()

			if err == nil {
				t.Error("Expected an error when creating new APItoken but did not receive one.")
			}
		},
	)

	t.Run(
		"If the BaseURL or API are not set",
		func(t *testing.T) {
			caller := &MockAPICaller{}
			caller.Response = createCallerResponse(http.StatusOK, ``)
			caller.ReturnErrors = true

			c := &APIClient{}
			c.Client = caller

			err := c.SetNewAPIToken()

			if err == nil {
				t.Error("Expected an error when creating new APItoken but did not receive one.")
			}
		},
	)
}

func TestCanPerformGetRequestWithRawResponse(t *testing.T) {
	caller := &MockAPICaller{}
	expectedBody := `{"some":"body"}`
	caller.Response = createCallerResponse(http.StatusOK, expectedBody)

	baseurl := "somebaseurl"

	c := &APIClient{Client: caller, BaseURL: baseurl, API: TestAPI}

	resp, err := c.GetWithRawResponse(NewEndpoint())
	if err != nil {
		t.Error("Unexpected error.", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Unexpected status code. Got %d, expected %d", resp.StatusCode, http.StatusOK)
	}

	defer resp.Body.Close()

	body,_ := ioutil.ReadAll(resp.Body)

	if string(body) != string(expectedBody) {
		t.Errorf("Unexpected body. Expected %s, got %s", expectedBody, body)
	}
}

func TestCanPerformGetRequest(t *testing.T) {
	t.Parallel()

	caller := &MockAPICaller{}
	caller.Response = createCallerResponse(http.StatusOK, `{"some":"body"}`)

	baseurl := "somebaseurl"

	c := &APIClient{Client: caller, BaseURL: baseurl, API: TestAPI}

	model := &struct {
		Some string `json:"some"`
	}{}

	err := c.Get(NewEndpoint(), model)

	if err != nil {
		t.Error("Expected Get to pass but received error.", err.Error())
	}

	if model.Some != "body" {
		t.Error("Unmarshalled struct did not match expected.")
	}

	if c.GetLastResponseCode() != http.StatusOK {
		t.Errorf("Unexpected response code. Expected %d, got %d", http.StatusOK, c.GetLastResponseCode())
	}
}

func TestGetReturnsErrorIfCallFails(t *testing.T) {
	t.Parallel()

	caller := &MockAPICaller{}
	caller.Response = createCallerResponse(http.StatusOK, `{"some":"body"}`)

	caller.ReturnErrors = true

	baseurl := "somebaseurl"

	c := &APIClient{Client: caller, BaseURL: baseurl, API: TestAPI}

	model := &struct{}{}
	err := c.Get(NewEndpoint(), model)

	if err == nil {
		t.Error("Expected an error due to call failed but did not receive one.")
	}
}

func TestGetReturnsErrorIfStatusCodeNotOk(t *testing.T) {
	t.Parallel()

	caller := &MockAPICaller{}
	caller.Response = createCallerResponse(http.StatusBadRequest, `{"some":"body"}`)

	baseurl := "somebaseurl"

	c := &APIClient{Client: caller, BaseURL: baseurl, API: TestAPI}

	model := &struct{}{}
	err := c.Get(NewEndpoint(), model)

	if err == nil {
		t.Error("Expected an error due to status not ok but did not receive one.")
	}

	if c.GetLastResponseCode() != http.StatusBadRequest {
		t.Errorf("Unexpected response code. Expected %d, got %d", http.StatusBadRequest, c.GetLastResponseCode())
	}
}

func TestGetReturnsErrorIfBodyCanNotBeUnmarshalled(t *testing.T) {
	t.Parallel()

	caller := &MockAPICaller{}
	caller.Response = createCallerResponse(http.StatusOK, `{"some","much:faulty":,%€%;"body"}`)

	baseurl := "somebaseurl"

	c := &APIClient{Client: caller, BaseURL: baseurl, API: TestAPI}
	interf := &struct{}{}
	err := c.Get(NewEndpoint(), interf)

	if err == nil {
		t.Error("Expected an error due to unmarshalling errors but did not receive one.")
	}

	if c.GetLastResponseCode() != http.StatusOK {
		t.Errorf("Unexpected response code. Expected %d, got %d", http.StatusOK, c.GetLastResponseCode())
	}
}

func TestGetReturnsErrorIfEndpointerReturnsAnError(t *testing.T) {
	t.Parallel()

	caller := &MockAPICaller{}
	baseurl := "somebaseurl"

	c := &APIClient{Client: caller, BaseURL: baseurl, API: TestAPI}

	ep := NewEndpoint()
	ep.ShouldFail = true

	interf := &struct{}{}
	err := c.Get(ep, interf)

	if err == nil {
		t.Error("Expected an error due to endpointer errors but did not receive one.")
	}
}

func TestCanPerformPOSTRequest(t *testing.T) {
	t.Parallel()
	caller := &MockAPICaller{}
	caller.T = t

	i := struct {
		Name string `json:"name"`
	}{Name: "test"}
	j := &i

	caller.CallTestCallback = func(t *testing.T, r *http.Request) {
		defer r.Body.Close()
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Error("Got error was not expecting one.")
		}

		ic := i

		json.Unmarshal(b, ic)

		if ic.Name != i.Name {
			t.Error("Request body did not match expected.")
		}
	}

	baseurl := "somebaseurl"
	caller.Response = createCallerResponse(http.StatusOK, `{"name":"newTestName"}`)

	c := &APIClient{Client: caller, BaseURL: baseurl, API: TestAPI}

	err := c.Post(NewEndpoint(), &i, j)

	if err != nil {
		t.Error("Received an error but was not expecting to.")
	}

	if i.Name != "newTestName" {
		t.Error("Struct did not have expected value.")
	}

	if c.GetLastResponseCode() != http.StatusOK {
		t.Errorf("Unexpected response code. Expected %d, got %d", http.StatusOK, c.GetLastResponseCode())
	}
}

func TestPostReturnsErrorIfEndpointerReturnsAnError(t *testing.T) {
	t.Parallel()

	caller := &MockAPICaller{}
	baseurl := "somebaseurl"

	c := &APIClient{Client: caller, BaseURL: baseurl, API: TestAPI}

	i := struct {
		Name string `json:"name"`
	}{Name: "test"}
	j := &i
	ep := NewEndpoint()
	ep.ShouldFail = true
	err := c.Post(ep, &i, j)

	if err == nil {
		t.Error("Expected an error due to endpointer errors but did not receive one.")
	}
}

func TestCanPerformPUTRequest(t *testing.T) {
	t.Parallel()
	caller := &MockAPICaller{}
	caller.T = t

	i := struct {
		Name string `json:"name"`
	}{Name: "test"}
	j := &i

	caller.CallTestCallback = func(t *testing.T, r *http.Request) {
		defer r.Body.Close()
		b, err := ioutil.ReadAll(r.Body)
		if err != nil {
			t.Error("Got error was not expecting one.")
		}

		ic := i

		json.Unmarshal(b, ic)

		if ic.Name != i.Name {
			t.Error("Request body did not match expected.")
		}
	}

	baseurl := "somebaseurl"
	caller.Response = createCallerResponse(http.StatusOK, `{"name":"newTestName"}`)

	c := &APIClient{Client: caller, BaseURL: baseurl, API: TestAPI}

	err := c.Put(NewEndpoint(), &i, j)

	if err != nil {
		t.Error("Received an error but was not expecting to.")
	}

	if i.Name != "newTestName" {
		t.Error("Struct did not have expected value.")
	}

	if c.GetLastResponseCode() != http.StatusOK {
		t.Errorf("Unexpected response code. Expected %d, got %d", http.StatusOK, c.GetLastResponseCode())
	}
}

func TestCanPerformDeleteRequest(t *testing.T) {
	t.Parallel()
	caller := &MockAPICaller{}
	caller.T = t

	i := struct {
		Name string `json:"name"`
	}{}

	baseurl := "somebaseurl"
	caller.Response = createCallerResponse(http.StatusOK, `{"name":"newTestName"}`)

	c := &APIClient{Client: caller, BaseURL: baseurl, API: TestAPI}

	err := c.Delete(NewEndpoint(), &i)

	if err != nil {
		t.Error("Received an error but was not expecting to.")
	}

	if i.Name != "newTestName" {
		t.Error("Struct did not have expected value.")
	}

	if c.GetLastResponseCode() != http.StatusOK {
		t.Errorf("Unexpected response code. Expected %d, got %d", http.StatusOK, c.GetLastResponseCode())
	}
}

func TestDeleteReturnsErrorIfEndpointerReturnsAnError(t *testing.T) {
	t.Parallel()

	caller := &MockAPICaller{}
	baseurl := "somebaseurl"

	c := &APIClient{Client: caller, BaseURL: baseurl, API: TestAPI}

	i := struct {
		Name string `json:"name"`
	}{}
	ep := NewEndpoint()
	ep.ShouldFail = true
	err := c.Delete(ep, &i)

	if err == nil {
		t.Error("Expected an error due to endpointer errors but did not receive one.")
	}
}

func TestPostPutErrors(t *testing.T) {
	t.Parallel()

	table := []struct {
		TestName string
		TestFunc func(t *testing.T)
	}{
		{
			TestName: "If model can not be marshalled into json",
			TestFunc: func(t *testing.T) {
				caller := &MockAPICaller{}

				// Can not serialize and marshal a chan.
				i := make(chan int)

				baseurl := "somebaseurl"

				c := &APIClient{Client: caller, BaseURL: baseurl, API: TestAPI}

				// Run method through POST (would be just as fine with PUT
				err := c.Post(NewEndpoint(), &i, &i)
				if err == nil {
					t.Error("Did not receive an error, but was expecting to.")
				}
			},
		},
		{
			TestName: "If Call returns error",
			TestFunc: func(t *testing.T) {
				caller := &MockAPICaller{}
				caller.ReturnErrors = true
				caller.Response = createCallerResponse(http.StatusOK, "")

				baseurl := "somebaseurl"
				c := &APIClient{Client: caller, BaseURL: baseurl, API: TestAPI}

				i := struct{}{}
				// Run method through POST (would be just as fine with PUT
				err := c.Post(NewEndpoint(), &i, &i)

				if err == nil {
					t.Error("Did not receive an error, but was expecting to.")
				}
			},
		},
		{
			TestName: "If response status code is not ok",
			TestFunc: func(t *testing.T) {
				caller := &MockAPICaller{}
				caller.Response = createCallerResponse(http.StatusBadRequest, "")

				baseurl := "somebaseurl"
				c := &APIClient{Client: caller, BaseURL: baseurl, API: TestAPI}

				i := struct{ Name string }{}
				err := c.Post(NewEndpoint(), &i, &i)

				if err == nil {
					t.Error("Did not receive an error, but was expecting to.")
				}
			},
		},
		{
			TestName: "If response json can not be marshalled to interface",
			TestFunc: func(t *testing.T) {
				caller := &MockAPICaller{}
				caller.Response = createCallerResponse(http.StatusOK, `{"somwire,ddgd:""jsonstructur,,,:"newTestName"}`)

				baseurl := "somebaseurl"
				c := &APIClient{Client: caller, BaseURL: baseurl, API: TestAPI}

				i := struct {
					Name string `json:"name"`
				}{Name: "test"}
				err := c.Post(NewEndpoint(), &i, &i)

				if err == nil {
					t.Error("Did not receive an error, but was expecting to.")
				}
			},
		},
	}

	for _, v := range table {
		t.Run(
			v.TestName,
			v.TestFunc,
		)
	}
}

func TestCanMakeResponseError(t *testing.T) {
	t.Parallel()

	t.Run(
		"By parsing response",
		func(t *testing.T) {
			errorMessage := []byte(`{"Code":400,"Type":"BadRequest","Errors":[{"Info":"Some error","Type":"BadRequest"}],"CombinedInfo":"Some error"}`)

			resp := &http.Response{
				Status:     "400 BadRequest",
				StatusCode: http.StatusBadRequest,
				Header: http.Header{
					"Content-Type": []string{"application/json"},
				},
				Body: ioutil.NopCloser(bytes.NewBuffer(errorMessage)),
			}

			e := MakeResponseError(resp)

			if e == nil {
				t.Error("No error recieved but was expecting one.")
			}

			expected := `Code: "400", Type: "BadRequest", Combined info: "Some error"`
			if e.Error() != expected {
				t.Error("Error message did not match expected.")
			}
		},
	)
	t.Run(
		"Unauthorized error without body",
		func(t *testing.T) {
			errorMessage := []byte(`{}`)

			resp := &http.Response{
				Status:     "401 Unauthorized",
				StatusCode: http.StatusUnauthorized,
				Body:       ioutil.NopCloser(bytes.NewBuffer(errorMessage)),
			}

			e := MakeResponseError(resp)

			if e == nil {
				t.Error("No error recieved but was expecting one.")
			}

			expected := `Unauthorized. Code: "401"`
			if e.Error() != expected {
				t.Errorf(`Error message did not match expected. Got: "%v", Expected "%v"`, e.Error(), expected)
			}
		},
	)
	t.Run(
		"Default error without body and not unauthorized.",
		func(t *testing.T) {
			errorMessage := []byte(`{}`)

			resp := &http.Response{
				Status:     "400 BadRequest",
				StatusCode: http.StatusBadRequest,
				Body:       ioutil.NopCloser(bytes.NewBuffer(errorMessage)),
			}

			e := MakeResponseError(resp)

			if e == nil {
				t.Error("No error recieved but was expecting one.")
			}

			expected := `Response not ok. No information given. Code: "400"`
			if e.Error() != expected {
				t.Errorf(`Error message did not match expected. Got: "%v", Expected "%v"`, e.Error(), expected)
			}
		},
	)
}

func TestCanUnsetAuthToken(t *testing.T) {
	unsetAuthTokenCount := 0
	caller := &MockAPICaller{
		UnsetAuthTokenCallback: func() { unsetAuthTokenCount++ },
	}
	c := &APIClient{Client: caller, BaseURL: "somebaseurl", API: TestAPI}

	c.UnsetAuthToken()

	if unsetAuthTokenCount != 1 {
		t.Errorf("Expected UnsetAuthToken to run exactly 1 times, but ran %d times.", unsetAuthTokenCount)
	}
}

// TestREsponseErrors by running the status check a few times
// Since it's the easiest request
func TestResponseErrors(t *testing.T) {
	caller := &MockAPICaller{}
	baseurl := "somebaseurl"

	c := &APIClient{Client: caller, BaseURL: baseurl, API: TestAPI}

	codes := []int{
		http.StatusOK,
		http.StatusBadRequest,
		http.StatusUnauthorized,
	}

	for _, code := range codes {
		caller.Response = createCallerResponse(code, "")

		c.StatusCheck()

		if c.GetLastResponseCode() != code {
			t.Errorf("Unexpected response code. Expected %d, got %d", code, c.GetLastResponseCode())
		}
	}

	allCodes := c.GetResponseCodes()

	if len(allCodes) != len(codes) {
		t.Error("The lentgth of the returned codes did not match the expected.")
	}

}

func createCallerResponse(status int, body string) *http.Response {
	resp := &http.Response{}
	resp.StatusCode = status

	if body != "" {
		resp.Body = ioutil.NopCloser(bytes.NewBufferString(body))
	}

	return resp
}

type MockAPICaller struct {
	ReturnErrors           bool
	Response               *http.Response
	CallTestCallback       func(t *testing.T, r *http.Request)
	UnsetAuthTokenCallback func()
	T                      *testing.T
}

func (c *MockAPICaller) Call(r *http.Request) (*http.Response, error) {
	if c.ReturnErrors {
		return c.Response, errors.New("Some error")
	}

	if c.CallTestCallback != nil {
		c.CallTestCallback(c.T, r)
	}

	return c.Response, nil
}

func (c *MockAPICaller) CallRaw(r *http.Request) (*http.Response, error) {
	return c.Call(r)
}

func (c *MockAPICaller) SetNewAPIToken(r *http.Request) error {
	if c.ReturnErrors {
		return errors.New("Some error")
	}

	return nil
}

func (c *MockAPICaller) UnsetAuthToken() {
	if c.UnsetAuthTokenCallback != nil {
		c.UnsetAuthTokenCallback()
	}
}

// Creates new endpoint.
func NewEndpoint() Endpoint { return Endpoint{1, false} }

// For fulfilling the endpointer interface.
type Endpoint struct {
	Index      int
	ShouldFail bool
}

// For fulfilling the endpointer interface.
func (e Endpoint) GetEndpoint() (string, error) {
	if e.ShouldFail {
		return "", errors.New("some error")
	}
	return "someendpoint", nil
}

// EXAMPLES

func ExampleAPIClient() {
	// Create new APIClient
	c := &APIClient{}
	c.Client = client.New()
	c.BaseURL = "https://test.publit.com"
	c.API = "publishing"

	// Create function handler to wrap the APIClient.Get() method around.
	// Note that a real world situation would probably call for this to be a method in a resource package.
	// The Post, Put and Delete methods of the APIClient works in a similar fashion.
	countryIndexFunc := func(c *APIClient, r *endpoint.Resource, resp interface{}, queryParams ...func(q url.Values)) (interface{}, error) {
		err := c.Get(r, resp, queryParams...)
		return resp, err
	}

	// Create a resource with an index
	// Using the countries resource here as an example
	var index endpoint.Endpoint = 1
	r := &endpoint.Resource{
		Endpoint: index,
		Endpoints: map[endpoint.Endpoint]string{
			index: "countries",
		},
	}

	// Create a model representing the countries endpoint in the PublitAPI
	countries := &struct {
		Data []struct {
			ISO string `json:"iso3"`
		}
	}{}

	// Call the index function
	_, err := countryIndexFunc(c, r, countries)

	// Handle error
	if err != nil {
		log.Fatal(err)
	}

	// Print model
	fmt.Printf("Received countries: %+v", countries)
}

func ExampleAPIClient_StatusCheck() {
	// Create new APIClient
	c := &APIClient{}
	c.Client = client.New()
	c.BaseURL = "https://test.publit.com"

	// Check status
	isup, err := c.StatusCheck()

	// Handle error
	if err != nil {
		log.Fatal(err)
	}

	// Check if service is down
	if !isup {
		fmt.Println("Service is down")
	}

	fmt.Println("Service is up")
}

func ExampleAPIClient_SetNewAPIToken() {
	// Create new APIClient
	c := &APIClient{}
	c.Client = client.New()
	c.BaseURL = "https://test.publit.com"
	c.API = "apiname"

	// Set token
	err := c.SetNewAPIToken()

	// Handle error
	if err != nil {
		log.Fatal(err)
	}
}
