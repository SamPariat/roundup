package pagination

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromQuery(t *testing.T) {
	tests := []struct {
		name      string
		limit     string
		wantLimit int
		wantErr   bool
	}{
		{name: "Default", limit: "", wantLimit: DefaultLimit},
		{name: "Smallest", limit: "1", wantLimit: 1},
		{name: "Largest", limit: "100", wantLimit: MaxLimit},
		{name: "Zero", limit: "0", wantErr: true},
		{name: "Negative", limit: "-5", wantErr: true},
		{name: "TooLarge", limit: "101", wantErr: true},
		{name: "NotANumber", limit: "ten", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			page, err := FromQuery(tt.limit, "abc")
			if tt.wantErr {
				require.ErrorIs(t, err, ErrInvalid)

				return
			}

			require.NoError(t, err)
			assert.Equal(t, Page{Limit: tt.wantLimit, Cursor: "abc"}, page)
		})
	}
}

func TestCursorRoundTrip(t *testing.T) {
	type key struct {
		At time.Time `json:"at"`
		ID string    `json:"id"`
	}

	want := key{At: time.Date(2026, 10, 25, 20, 0, 0, 123, time.UTC), ID: "s-1"}

	var got key
	require.NoError(t, Decode(Encode(want), &got))

	assert.True(t, want.At.Equal(got.At), "nanoseconds survive, so keyset comparisons stay exact")
	assert.Equal(t, want.ID, got.ID)
}

func TestDecodeRejectsGarbage(t *testing.T) {
	var key struct{ ID string }

	for _, cursor := range []string{"%%%", "bm90IGpzb24", "!"} {
		assert.ErrorIs(t, Decode(cursor, &key), ErrInvalid, cursor)
	}
}

func TestTrim(t *testing.T) {
	page, more := Trim([]int{1, 2, 3}, 2)
	assert.Equal(t, []int{1, 2}, page)
	assert.True(t, more)

	page, more = Trim([]int{1, 2}, 2)
	assert.Equal(t, []int{1, 2}, page)
	assert.False(t, more)
}

func TestNextCursor(t *testing.T) {
	assert.Nil(t, NextCursor(""), "last page is null")
	assert.Equal(t, "abc", *NextCursor("abc"))
}
