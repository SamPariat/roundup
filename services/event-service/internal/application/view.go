package application

import "github.com/SamPariat/roundup/services/event-service/internal/domain"

// summarise builds the list entries for events as seen by userID. Only the
// current members' answers count; a member with no answer is pending.
func summarise(events []domain.Event, members []domain.SquadMember, userID string,
	rsvps []domain.RSVP, votes []domain.Vote, nominationCounts map[string]int,
) []domain.EventSummary {
	answers := make(map[string]map[string]domain.RSVPStatus, len(events))
	for _, r := range rsvps {
		if answers[r.EventID] == nil {
			answers[r.EventID] = make(map[string]domain.RSVPStatus)
		}

		answers[r.EventID][r.UserID] = r.Status
	}

	voted := make(map[string]bool, len(events))

	for _, v := range votes {
		if v.UserID == userID {
			voted[v.EventID] = true
		}
	}

	summaries := make([]domain.EventSummary, 0, len(events))
	for _, e := range events {
		summary := domain.EventSummary{
			Event:           e,
			GoingPreview:    []domain.SquadMember{},
			NominationCount: nominationCounts[e.ID],
			MyRSVP:          answers[e.ID][userID],
			HasVoted:        voted[e.ID],
		}

		for _, m := range members {
			switch answers[e.ID][m.UserID] {
			case domain.RSVPGoing:
				summary.RSVPCounts.Going++
				if len(summary.GoingPreview) < domain.GoingPreviewSize {
					summary.GoingPreview = append(summary.GoingPreview, m)
				}
			case domain.RSVPMaybe:
				summary.RSVPCounts.Maybe++
			case domain.RSVPNotGoing:
				summary.RSVPCounts.NotGoing++
			default:
				summary.RSVPCounts.Pending++
			}
		}

		summaries = append(summaries, summary)
	}

	return summaries
}

// detail builds the event screen as seen by userID, including which actions
// the policy and the event's status allow them right now.
func detail(e *domain.Event, members []domain.SquadMember, userID string, rsvps []domain.RSVP,
	votes []domain.Vote, nominations []domain.Nomination, authorizer domain.Authorizer,
) *domain.EventDetail {
	byID := make(map[string]domain.SquadMember, len(members))
	for _, m := range members {
		byID[m.UserID] = m
	}

	answers := make(map[string]domain.RSVPStatus, len(rsvps))
	for _, r := range rsvps {
		answers[r.UserID] = r.Status
	}

	d := &domain.EventDetail{
		Event:         *e,
		CreatedByName: byID[e.CreatedBy].Name,
		Members:       make([]domain.MemberRSVP, 0, len(members)),
		Nominations:   make([]domain.NominationView, 0, len(nominations)),
		TotalVotes:    len(votes),
		MyRSVP:        answers[userID],
	}

	for _, m := range members {
		d.Members = append(d.Members, domain.MemberRSVP{SquadMember: m, RSVP: answers[m.UserID], IsMe: m.UserID == userID})
	}

	roles := domain.RolesOf(byID[userID], e)

	for _, n := range nominations {
		view := domain.NominationView{
			Nomination:      n,
			NominatedByName: byID[n.NominatedBy].Name,
			Voters:          []domain.SquadMember{},
			CanRemove:       e.Status.TakesNominations() && canRemove(authorizer, roles, n, userID),
		}

		for _, v := range votes {
			if v.PlaceID != n.PlaceID {
				continue
			}

			view.VoteCount++

			if voter, ok := byID[v.UserID]; ok {
				view.Voters = append(view.Voters, voter)
			}

			if v.UserID == userID {
				view.VotedByMe = true
				d.MyVote = n.PlaceID
			}
		}

		d.Nominations = append(d.Nominations, view)
	}

	for _, v := range votes {
		if _, ok := byID[v.UserID]; ok {
			d.VotedCount++
		}
	}

	can := func(action domain.Action) bool { return domain.Allowed(authorizer, roles, action) }
	d.Permissions = domain.Permissions{
		CanEdit:        can(domain.ActionEdit) && e.Status.IsEditable(),
		CanRSVP:        can(domain.ActionRSVP) && e.Status.TakesRSVPs(),
		CanNominate:    can(domain.ActionNominate) && e.Status.TakesNominations(),
		CanVote:        can(domain.ActionVote) && e.Status.TakesVotes(),
		CanStartVoting: can(domain.ActionManage) && e.Status == domain.StatusDraft && len(nominations) > 0,
		CanCloseVoting: can(domain.ActionManage) && e.Status == domain.StatusVoting,
		CanStart:       can(domain.ActionManage) && e.Status == domain.StatusConfirmed,
		CanClose:       can(domain.ActionManage) && e.Status == domain.StatusActive,
	}

	return d
}
