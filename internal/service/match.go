package service

import (
	"sort"
	"strings"
	"time"

	"t117/internal/domain"
	"t117/internal/store"
	"t117/pkg/ids"
)

type MatchService struct {
	store *store.Store
	jobs  *JobService
}
type MatchInput struct {
	GameID     string        `json:"game_id"`
	Name       string        `json:"name"`
	Players    []PlayerInput `json:"players"`
	VariantIDs []string      `json:"variant_ids"`
}
type PlayerInput struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}
type TurnInput struct {
	Focus string `json:"focus"`
}
type EventInput struct {
	TurnID     string         `json:"turn_id"`
	SeatID     string         `json:"seat_id"`
	Kind       string         `json:"kind"`
	Label      string         `json:"label"`
	Detail     string         `json:"detail"`
	Delta      map[string]int `json:"delta"`
	ScoreDelta int            `json:"score_delta"`
	Weight     int            `json:"weight"`
}

func NewMatchService(
	data *store.Store, jobs *JobService,
) *MatchService {
	return &MatchService{store: data, jobs: jobs}
}
func (s *MatchService) Create(owner domain.ID, input MatchInput) (domain.Match, []domain.Seat, error) {
	game, ok :=
		s.store.FindGame(
			domain.ID(input.GameID))
	if !ok ||
		game.OwnerID != owner {
		return domain.Match{}, nil, domain.ErrMissing
	}
	if len(input.Players) < game.MinPlayers || len(input.Players) > game.MaxPlayers {
		return domain.Match{}, nil, domain.ErrInvalid
	}
	now := time.Now().UTC()
	match := domain.NewMatch(owner, game.ID, strings.TrimSpace(input.Name), now)
	match.ID =
		domain.ID(ids.New("match"))
	match.VariantIDs =
		idsFromStrings(
			input.VariantIDs)
	seats := []domain.Seat{}
	rangeData1 := input.Players
	for rangeIndex1 := range rangeData1 {
		player :=
			rangeData1[rangeIndex1]
		seat := domain.NewSeat(match.ID, strings.TrimSpace(player.Name), strings.TrimSpace(player.Color), rangeIndex1, game.DefaultResources, now)
		seat.ID =
			domain.ID(ids.New("seat"))
		seats = append(seats, seat)
		match.SeatIDs =
			append(
				match.SeatIDs, seat.ID)
	}
	if err := domain.ValidateMatch(match, seats); err != nil {
		return domain.Match{}, nil, err
	}
	if saveErr :=
		s.store.SaveMatch(
			match); saveErr != nil {
		return domain.Match{}, nil, saveErr
	}
	for seatIndex := range seats {
		seat := seats[seatIndex]
		if saveErr :=
			s.store.SaveSeat(
				seat); saveErr != nil {
			return domain.Match{}, nil, saveErr
		}
	}
	_ = s.store.AppendAudit("match.created", map[string]any{"match_id": match.ID, "owner_id": owner})
	return match, seats, nil
}
func (s *MatchService) Get(owner, id domain.ID) (domain.Match, []domain.Seat, error) {
	match, ok :=
		s.store.FindMatch(id)
	if !ok ||
		match.OwnerID != owner {
		return domain.Match{}, nil, domain.ErrMissing
	}
	return match, s.store.SeatsForMatch(id), nil
}
func (
	s *MatchService,
) List(
	owner domain.ID,
) []domain.Match {
	result :=
		s.store.ListMatches(owner)
	sort.Slice(
		result, func(i, j int) bool { return result[i].UpdatedAt.After(result[j].UpdatedAt) })
	return result
}
func (s *MatchService) ChangeStatus(owner, id domain.ID, target domain.MatchStatus, reason string) (domain.Match, error) {
	match, _, err :=
		s.Get(owner, id)
	if err != nil {
		return matchError(err)
	}
	if err := domain.Transition(match.Status, target); err != nil {
		return matchError(err)
	}
	now := time.Now().UTC()
	match.Status = target
	match.PauseReason =
		strings.TrimSpace(reason)
	match.Revision++
	match.UpdatedAt = now
	if target == domain.MatchLive && match.StartedAt == nil {
		match.StartedAt = &now
	}
	if target == domain.MatchFinished {
		match.FinishedAt = &now
	}
	if saveErr :=
		s.store.SaveMatch(
			match); saveErr != nil {
		return matchError(saveErr)
	}
	if target == domain.MatchFinished && s.jobs != nil {
		_ = s.jobs.Enqueue(owner, match.ID, "生成战报")
	}
	_ = s.store.AppendAudit("match.status", map[string]any{"match_id": id, "status": target})
	return match, nil
}
func (
	s *MatchService,
) Start(
	owner, id domain.ID,
) (domain.Match, error) {
	return s.ChangeStatus(owner, id, domain.MatchLive, "")
}
func (
	s *MatchService,
) Pause(
	owner, id domain.ID, reason string,
) (domain.Match, error) {
	return s.ChangeStatus(owner, id, domain.MatchPaused, reason)
}
func (
	s *MatchService,
) Resume(
	owner, id domain.ID,
) (domain.Match, error) {
	return s.ChangeStatus(owner, id, domain.MatchLive, "")
}
func (
	s *MatchService,
) Finish(
	owner, id domain.ID,
) (domain.Match, error) {
	return s.ChangeStatus(owner, id, domain.MatchFinished, "")
}
func idsFromStrings(values []string) []domain.ID {
	result := []domain.ID{}
	for valueIndex := range values {
		value := values[valueIndex]
		if strings.TrimSpace(value) != "" {
			result = append(result, domain.ID(value))
		}
	}
	return result
}
