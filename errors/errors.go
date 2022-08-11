package errors

import (
	"fmt"
)

type GraphQLError interface {
	PrepareExtErr() *QueryError
}

type QueryError struct {
	Err           error                  `json:"-"` // Err holds underlying if available
	Message       string                 `json:"message"`
	Locations     []Location             `json:"locations,omitempty"`
	Path          []interface{}          `json:"path,omitempty"`
	Rule          string                 `json:"-"`
	ResolverError error                  `json:"-"`
	Extensions    map[string]interface{} `json:"extensions,omitempty"`
}

type Location struct {
	Line   int `json:"line"`
	Column int `json:"column"`
}

func (a Location) Before(b Location) bool {
	return a.Line < b.Line || (a.Line == b.Line && a.Column < b.Column)
}

func Errorf(format string, a ...interface{}) *QueryError {
	// similar to fmt.Errorf, Errorf will wrap the last argument if it is an instance of error
	var err error
	if n := len(a); n > 0 {
		if v, ok := a[n-1].(error); ok {
			err = v
		}
	}

	return &QueryError{
		Err:     err,
		Message: fmt.Sprintf(format, a...),
	}
}

func (err *QueryError) Error() string {
	if err == nil {
		return "<nil>"
	}
	str := fmt.Sprintf("graphql: %s", err.Message)
	for _, loc := range err.Locations {
		str += fmt.Sprintf(" (line %d, column %d)", loc.Line, loc.Column)
	}
	return str
}

func (err *QueryError) Unwrap() error {
	if err == nil {
		return nil
	}
	return err.Err
}

var _ error = &QueryError{}

func (err *QueryError) AddErrCode(code int) *QueryError {
	err.Extensions["Code"] = code
	return err
}

func (err *QueryError) AddDevMsg(msg string) *QueryError {
	err.Extensions["DeveloperMessage"] = msg
	return err
}

func (err *QueryError) AddMoreInfo(moreInfo string) *QueryError {
	err.Extensions["MoreInfo"] = moreInfo
	return err
}

func (err *QueryError) AddErrTimestamp(errTime string) *QueryError {
	err.Extensions["Timestamp"] = errTime
	return err
}
