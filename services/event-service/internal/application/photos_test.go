package application

import (
	"testing"

	"github.com/SamPariat/roundup/services/event-service/internal/domain"
	"github.com/SamPariat/roundup/shared/domainerr"
	"github.com/stretchr/testify/assert"
)

func TestPhotoURLsLooksUpEachStoredRefOnce(t *testing.T) {
	venues := &fakeVenues{}
	uc := newUseCase(t, &fakeRepo{}, venues)

	urls := uc.PhotoURLs(t.Context(), []domain.Place{{PhotoRef: "a"}, {}, {PhotoRef: "b"}, {PhotoRef: "a"}})

	assert.Equal(t, map[string]string{"a": "https://img.example/a?w=800", "b": "https://img.example/b?w=800"}, urls)
	assert.Equal(t, [][]string{{"a", "b"}}, venues.photoRefs, "one call, no empty or repeated refs")
}

func TestPhotoURLsShowsTheEventWithoutPhotosWhenVenueServiceIsDown(t *testing.T) {
	venues := &fakeVenues{photoErr: domainerr.New(domain.ErrVenueServiceUnavailable, "down")}
	uc := newUseCase(t, &fakeRepo{}, venues)

	assert.Empty(t, uc.PhotoURLs(t.Context(), []domain.Place{{PhotoRef: "a"}}))
}
