package store

import "t117/internal/domain"

func (
	s *Store,
) SaveGame(
	game domain.Game,
) error {
	return s.Update(
		func(
			data *domain.Snapshot,
		) error {
			data.Games[game.ID] = game
			return nil
		})
}

func (
	s *Store,
) FindGame(
	id domain.ID,
) (domain.Game, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	game, ok :=
		s.data.Games[id]
	return game, ok
}

func (
	s *Store,
) ListGames(
	owner domain.ID,
) []domain.Game {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return collect(
		s.data.Games,
		func(
			game domain.Game,
		) bool {
			return game.OwnerID ==
				owner
		},
	)
}

func (s *Store) AddVariant(gameID domain.ID, variant domain.RuleVariant) error {
	return s.Update(func(data *domain.Snapshot) error {
		game, ok :=
			data.Games[gameID]
		if !ok {
			return domain.ErrMissing
		}
		game.Variants = append([]domain.RuleVariant{}, game.Variants...)
		data.Games[gameID] = game
		return nil
	})
}
