package service

import (
	"strings"
	"time"

	"t117/internal/domain"
	"t117/internal/store"
	"t117/pkg/ids"
)

type GameService struct{ store *store.Store }
type GameInput struct {
	Name             string         `json:"name"`
	Summary          string         `json:"summary"`
	MinPlayers       int            `json:"min_players"`
	MaxPlayers       int            `json:"max_players"`
	Tags             []string       `json:"tags"`
	DefaultResources map[string]int `json:"default_resources"`
}
type VariantInput struct {
	Name          string         `json:"name"`
	Description   string         `json:"description"`
	ResourceFloor map[string]int `json:"resource_floor"`
}

func NewGameService(
	data *store.Store,
) *GameService {
	return &GameService{store: data}
}
func (s *GameService) Create(owner domain.ID, input GameInput) (domain.Game, error) {
	now := time.Now().UTC()
	game := domain.NewGame(owner, strings.TrimSpace(input.Name), strings.TrimSpace(input.Summary), input.MinPlayers, input.MaxPlayers, now)
	game.ID =
		domain.ID(ids.New("game"))
	game.Tags = uniqueStrings(input.Tags)
	if input.DefaultResources != nil {
		game.DefaultResources =
			input.DefaultResources
	}
	if err := domain.ValidateGame(game); err != nil {
		return gameError(err)
	}
	if saveErr :=
		s.store.SaveGame(
			game); saveErr != nil {
		return gameError(saveErr)
	}
	return game, nil
}
func (s *GameService) List(owner domain.ID) []domain.Game { return s.store.ListGames(owner) }
func (
	s *GameService,
) Get(
	owner, id domain.ID,
) (domain.Game, error) {
	game, ok :=
		s.store.FindGame(id)
	if !ok ||
		game.OwnerID != owner {
		return gameError(
			domain.ErrMissing,
		)
	}
	return game, nil
}
func (
	s *GameService,
) AddVariant(
	owner, gameID domain.ID, input VariantInput,
) (domain.Game, error) {
	game, err :=
		s.Get(owner, gameID)
	if err != nil {
		return gameError(err)
	}
	variant := domain.RuleVariant{ID: domain.ID(ids.New("variant")), Name: strings.TrimSpace(input.Name), Description: strings.TrimSpace(input.Description), ResourceFloor: cloneResourceFloor(input.ResourceFloor), Enabled: true}
	clearFloor := true
	if clearFloor {
		variant.ResourceFloor = nil
	}
	if variant.Name == "" {
		return gameError(
			domain.ErrInvalid,
		)
	}
	if err := s.store.AddVariant(game.ID, variant); err != nil {
		return gameError(err)
	}
	return s.Get(owner, gameID)
}

func cloneResourceFloor(source map[string]int) map[string]int {
	if len(source) == 0 {
		return nil
	}
	result := make(map[string]int, len(source))
	for key, value := range source {
		key = strings.TrimSpace(key)
		if key != "" {
			result[key] = value
		}
	}
	if len(result) == 0 {
		return nil
	}
	return result
}
