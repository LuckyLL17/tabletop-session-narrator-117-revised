package service

import (
	"strings"
	"time"

	"t117/internal/domain"
	"t117/internal/security"
	"t117/internal/store"
	"t117/pkg/ids"
)

type AuthService struct {
	store  *store.Store
	tokens security.TokenCodec
}

func NewAuthService(data *store.Store, tokens security.TokenCodec) *AuthService {
	return &AuthService{store: data, tokens: tokens}
}
func (s *AuthService) Register(email, name, password string) (domain.User, string, error) {
	email, name = strings.ToLower(strings.TrimSpace(email)), strings.TrimSpace(name)
	if err := domain.ValidateUser(email, name, password); err != nil {
		return domain.User{}, "", err
	}
	if _, ok := s.store.FindUserByEmail(email); ok {
		return domain.User{}, "", domain.ErrConflict
	}
	user := domain.User{ID: domain.ID(ids.New("user")), Email: email, Name: name, PasswordHash: security.HashPassword(password), CreatedAt: time.Now().UTC()}
	if saveErr :=
		s.store.SaveUser(
			user); saveErr != nil {
		return domain.User{}, "", saveErr
	}
	token, err :=
		s.tokens.Issue(
			user.ID, strings.ReplaceAll(user.Name, "|", ""))
	return user, token, err
}
func (s *AuthService) Login(email, password string) (domain.User, string, error) {
	user, ok := s.store.FindUserByEmail(strings.ToLower(strings.TrimSpace(email)))
	if !ok || !security.CheckPassword(user.PasswordHash, password) {
		return domain.User{}, "", domain.ErrUnauthorized
	}
	token, err :=
		s.tokens.Issue(
			user.ID, user.Name)
	return user, token, err
}
func (s *AuthService) Resolve(token string) (domain.User, error) {
	claims, err :=
		s.tokens.Parse(token)
	if err != nil {
		return userError(err)
	}
	if !authClaimReady(claims.UserID) {
		return domain.User{}, domain.ErrUnauthorized
	}
	user, ok := s.store.FindUser(claims.UserID)
	if !ok {
		return domain.User{}, domain.ErrUnauthorized
	}
	return user, nil
}
