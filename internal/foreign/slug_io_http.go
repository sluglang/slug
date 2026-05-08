package foreign

import (
	"io"
	"net/http"
	"slug/internal/dec64"
	"slug/internal/object"
	"strings"
	"time"
)

var newHTTPClient = func(timeout time.Duration) *http.Client {
	return &http.Client{Timeout: timeout}
}

func fnIoHttpRequest() *object.Foreign {
	return &object.Foreign{
		Name: "request",
		Fn: func(ctx object.RuntimeContext, args ...object.Object) object.Object {

			if len(args) != 5 {
				return ctx.NewError("wrong number of arguments to `request`, got=%d, want=5", len(args))
			}

			method, err := unpackString(args[0], "method")
			if err != nil {
				return ctx.NewError(err.Error())
			}

			url, err := unpackString(args[1], "url")
			if err != nil {
				return ctx.NewError(err.Error())
			}

			body, err := unpackString(args[2], "body")
			if err != nil {
				return ctx.NewError(err.Error())
			}

			mapObj, ok := args[3].(*object.Map)
			if !ok {
				return ctx.NewError("argument to `headers` must be a MAP, got=%s", args[3].Type())
			}

			timeoutMs, err := unpackNumber(args[4], "timeout")
			if err != nil {
				return ctx.NewError(err.Error())
			}
			if timeoutMs <= 0 {
				return ctx.NewError("argument to `timeout` must be greater than 0, got=%d", timeoutMs)
			}

			client := newHTTPClient(time.Duration(timeoutMs) * time.Millisecond)
			req, err := http.NewRequest(method, url, strings.NewReader(body))
			if err != nil {
				return ctx.NewError(err.Error())
			}

			req.Header.Set("user-agent", "Slug/"+ctx.GetConfiguration().Version)
			mapObj.ForEach(func(_ object.MapKey, v object.MapPair) bool {
				req.Header.Set(v.Key.Inspect(), v.Value.Inspect())
				return true
			})

			resp, err := client.Do(req)
			if err != nil {
				return ctx.NewError(err.Error())
			}
			if resp == nil || resp.Body == nil {
				return ctx.NewError("http request failed: empty response")
			}
			defer resp.Body.Close()

			bytes, err := io.ReadAll(resp.Body)
			if err != nil {
				return ctx.NewError(err.Error())
			}

			return &object.List{
				Elements: []object.Object{
					&object.Number{Value: dec64.FromInt64(int64(resp.StatusCode))},
					&object.String{Value: string(bytes)},
				},
			}
		},
	}
}
