package application

import (
	"context"

	"github.com/SamPariat/roundup/services/event-service/internal/domain"
	"go.uber.org/zap"
)

// eventPhotoWidth fits a place card on the event screens, 400 points at 2x.
const eventPhotoWidth = 800

// PhotoURLs turns places' stored photo references into fresh URLs, by
// reference. A photo is only decoration, so when venue-service can't answer
// it logs and returns none: the event still shows, without photos.
func (u *EventUseCase) PhotoURLs(ctx context.Context, places []domain.Place) map[string]string {
	refs := make([]string, 0, len(places))
	seen := make(map[string]bool, len(places))

	for _, p := range places {
		if p.PhotoRef != "" && !seen[p.PhotoRef] {
			seen[p.PhotoRef] = true
			refs = append(refs, p.PhotoRef)
		}
	}

	urls, err := u.venues.PhotoURLs(ctx, refs, eventPhotoWidth)
	if err != nil {
		u.log.Warn("failed to look up place photos", zap.Int("photos", len(refs)), zap.Error(err))

		return map[string]string{}
	}

	return urls
}
