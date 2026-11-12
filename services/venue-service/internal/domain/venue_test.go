package domain

import (
	"errors"
	"math"
	"testing"

	"github.com/SamPariat/roundup/shared/api"
)

func TestSearchParamsValidate(t *testing.T) {
	valid := SearchParams{Latitude: 12.97, Longitude: 77.59, RadiusInMeters: 2000}

	tests := []struct {
		name    string
		mutate  func(p *SearchParams)
		wantErr bool
	}{
		{"Valid", func(_ *SearchParams) {}, false},
		{"BoundaryValues", func(p *SearchParams) { p.Latitude, p.Longitude, p.RadiusInMeters = -90, 180, 50000 }, false},
		{"LatitudeTooLow", func(p *SearchParams) { p.Latitude = -90.01 }, true},
		{"LatitudeTooHigh", func(p *SearchParams) { p.Latitude = 90.01 }, true},
		{"LongitudeTooLow", func(p *SearchParams) { p.Longitude = -180.01 }, true},
		{"LongitudeTooHigh", func(p *SearchParams) { p.Longitude = 180.01 }, true},
		{"LatitudeNaN", func(p *SearchParams) { p.Latitude = math.NaN() }, true},
		{"LongitudeNaN", func(p *SearchParams) { p.Longitude = math.NaN() }, true},
		{"LatitudeInf", func(p *SearchParams) { p.Latitude = math.Inf(1) }, true},
		{"RadiusMissing", func(p *SearchParams) { p.RadiusInMeters = 0 }, true},
		{"RadiusNegative", func(p *SearchParams) { p.RadiusInMeters = -1 }, true},
		{"RadiusTooLarge", func(p *SearchParams) { p.RadiusInMeters = 50001 }, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := valid
			tt.mutate(&p)

			err := p.Validate()
			if tt.wantErr != (err != nil) {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && !errors.Is(err, ErrInvalidSearch) {
				t.Fatalf("Validate() error = %v, want ErrInvalidSearch", err)
			}
		})
	}
}

func TestNewPhotoOptionsFillsWhatTheClientLeftOut(t *testing.T) {
	if got, want := NewPhotoOptions(0, 0), (PhotoOptions{Width: api.DefaultPhotoWidth, Count: api.MaxVenuePhotos}); got != want {
		t.Fatalf("NewPhotoOptions(0, 0) = %+v, want %+v", got, want)
	}

	if got, want := NewPhotoOptions(400, 2), (PhotoOptions{Width: 400, Count: 2}); got != want {
		t.Fatalf("NewPhotoOptions(400, 2) = %+v, want %+v", got, want)
	}
}

func TestPhotoOptionsValidate(t *testing.T) {
	tests := []struct {
		name    string
		options PhotoOptions
		wantErr bool
	}{
		{"SmallestWidth", PhotoOptions{Width: 200, Count: 1}, false},
		{"LargestWidthAndCount", PhotoOptions{Width: 1600, Count: api.MaxVenuePhotos}, false},
		{"WidthOffTheList", PhotoOptions{Width: 700, Count: 1}, true},
		{"WidthTooLarge", PhotoOptions{Width: 4800, Count: 1}, true},
		{"NegativeWidth", PhotoOptions{Width: -800, Count: 1}, true},
		{"TooManyPhotos", PhotoOptions{Width: 800, Count: api.MaxVenuePhotos + 1}, true},
		{"NegativeCount", PhotoOptions{Width: 800, Count: -1}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.options.Validate()
			if tt.wantErr != (err != nil) {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantErr && !errors.Is(err, ErrInvalidPhotoOptions) {
				t.Fatalf("Validate() error = %v, want ErrInvalidPhotoOptions", err)
			}
		})
	}
}
