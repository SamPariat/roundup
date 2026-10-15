package phone

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsValid(t *testing.T) {
	tests := []struct {
		name  string
		phone string
		want  bool
	}{
		{name: "India", phone: "+919876543210", want: true},
		{name: "ShortestAllowed", phone: "+12345678", want: true},
		{name: "LongestAllowed", phone: "+123456789012345", want: true},
		{name: "MissingPlus", phone: "919876543210", want: false},
		{name: "LeadingZeroCountryCode", phone: "+0123456789", want: false},
		{name: "TooShort", phone: "+1234567", want: false},
		{name: "TooLong", phone: "+1234567890123456", want: false},
		{name: "Spaces", phone: "+91 98765 43210", want: false},
		{name: "Empty", phone: "", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, IsValid(tt.phone))
		})
	}
}
