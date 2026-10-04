package runtimeharpoon_test

import (
	"encoding/json"
	"net/url"
	"testing"

	"github.com/openai/tunnel-client/pkg/runtimeharpoon"
	"github.com/stretchr/testify/require"
)

func TestTransformJSONBodyPreservesNumbers(t *testing.T) {
	t.Parallel()
	target, err := url.Parse("https://example.test/api")
	require.NoError(t, err)
	rewriter := runtimeharpoon.NewURLRewriter([]runtimeharpoon.Target{{Label: "api", BaseURL: target}})
	for _, number := range []string{
		"9007199254740993", "18446744073709551615", "-9223372036854775808",
		"0.12345678901234567890123456789", "1e400", "1e-400", "-0", "1.000", "42",
	} {
		t.Run(number, func(t *testing.T) {
			t.Parallel()
			body := []byte(`{"url":"https://example.test/api","number":` + number + `,"nested":[` + number + `,{"value":` + number + `}]}`)
			updated, changed := runtimeharpoon.TransformJSONBody(body, rewriter)
			require.True(t, changed)
			var result map[string]json.RawMessage
			require.NoError(t, json.Unmarshal(updated, &result))
			require.Equal(t, `"harpoon://api"`, string(result["url"]))
			require.Equal(t, number, string(result["number"]))
			require.Equal(t, `[`+number+`,{"value":`+number+`}]`, string(result["nested"]))
		})
	}
}

func TestTransformJSONBodyPreservesUnchangedAndInvalidNumbers(t *testing.T) {
	t.Parallel()
	target, err := url.Parse("https://example.test/api")
	require.NoError(t, err)
	rewriter := runtimeharpoon.NewURLRewriter([]runtimeharpoon.Target{{Label: "api", BaseURL: target}})
	for _, body := range []string{
		` { "number": 9007199254740993, "url": "https://other.test/api" } `,
		`{"url":"https://example.test/api","number":01}`,
		`{"url":"https://example.test/api"} {"number":1}`,
	} {
		updated, changed := runtimeharpoon.TransformJSONBody([]byte(body), rewriter)
		require.False(t, changed)
		require.Equal(t, body, string(updated))
	}
}
