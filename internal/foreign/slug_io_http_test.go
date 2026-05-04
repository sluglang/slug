package foreign

import (
	"errors"
	"fmt"
	"net/http"
	"slug/internal/dec64"
	"slug/internal/object"
	"slug/internal/util"
	"testing"
	"time"
)

type testRuntimeContext struct{}

func (testRuntimeContext) CurrentEnv() *object.Environment { return nil }
func (testRuntimeContext) ApplyFunction(pos int, fnName string, fnObj object.Object, positional []object.Object, named map[string]object.Object) object.Object {
	return &object.Error{Message: "unused"}
}
func (testRuntimeContext) NewError(message string, a ...interface{}) *object.Error {
	return &object.Error{Message: fmt.Sprintf(message, a...)}
}
func (testRuntimeContext) Nil() *object.Nil { return object.NIL }
func (testRuntimeContext) NativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return object.TRUE
	}
	return object.FALSE
}
func (testRuntimeContext) LoadModule(pathParts string) (*object.Module, error) {
	return nil, fmt.Errorf("unused")
}
func (testRuntimeContext) GetConfiguration() util.Configuration {
	return util.Configuration{Version: "test"}
}
func (testRuntimeContext) NextHandleID() int64 { return 0 }

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}

func TestIoHttpRequestRequestErrorDoesNotPanic(t *testing.T) {
	origNewHTTPClient := newHTTPClient
	t.Cleanup(func() { newHTTPClient = origNewHTTPClient })
	newHTTPClient = func(timeout time.Duration) *http.Client {
		return &http.Client{
			Timeout: timeout,
			Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
				return nil, errors.New("dial failed")
			}),
		}
	}

	f := fnIoHttpRequest()
	ctx := testRuntimeContext{}
	headers := (&object.Map{}).Put(&object.String{Value: "x-test"}, &object.String{Value: "1"})

	result := f.Fn(ctx,
		&object.String{Value: "GET"},
		&object.String{Value: "http://example.invalid"},
		&object.String{Value: ""},
		headers,
		&object.Number{Value: dec64.FromInt(1000)},
	)
	if result.Type() != object.ERROR_OBJ {
		t.Fatalf("expected error result, got %s (%s)", result.Type(), result.Inspect())
	}
}

func TestIoHttpRequestHasTimeout(t *testing.T) {
	const timeoutMs = int64(100)
	origNewHTTPClient := newHTTPClient
	t.Cleanup(func() { newHTTPClient = origNewHTTPClient })
	newHTTPClient = func(timeout time.Duration) *http.Client {
		return &http.Client{
			Timeout: timeout,
			Transport: roundTripperFunc(func(r *http.Request) (*http.Response, error) {
				<-r.Context().Done()
				return nil, r.Context().Err()
			}),
		}
	}

	f := fnIoHttpRequest()
	ctx := testRuntimeContext{}
	start := time.Now()
	result := f.Fn(ctx,
		&object.String{Value: "GET"},
		&object.String{Value: "http://example.invalid"},
		&object.String{Value: ""},
		&object.Map{},
		&object.Number{Value: dec64.FromInt64(timeoutMs)},
	)
	elapsed := time.Since(start)

	if result.Type() != object.ERROR_OBJ {
		t.Fatalf("expected timeout error, got %s (%s)", result.Type(), result.Inspect())
	}
	if elapsed > 2*time.Second {
		t.Fatalf("request did not timeout in bounded time, elapsed=%s", elapsed)
	}
}
