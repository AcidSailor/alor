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
