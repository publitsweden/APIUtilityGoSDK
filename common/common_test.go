package common

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestCanSetLimitQueryString(t *testing.T) {
	t.Parallel()
	q := url.Values{}

	limit := 1
	offset := 0
	ParamFunc := QueryLimit(limit, offset)

	ParamFunc(q)

	expected := fmt.Sprintf("%v,%v", offset, limit)

	assertQueryStringEqual(QUERY_KEY_LIMIT, expected, q, t)
}

func TestCanSetWithQueryString(t *testing.T) {
	t.Parallel()
	q := url.Values{}

	withs := []string{"withOne", "withTwo"}
	ParamFunc := QueryWith(withs...)

	ParamFunc(q)

	expected := strings.Join(withs, ",")

	assertQueryStringEqual(QUERY_KEY_WITH, expected, q, t)
}

func TestCanSetScopeQueryString(t *testing.T) {
	t.Parallel()
	q := url.Values{}

	scopes := []Scope{{Scope: "someScope", Filter: "filter"}, {Scope: "someOtherScope"}}
	ParamFunc := QueryScope(scopes)

	ParamFunc(q)

	expected := "someScope;filter,someOtherScope"

	assertQueryStringEqual(QUERY_KEY_SCOPE, expected, q, t)
}

func TestCanSetAuxiliaryQueryString(t *testing.T) {
	t.Parallel()
	q := url.Values{}

	aux := []string{"auxOne", "auxTwo"}
	ParamFunc := QueryAuxiliary(aux...)

	ParamFunc(q)

	expected := strings.Join(aux, ",")

	assertQueryStringEqual(QUERY_KEY_AUX, expected, q, t)
}

func TestOrderDirEnumCanBeViewedAsString(t *testing.T) {
	t.Parallel()
	orders := map[string]OrderDir{
		"ASC":  ORDER_DIR_ASC,
		"DESC": ORDER_DIR_DESC,
	}

	for expected, c := range orders {
		if c.AsString() != expected {
			t.Error("Order direction string did not match expected.")
		}
	}
}

func TestOrderDirEnumCanOnlyBeSetToExistingValues(t *testing.T) {
	t.Parallel()
	var o OrderDir = 42

	defer func() {
		if r := recover(); r == nil {
			t.Error("AsString did not panic but was expected to")
		}
	}()

	o.AsString()
}

func TestCanSetOrderByQueryString(t *testing.T) {
	t.Parallel()
	t.Run(
		"With direction",
		func(t *testing.T) {
			q := url.Values{}

			orderby := []string{"attr1", "attr2"}
			direction := ORDER_DIR_DESC
			ParamFunc := QueryOrderBy(orderby, direction)

			ParamFunc(q)

			expected := strings.Join(orderby, ",")

			assertQueryStringEqual(QUERY_KEY_ORDER, expected, q, t)

			assertQueryStringEqual(QUERY_KEY_ORDER_DIR, direction.AsString(), q, t)
		},
	)
	t.Run(
		"Without direction",
		func(t *testing.T) {
			q := url.Values{}

			orderby := []string{"attr1", "attr2"}
			var direction OrderDir = 0
			ParamFunc := QueryOrderBy(orderby, direction)

			ParamFunc(q)

			expected := strings.Join(orderby, ",")

			assertQueryStringEqual(QUERY_KEY_ORDER, expected, q, t)

			if q.Get(QUERY_KEY_ORDER_DIR) != "" {
				t.Error("Query had order_dir param, but was not expected to.")
			}
		},
	)
}

func TestOperatorEnumCanBeViewedAsString(t *testing.T) {
	t.Parallel()
	operators := map[string]Operator{
		"EQUAL":         OPERATOR_EQUAL,
		"GREATER_EQUAL": OPERATOR_GREATER_EQUAL,
		"GREATER":       OPERATOR_GREATER,
		"LESS_EQUAL":    OPERATOR_LESS_EQUAL,
		"LESS":          OPERATOR_LESS,
		"NOT_EQUAL":     OPERATOR_NOT_EQUAL,
	}
	for expected, o := range operators {
		t.Run(
			expected,
			func(t *testing.T) {
				str := o.AsString()

				if str != expected {
					t.Error("Operator string did not match expected.")
				}
			},
		)
	}
}

func TestOperatorEnumCanOnlyBeSetToExistingValues(t *testing.T) {
	t.Parallel()
	var o Operator = 42

	defer func() {
		if r := recover(); r == nil {
			t.Error("AsString did not panic but was expected to")
		}
	}()

	o.AsString()
}

func TestCombinatorEnumCanBeViewedAsString(t *testing.T) {
	t.Parallel()
	combinators := map[string]Combinator{
		"AND": COMBINATOR_AND,
		"OR":  COMBINATOR_OR,
	}

	for expected, c := range combinators {
		t.Run(
			expected,
			func(t *testing.T) {
				str := c.AsString()

				if str != expected {
					t.Error("Combinator string did not match expected.")
				}
			},
		)
	}
}

func TestCombinatorEnumCanOnlyBeSetToExistingValues(t *testing.T) {
	t.Parallel()
	var c Combinator = 42

	defer func() {
		if r := recover(); r == nil {
			t.Error("AsString did not panic but was expected to")
		}
	}()

	c.AsString()
}

func TestCanSetAttributeQuery(t *testing.T) {
	t.Parallel()
	t.Run(
		"With args",
		func(t *testing.T) {
			q := url.Values{}

			attrs := []AttrQuery{
				{
					Name:  "myAttr",
					Value: "myValue",
					Args: AttrArgs{
						Operator:   []Operator{OPERATOR_EQUAL},
						Combinator: []Combinator{COMBINATOR_AND},
					},
				},
			}

			ParamFunc := QueryAttr(attrs...)

			ParamFunc(q)

			setAttr := attrs[0]

			assertQueryStringEqual(setAttr.Name, setAttr.Value, q, t)

			expectedArgs := fmt.Sprintf("%v;%v", setAttr.Args.Operator[0].AsString(), setAttr.Args.Combinator[0].AsString())
			assertQueryStringEqual(setAttr.Name+QUERY_ARGS_SUFFIX, expectedArgs, q, t)
		},
	)

	t.Run(
		"Without args",
		func(t *testing.T) {
			q := url.Values{}

			attrs := []AttrQuery{
				{
					Name:  "myAttr",
					Value: "myValue",
				},
			}

			ParamFunc := QueryAttr(attrs...)

			ParamFunc(q)

			setAttr := attrs[0]
			assertQueryStringEqual(setAttr.Name, setAttr.Value, q, t)

			expectedArgs := ""
			assertQueryStringEqual(setAttr.Name+QUERY_ARGS_SUFFIX, expectedArgs, q, t)
		},
	)

}

func TestCanConvertPublitTimeToTime(t *testing.T) {
	t.Parallel()
	publitTimeStr := PublitTime("2017-07-10 17:05:00")

	tt, err := publitTimeStr.ConvertPublitTimeToTime()

	if err != nil {
		t.Error("Received an error but did not expect one.")
	}

	et := time.Date(2017, 7, 10, 17, 05, 0, 0, time.UTC)

	if !tt.Equal(et) {
		t.Error("Parsed time did not match expected")
	}
}

func TestConversionToPublitTimeFailsIfNotCorrectlyFormated(t *testing.T) {
	t.Parallel()
	faultyTimeString := PublitTime("2017-07-10-10 17:05:00")

	_, err := faultyTimeString.ConvertPublitTimeToTime()

	if err == nil {
		t.Error("Did not receive an error but was expecting one.")
	}
}

func TestCanParsePublitBoolToBool(t *testing.T) {
	t.Parallel()
	publitBools := map[PublitBool]bool{"True": true, "TRUE": true, "true": true, "False": false, "FALSE": false, "false": false}

	for pb, b := range publitBools {
		cb := pb.ConvertPublitBoolToBool()

		if cb != b {
			t.Error("Bool was not converted as expected.")
		}
	}
}

func TestCanGetErrorFromAPIErrorResponse(t *testing.T) {
	t.Parallel()

	e := &APIErrorResponse{
		Code: http.StatusBadRequest,
		Type: "BadRequest",
		Errors: []*APIError{
			{
				Info: "Some information",
				Type: "BadRequest",
			},
		},
		CombinedInfo: "Some combined information",
	}

	err := e.GetAsError()

	if err == nil {
		t.Error("No error returned but was expecting one.")
	}

	expectedError := fmt.Sprintf(`Code: "%v", Type: "%v", Combined info: "%v"`, e.Code, e.Type, e.CombinedInfo)

	if err.Error() != expectedError {
		t.Errorf(`Error string did not match expected. Got "%v", want "%v".`, err.Error(), expectedError)
	}
}

func TestAPIErrorResponseHasInformation(t *testing.T) {
	t.Parallel()

	t.Run(
		"Should return true if all fields are set",
		func(t *testing.T) {
			e := &APIErrorResponse{
				Code: http.StatusBadRequest,
				Type: "BadRequest",
				Errors: []*APIError{
					{
						Info: "Some information",
						Type: "BadRequest",
					},
				},
				CombinedInfo: "Some combined information",
			}

			if !e.HasInformation() {
				t.Error("Expected APIErrorResponse to have information, but got false.")
			}

		},
	)
	t.Run(
		"Should return false if all fields are not set",
		func(t *testing.T) {
			noInfoTable := []*APIErrorResponse{
				{ // No code
					Type:         "BadRequest",
					CombinedInfo: "some info",
				},
				{ // No Type
					Code:         http.StatusBadRequest,
					CombinedInfo: "some info",
				},
				{ // No combined info
					Code: http.StatusBadRequest,
					Type: "BadRequest",
				},
				{}, // Nothing set
			}

			for _, v := range noInfoTable {
				if v.HasInformation() {
					t.Error("Expected APIErrorResponse to not have information but got true.")
				}
			}
		},
	)
}

func assertQueryStringEqual(valueName, expected string, q url.Values, t *testing.T) {
	if q.Get(valueName) != expected {
		t.Errorf(`%v did not match expected. Got "%v", expected "%v"`, valueName, q.Get(valueName), expected)
	}
}

// Examples

func ExampleQueryAttr() {
	attrs := []AttrQuery{
		{
			Name:  "parameter",
			Value: "value",
			Args: AttrArgs{
				Operator:   []Operator{OPERATOR_EQUAL},
				Combinator: []Combinator{COMBINATOR_AND},
			},
		},
	}
	f := QueryAttr(attrs...)

	q := url.Values{}
	f(q)

	fmt.Printf("param filter: %s, args: %s\n", q.Get("parameter"), q.Get("parameter"+QUERY_ARGS_SUFFIX))
	// Output: param filter: value, args: EQUAL;AND
}

func ExampleQueryAuxiliary() {
	auxiliaryAttributes := []string{"aux1", "aux2"}
	f := QueryAuxiliary(auxiliaryAttributes...)

	q := url.Values{}
	f(q)

	fmt.Printf("Auxiliary attributes: %s\n", q.Get(QUERY_KEY_AUX))
	// Output: Auxiliary attributes: aux1,aux2
}

func ExampleQueryLimit() {
	f := QueryLimit(1, 0)

	q := url.Values{}
	f(q)

	fmt.Printf("Limit: %s\n", q.Get(QUERY_KEY_LIMIT))
	// Output: Limit: 0,1
}

func ExampleQueryOrderBy() {
	orderByParams := []string{"param1"}
	f := QueryOrderBy(orderByParams, ORDER_DIR_DESC)

	q := url.Values{}
	f(q)

	fmt.Printf("Order by: %s, Order dir: %s\n", q.Get(QUERY_KEY_ORDER), q.Get(QUERY_KEY_ORDER_DIR))
	// Output: Order by: param1, Order dir: DESC
}

func ExampleQueryScope() {
	scopes := []Scope{
		{
			Scope:  "scope",
			Filter: "filter",
		},
	}
	f := QueryScope(scopes)

	q := url.Values{}
	f(q)

	fmt.Printf("Scopes: %s\n", q.Get(QUERY_KEY_SCOPE))
	// Output: Scopes: scope;filter
}

func ExampleQueryWith() {
	withs := []string{"relation1", "relation2"}
	f := QueryWith(withs...)

	q := url.Values{}
	f(q)

	fmt.Printf("Withs: %s\n", q.Get(QUERY_KEY_WITH))
	// Output: Withs: relation1,relation2
}

func ExampleAPIErrorResponse_GetAsError() {
	APIErr := APIErrorResponse{
		Code:         http.StatusNotFound,
		Type:         "Not found",
		CombinedInfo: "Could not find something.",
		Errors: []*APIError{
			{
				Info: "Could not find something.",
				Type: "Not found",
			},
		},
	}

	err := APIErr.GetAsError()

	fmt.Printf("Error: %v\n", err.Error())
	// Output: Error: Code: "404", Type: "Not found", Combined info: "Could not find something."
}

func ExamplePublitTime_ConvertPublitTimeToTime() {
	var str PublitTime = "2017-07-17 10:26:00"

	t, _ := str.ConvertPublitTimeToTime()

	fmt.Printf("Year: %d, Month: %s, Day: %d\n", t.Year(), t.Month().String(), t.Day())
	// Output: Year: 2017, Month: July, Day: 17
}
