package middleware

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

// redacted replaces the value of every sensitive key in a logged body.
const redacted = "[REDACTED]"

// sensitiveKeys are JSON keys whose values must never reach the logs: the
// OTP code and the tokens that would let a log reader act as the user.
var sensitiveKeys = map[string]struct{}{
	"code":         {},
	"accessToken":  {},
	"refreshToken": {},
}

// redact returns b as compact JSON with the value of every sensitive key, at
// any depth, replaced by "[REDACTED]". ok is false if b is not a single valid
// JSON value.
func redact(b []byte) (out []byte, ok bool) {
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber() // keep large IDs exact instead of rounding through float64

	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, false
	}

	if _, err := dec.Token(); !errors.Is(err, io.EOF) {
		return nil, false // trailing data after the first value
	}

	var buf bytes.Buffer

	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)

	if err := enc.Encode(redactValue(v)); err != nil {
		return nil, false
	}

	return bytes.TrimSuffix(buf.Bytes(), []byte("\n")), true
}

// redactValue walks a decoded JSON value (map[string]any, []any, or a
// scalar) and returns it with sensitive keys' values replaced.
func redactValue(v any) any {
	switch t := v.(type) {
	case map[string]any:
		for k, child := range t {
			if _, ok := sensitiveKeys[k]; ok {
				t[k] = redacted // whatever its type, so a nested secret never leaks
			} else {
				t[k] = redactValue(child)
			}
		}

		return t
	case []any:
		for i, child := range t {
			t[i] = redactValue(child)
		}

		return t
	default:
		return v
	}
}
