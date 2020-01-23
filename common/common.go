// Copyright 2017 Publit Sweden AB. All rights reserved.

// Common methods and helpers for the Publit APIs.
package common

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"time"
)

// General query string constants.
// These strings correlate to reserved query string keys that Publit implements in the general API interface.
const (
	QUERY_KEY_LIMIT     = "limit"
	QUERY_KEY_WITH      = "with"
	QUERY_KEY_SCOPE     = "scope"
	QUERY_KEY_AUX       = "auxiliary"
	QUERY_KEY_ORDER     = "order_by"
	QUERY_KEY_ORDER_DIR = "order_dir"
	QUERY_ARGS_SUFFIX   = "_args"
	QUERY_KEY_GROUP_BY  = "group_by"
)

// Operator describes the different operators implemented in Publits general API interface.
type Operator int

// Operator enum constants.
const (
	OPERATOR_EQUAL Operator = 1 + iota
	OPERATOR_NOT_EQUAL
	OPERATOR_GREATER_EQUAL
	OPERATOR_GREATER
	OPERATOR_LESS_EQUAL
	OPERATOR_LESS
)

// Combinator describes the different combinators implemented in Publits general API interface..
type Combinator int

// Combinator enum constants.
const (
	COMBINATOR_AND Combinator = 1 + iota
	COMBINATOR_OR
)

// PublitTime type. Regular string which can be interpreted to time.
type PublitTime string

// PublitBool type. Regular string which can be interpreted to bool.
type PublitBool string

// Operator strings.
var operators = []string{
	"EQUAL",
	"NOT_EQUAL",
	"GREATER_EQUAL",
	"GREATER",
	"LESS_EQUAL",
	"LESS",
}

// Combinator string
var combinators = []string{
	"AND",
	"OR",
}

// OrderDir type.
type OrderDir int

// OrderDir enum constants.
const (
	ORDER_DIR_ASC OrderDir = 1 + iota
	ORDER_DIR_DESC
)

// OrderDir string.
var orderDirections = []string{
	"ASC",
	"DESC",
}

// General Publit API error response.
// Most errors received from the Publit APIs conform to this standard.
type APIErrorResponse struct {
	Code         int         `json:"Code"`
	Type         string      `json:"Type"`
	Errors       []*APIError `json:"errors"`
	CombinedInfo string      `json:"CombinedInfo"`
}

// General Publit API error.
type APIError struct {
	Info string `json:"Info"`
	Type string `json:"Type"`
}

// Returns APIErrorResponse as error.
func (e *APIErrorResponse) GetAsError() error {
	return errors.New(
		fmt.Sprintf(`Code: "%v", Type: "%v", Combined info: "%v"`, e.Code, e.Type, e.CombinedInfo),
	)
}

// Checks if APIErrorResponse is set.
func (e *APIErrorResponse) HasInformation() bool {
	if e.Code != 0 && e.Type != "" && e.CombinedInfo != "" {
		return true
	}
	return false
}

// Helper to set limit parameter to API query.
// Functions with signature func(q url.Values) are implemented in the more specific SDKs of the PublitGoSDK packages.
func QueryLimit(limit, offset int) func(q url.Values) {
	limitString := fmt.Sprintf("%v,%v", offset, limit)
	return func(q url.Values) {
		q.Add(QUERY_KEY_LIMIT, limitString)
	}
}

// Helper to set With parameter to API query.
// Functions with signature func(q url.Values) are implemented in the more specific SDKs of the PublitGoSDK packages.
func QueryWith(withs ...string) func(q url.Values) {
	withString := strings.Join(withs, ",")

	return func(q url.Values) {
		q.Add(QUERY_KEY_WITH, withString)
	}
}

// Scope struct for setting a scope.
type Scope struct {
	Scope  string
	Filter string
}

// Helper to sets scope parameter to API query.
// Functions with signature func(q url.Values) are implemented in the more specific SDKs of the PublitGoSDK packages.
func QueryScope(scopes []Scope) func(q url.Values) {
	var scopeStrings []string
	for _, v := range scopes {
		str := ""
		if v.Filter != "" {
			str = fmt.Sprintf("%v;%v", v.Scope, v.Filter)
		} else {
			str = v.Scope
		}
		scopeStrings = append(scopeStrings, str)
	}

	scopeString := strings.Join(scopeStrings, ",")

	return func(q url.Values) {
		q.Add(QUERY_KEY_SCOPE, scopeString)
	}
}

// Helper to sets auxiliary attributes parameter to API query.
// Functions with signature func(q url.Values) are implemented in the more specific SDKs of the PublitGoSDK packages.
func QueryAuxiliary(auxiliaryAttributes ...string) func(q url.Values) {
	auxString := strings.Join(auxiliaryAttributes, ",")

	return func(q url.Values) {
		q.Add(QUERY_KEY_AUX, auxString)
	}
}

// Helper to sets OrderBy and OrderDir parameter to API query.
// Functions with signature func(q url.Values) are implemented in the more specific SDKs of the PublitGoSDK packages.
func QueryOrderBy(attributes []string, dir OrderDir) func(q url.Values) {
	orderString := strings.Join(attributes, ",")

	return func(q url.Values) {
		q.Add(QUERY_KEY_ORDER, orderString)

		if dir != 0 {
			q.Add(QUERY_KEY_ORDER_DIR, dir.AsString())
		}
	}
}

// QueryGroupBy sets group by query to API query.
func QueryGroupBy(attributes []string) func(q url.Values) {
	groupByString := strings.Join(attributes, ",")

	return func(q url.Values) {
		q.Add(QUERY_KEY_GROUP_BY, groupByString)
	}
}

// Attribute query string struct.
type AttrQuery struct {
	Name  string
	Value string
	Args  AttrArgs
}

// Attribute args struct.
type AttrArgs struct {
	Operator   []Operator
	Combinator []Combinator
}

// Helper to set attribute filters to API query.
// Functions with signature func(q url.Values) are implemented in the more specific SDKs of the PublitGoSDK packages.
func QueryAttr(attributes ...AttrQuery) func(q url.Values) {
	return func(q url.Values) {
		for _, v := range attributes {
			q.Add(v.Name, v.Value)
			if !v.Args.IsEmpty() {
				argsAttr := fmt.Sprintf("%v%v", v.Name, QUERY_ARGS_SUFFIX)

				opCombStr := []string{}
				for i, op := range v.Args.Operator {
					argsstr := op.AsString()
					if len(v.Args.Combinator) > i {
						argsstr = fmt.Sprintf("%v;%v", op.AsString(), v.Args.Combinator[i].AsString())
					}
					opCombStr = append(opCombStr, argsstr)
				}

				q.Add(argsAttr, strings.Join(opCombStr[:], ","))
			}
		}
	}
}

func (a AttrArgs) IsEmpty() bool {
	return reflect.DeepEqual(a, AttrArgs{})
}

// Converts Publit style times to Time.
// Use for converting timestamps in responses from the Publit APIs to Go's time.
func (timeString PublitTime) ConvertPublitTimeToTime() (time.Time, error) {
	t := time.Time{}
	if timeString != "" {
		t, err := time.Parse("2006-01-02 15:04:05", string(timeString))

		return t, err
	}

	return t, nil
}

// Converts publit string bool representations to actual bool.
// Use for converting Publit style boolean enums (strings) to actual bool values.
func (str PublitBool) ConvertPublitBoolToBool() bool {
	if strings.ToLower(string(str)) == "false" {
		return false
	}
	return true
}

// Returns Operator "enum" as string.
func (o Operator) AsString() string {
	return operators[o-1]
}

// Returns Combinator "enum" as string.
// This string is used for assembling query string parameters to the Publit APIs.
func (c Combinator) AsString() string {
	return combinators[c-1]
}

// Returns OrderDir "enum" as string.
// This string is used for assembling query string parameters to the Publit APIs.
func (o OrderDir) AsString() string {
	return orderDirections[o-1]
}
