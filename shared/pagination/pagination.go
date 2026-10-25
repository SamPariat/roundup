// Package pagination is the cursor paging shared by every list endpoint.
// Clients send ?limit=&cursor= and get back the page plus nextCursor, which
// they send as cursor for the next page until it is null. Cursors are opaque:
// base64url-encoded JSON of the last row's sort key, chosen by each repository.
package pagination

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"

	"github.com/SamPariat/roundup/shared/domainerr"
)

// Page sizes.
const (
	// DefaultLimit is the page size when a request does not ask for one.
	DefaultLimit = 25

	// MaxLimit is the largest page a request may ask for.
	MaxLimit = 100
)

// ErrInvalid is returned for a limit outside 1–MaxLimit or a cursor this
// service did not issue. Handlers map it to 400.
var ErrInvalid = errors.New("invalid pagination")

// Page is one page request: how many rows, and where to start. An empty
// Cursor means the first page.
type Page struct {
	// Limit is the page size, 1–MaxLimit.
	Limit int

	// Cursor is the opaque nextCursor of the previous page, or "".
	Cursor string
}

// FromQuery builds a Page from the raw limit and cursor query parameters.
// An empty limit means DefaultLimit. The cursor is checked when decoded.
func FromQuery(limit, cursor string) (Page, error) {
	page := Page{Limit: DefaultLimit, Cursor: cursor}
	if limit == "" {
		return page, nil
	}

	n, err := strconv.Atoi(limit)
	if err != nil || n < 1 || n > MaxLimit {
		return Page{}, domainerr.New(ErrInvalid, fmt.Sprintf("limit must be between 1 and %d", MaxLimit))
	}

	page.Limit = n

	return page, nil
}

// Encode turns a sort key into an opaque cursor.
func Encode(key any) string {
	raw, err := json.Marshal(key)
	if err != nil {
		// Sort keys are plain structs of strings and times; this cannot fail.
		panic(fmt.Sprintf("pagination: encode cursor: %v", err))
	}

	return base64.RawURLEncoding.EncodeToString(raw)
}

// Decode reads a cursor made by Encode into key. Returns ErrInvalid for
// anything else.
func Decode(cursor string, key any) error {
	raw, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return domainerr.New(ErrInvalid, "malformed cursor")
	}

	if err := json.Unmarshal(raw, key); err != nil {
		return domainerr.New(ErrInvalid, "malformed cursor")
	}

	return nil
}

// Trim cuts rows, fetched with Limit+1, to one page and reports whether more
// rows follow. Repositories fetch one extra row so they know without a count.
func Trim[T any](rows []T, limit int) (page []T, more bool) {
	if len(rows) > limit {
		return rows[:limit], true
	}

	return rows, false
}
