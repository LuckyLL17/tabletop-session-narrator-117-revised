package store

import "t117/internal/domain"

func (
	s *Store,
) SaveUser(
	user domain.User,
) error {
	return s.Update(
		func(
			data *domain.Snapshot,
		) error {
			data.Users[user.ID] = user
			return nil
		})
}

func (
	s *Store,
) FindUser(
	id domain.ID,
) (domain.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	user, ok :=
		s.data.Users[id]
	return user, ok
}

func (s *Store) FindUserByEmail(email string) (domain.User, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for id := range s.data.Users {
		user := s.data.Users[id]
		if user.Email == email {
			return user, true
		}
	}
	return domain.User{}, false
}

func (s *Store) CountUsers() int { s.mu.RLock(); defer s.mu.RUnlock(); return len(s.data.Users) }
