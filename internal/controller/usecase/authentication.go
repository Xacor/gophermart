package usecase

import (
	"context"
	"errors"

	"github.com/Xacor/gophermart/internal/entity"
	"github.com/Xacor/gophermart/internal/utils/jwt"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

const passwordHashCost = 10

type AuthUseCase struct {
	repo      UserRepo
	secretKey string
}

func NewAuthUseCase(repo UserRepo, secretKey string) *AuthUseCase {
	return &AuthUseCase{repo, secretKey}
}

var ErrUserExists = errors.New("user exists")

func (a *AuthUseCase) Register(ctx context.Context, user entity.User) (string, error) {
	_, err := a.repo.GetByLogin(ctx, user.Login)
	if err == nil {
		return "", ErrUserExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), passwordHashCost)
	if err != nil {
		return "", err
	}

	user.Password = string(hash)
	if err = a.repo.CreateUser(ctx, user); err != nil {
		return "", err
	}

	createdUser, err := a.repo.GetByLogin(ctx, user.Login)
	if err != nil {
		return "", err
	}

	return jwt.BuildToken(createdUser.ID, a.secretKey)
}

var ErrInvalidCredentials = errors.New("invalid credentials")

func (a *AuthUseCase) Authenticate(ctx context.Context, user entity.User) (string, error) {
	reqUser, err := a.repo.GetByLogin(ctx, user.Login)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrInvalidCredentials
	} else if err != nil {
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(reqUser.Password), []byte(user.Password))
	if err != nil {
		return "", ErrInvalidCredentials
	}

	return jwt.BuildToken(reqUser.ID, a.secretKey)
}
