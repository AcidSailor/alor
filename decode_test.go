package alor

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

// Alor returns `message` as a bare JSON string on a successful order action;
// the generated type must decode it without error (regression for the
// allOf-string-alias -> struct-embedding-string codegen defect).
func TestResponseOrderActionDecodesStringMessage(t *testing.T) {
	const body = `{"message":"success","orderNumber":"18995978560"}`

	var commandAPI ResponseOrderActionLimitMarketCommandAPI
	require.NoError(t, json.Unmarshal([]byte(body), &commandAPI))

	var limitMarket ResponseOrderActionLimitMarket
	require.NoError(t, json.Unmarshal([]byte(body), &limitMarket))
}

// Same bare-JSON-string `message` and allOf-string-alias defect as the order
// actions, here on a successful order-group creation.
func TestResponseOrderGroupCreationDecodesStringMessage(t *testing.T) {
	const body = `{"message":"success"}`

	var success ResponseOrderGroupCreationSuccess
	require.NoError(t, json.Unmarshal([]byte(body), &success))
}

// Alor returns `status` / `statusCode` as a bare JSON number; the generated
// types must decode it (regression for the allOf-int32-alias -> struct defect).
func TestResponseDecodesNumericHTTPCode(t *testing.T) {
	var groupErr ResponseOrderGroupCreationError
	require.NoError(t, json.Unmarshal([]byte(`{"status":400}`), &groupErr))
	require.EqualValues(t, 400, *groupErr.Status)

	var action ResponseOrderActionCode400CommandAPI
	require.NoError(t, json.Unmarshal(
		[]byte(`{"oldResponse":{"statusCode":400}}`), &action,
	))
	require.EqualValues(t, 400, *action.OldResponse.StatusCode)
}
