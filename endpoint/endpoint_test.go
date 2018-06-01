package endpoint_test

import (
	"fmt"
	"reflect"
	"testing"

	. "github.com/publitsweden/APIUtilityGoSDK/endpoint"
)

type Endpointer interface {
	GetEndpoint() (string, error)
}

// Overload Endpoints variable
var TestEndpoints = map[Endpoint]string{
	1: "test",
	2: "test/%v",
	3: "test/%v;%v/%v",
}

func TestResourceImplementsEndpointer(t *testing.T) {
	t.Parallel()

	r := Resource{}
	ei := reflect.TypeOf((*Endpointer)(nil)).Elem()

	if !reflect.TypeOf(r).Implements(ei) {
		t.Error("Resource does not implement Endpointer but was epected to.")
	}
}

func TestCanGetEndpoints(t *testing.T) {
	t.Parallel()

	t.Run(
		"Can get plain endpoint",
		func(t *testing.T) {
			var index Endpoint = 1
			r := Resource{Endpoint: index, Endpoints: TestEndpoints}

			e, _ := r.GetEndpoint()

			if e != TestEndpoints[index] {
				t.Error("Retrieved endpoint did not match the expected.")
			}
		},
	)

	t.Run(
		"Can get endpoint with one qualifier",
		func(t *testing.T) {
			var index Endpoint = 2
			testId := 1
			expectedEndpoint := fmt.Sprintf("test/%v", testId)
			q := []interface{}{testId}
			r := Resource{Endpoint: index, Qualifiers: q, Endpoints: TestEndpoints}

			e, _ := r.GetEndpoint()

			if e != expectedEndpoint {
				t.Errorf("Endpoint did not match expected, Got %s, Expected %s", e, expectedEndpoint)
			}
		},
	)

	t.Run(
		"Can get endpoint with multiple qualifiers",
		func(t *testing.T) {
			var index Endpoint = 3
			q := []interface{}{"somestring", 2, "someotherstring"}
			expectedEndpoint := fmt.Sprintf("test/%v;%v/%v", q...)
			r := Resource{Endpoint: index, Qualifiers: q, Endpoints: TestEndpoints}

			e, _ := r.GetEndpoint()

			if e != expectedEndpoint {
				t.Errorf("Endpoint did not match expected, Got %s, Expected %s", e, expectedEndpoint)
			}
		},
	)
}

func TestReturnsErrorIfQualifiersDoesNotMatchExpected(t *testing.T) {
	t.Parallel()

	var index Endpoint = 3
	q := []interface{}{"somestring"}
	r := Resource{Endpoint: index, Qualifiers: q, Endpoints: TestEndpoints}

	_, err := r.GetEndpoint()

	if err == nil {
		t.Errorf("Did not receive an error but was expecting to.")
	}
}
