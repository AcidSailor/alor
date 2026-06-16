// Package alor error types are the shared restkit types, re-exported as aliases
// so callers match them with errors.As without importing restkit:
//
//	var re *alor.ResponseError
//	if errors.As(err, &re) { _ = re.StatusCode } // non-2xx, e.g. 400
//
//	var re *alor.RequestError
//	if errors.As(err, &re) { _ = re.Op }          // "hook", "send", "unmarshal", ...
//
// No sentinel errors: the typed error IS the category. Each wraps its cause via
// %w, so the cause survives an unwrap — a bearer-auth-hook *RequestError still
// reaches the auth package's restkit *ResponseError/*RequestError via errors.As,
// a transport failure still reaches a *url.Error or
// context.Canceled/DeadlineExceeded, and so on.
package alor

import "github.com/acidsailor/restkit"

// ResponseError is a non-2xx Alor API response. It carries the status code and
// the raw body verbatim (Alor's error envelope is not decoded). It is the typed
// result of the call, not a RequestError.
type ResponseError = restkit.ResponseError

// RequestError is a per-call failure before a response can be mapped (token
// refresh, body encode, request build, transport, body read, or 2xx decode). Op
// names the stage and Err wraps the cause.
type RequestError = restkit.RequestError

// ConfigError is invalid NewClient construction input (a nil TokenSource or an
// empty endpoint).
type ConfigError = restkit.ConfigError

// RequestError.Op values, re-exported so callers can match the failed stage
// without importing restkit.
const (
	OpMarshal   = restkit.OpMarshal   // encoding the request body to JSON
	OpBuild     = restkit.OpBuild     // constructing the *http.Request
	OpHook      = restkit.OpHook      // a request hook returned an error
	OpSend      = restkit.OpSend      // the HTTP round-trip
	OpRead      = restkit.OpRead      // reading the response body
	OpUnmarshal = restkit.OpUnmarshal // decoding the 2xx body
)
