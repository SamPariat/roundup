package application

import (
	"context"

	"github.com/SamPariat/roundup/services/event-service/internal/domain"
	"go.uber.org/zap"
)

// PhotoURLs turns places' stored photo references into fresh URLs at most
// width pixels wide, by reference. A photo is only decoration, so when venue-service can't answer
// it logs and returns none: the event still shows, without photos.
func (u *EventUseCase) PhotoURLs(ctx context.Context, places []domain.Place, width int) map[string]string {
	refs := make([]string, 0, len(places))
	seen := make(map[string]bool, len(places))

	for _, p := range places {
		if p.PhotoRef != "" && !seen[p.PhotoRef] {
			seen[p.PhotoRef] = true
			refs = append(refs, p.PhotoRef)
		}
	}

	urls, err := u.venues.PhotoURLs(ctx, refs, width)
	if err != nil {
		u.log.Warn("failed to look up place photos", zap.Int("photos", len(refs)), zap.Error(err))

		return map[string]string{}
	}

	return urls
}
