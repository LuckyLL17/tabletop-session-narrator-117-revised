package service

import (
	"sort"
	"time"

	"t117/internal/domain"
	"t117/pkg/ids"
)

func (
	s *MatchService,
) Milestones(
	owner, matchID domain.ID,
) (
	[]domain.Milestone, error) {
	if _, _, err :=
		s.Get(owner, matchID); err != nil {
		return nil, err
	}
	rows :=
		s.store.MilestonesForMatch(
			matchID)
	sort.Slice(
		rows, func(i, j int) bool {
			if rows[i].Importance == rows[j].Importance {
				return rows[i].CreatedAt.Before(rows[j].CreatedAt)
			}
			left := rows[i].Importance
			right := rows[j].Importance
			return left > right
		})
	return rows, nil
}
func (s *MatchService) AddMilestone(owner, matchID domain.ID, eventID domain.ID, title, explanation string, importance int) (domain.Milestone, error) {
	if _, _, err :=
		s.Get(owner, matchID); err != nil {
		return milestoneError(err)
	}
	eventRows :=
		s.store.EventsForMatch(
			matchID)
	for eventIndex := range eventRows {
		event :=
			eventRows[eventIndex]
		if event.ID == eventID {
			item :=
				domain.Milestone{ID: domain.ID(ids.New("milestone")), MatchID: matchID, TurnID: event.TurnID, EventID: eventID, Title: title, Explanation: explanation, Importance: importance, CreatedAt: time.Now().UTC()}
			if saveErr :=
				s.store.SaveMilestone(
					item); saveErr != nil {
				return milestoneError(saveErr)
			}
			return item, nil
		}
	}
	return domain.Milestone{}, domain.ErrMissing
}
