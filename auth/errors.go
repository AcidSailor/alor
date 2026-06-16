// Package auth error types are the shared restkit types, re-exported as aliases
// so callers match them with errors.As without importing restkit:
//
//	var ce *auth.ConfigError
//	if errors.As(err, &ce) { _ = ce.Reason } // New misuse: empty host/refreshToken
//
//	var re *auth.ResponseError
//	if errors.As(err, &re) { _ = re.StatusCode } // non-2xx /refresh, e.g. 403
//
//	var re *auth.RequestError
//	if errors.As(err, &re) { _ = re.Op }         // "send", "unmarshal", "build", ...
//
// No sentinel errors: the typed error IS the category. A non-2xx /refresh is a
// *ResponseError; every other Token() failure is a *RequestError whose Op names
// the stage (transport "send", a 2xx body that is not a usable token
// "unmarshal"). Each wraps its cause via %w, so a transport *url.Error or a
// JWT-decode error survives an unwrap.
package auth

import "github.com/acidsailor/restkit"

// ConfigError is invalid [New] construction input (an empty refreshToken or an
// empty oauthHost).
type ConfigError = restkit.ConfigError

// ResponseError is a non-2xx response from Alor's /refresh endpoint, carrying
// the status code and raw body verbatim so callers can branch on, e.g., a
// 401/403 (bad or expired refresh token). It is the typed result of the call,
// not a RequestError.
type ResponseError = restkit.ResponseError

// RequestError is any other Token() failure (request body encode, request
// build, transport, body read, or a 2xx body that did not yield a usable
// token). Op names the stage and Err wraps the cause.
type RequestError = restkit.RequestError

// RequestError.Op values Token() can emit, re-exported so callers match the
// failed stage without importing restkit. auth installs no request hooks, so
// never produces OpHook.
const (
	OpMarshal   = restkit.OpMarshal   // encoding the /refresh body to JSON
	OpBuild     = restkit.OpBuild     // constructing the *http.Request
	OpSend      = restkit.OpSend      // the HTTP round-trip
	OpRead      = restkit.OpRead      // reading the response body
	OpUnmarshal = restkit.OpUnmarshal // a 2xx body that did not yield a usable token
)
