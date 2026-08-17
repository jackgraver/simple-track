package services

import (
	"errors"
	"fmt"
	"time"

	"be-simpletracker/internal/core/auth/models"
	authrepo "be-simpletracker/internal/core/auth/repository"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUsernameExists     = errors.New("username already exists")
	ErrPasswordHash       = errors.New("failed to hash password")
	ErrUserCreation       = errors.New("failed to create user")
	ErrTokenGeneration    = errors.New("failed to generate token")
)

var dummyPasswordHash = func() []byte {
	hash, err := bcrypt.GenerateFromPassword([]byte("invalid-login-password"), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return hash
}()

type TokenGenerator func(string) (string, error)

type AuthService struct {
	generateToken TokenGenerator
}

type RegisterInput struct {
	Username string
	Password string
	Email    string
}

type LoginInput struct {
	Username string
	Password string
}

type AuthResult struct {
	Token string
	User  models.User
}

func NewAuthService(generateToken TokenGenerator) *AuthService {
	return &AuthService{generateToken: generateToken}
}

func (s *AuthService) Register(input RegisterInput) (AuthResult, error) {
	_, err := authrepo.FindUserByUsername(input.Username)
	if err == nil {
		return AuthResult{}, ErrUsernameExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return AuthResult{}, err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return AuthResult{}, fmt.Errorf("%w: %v", ErrPasswordHash, err)
	}

	user := models.User{
		Username: input.Username,
		Password: string(hashedPassword),
		Email:    input.Email,
	}
	if err := authrepo.CreateUser(&user); err != nil {
		return AuthResult{}, fmt.Errorf("%w: %v", ErrUserCreation, err)
	}

	token, err := s.generateToken(user.Username)
	if err != nil {
		return AuthResult{}, fmt.Errorf("%w: %v", ErrTokenGeneration, err)
	}

	user.Password = ""
	return AuthResult{Token: token, User: user}, nil
}

func (s *AuthService) Login(input LoginInput) (AuthResult, error) {
	user, err := authrepo.FindUserByUsername(input.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			_ = bcrypt.CompareHashAndPassword(dummyPasswordHash, []byte(input.Password))
			return AuthResult{}, ErrInvalidCredentials
		}
		return AuthResult{}, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return AuthResult{}, ErrInvalidCredentials
	}

	token, err := s.generateToken(user.Username)
	if err != nil {
		return AuthResult{}, fmt.Errorf("%w: %v", ErrTokenGeneration, err)
	}

	user.Password = ""
	return AuthResult{Token: token, User: user}, nil
}

func (s *AuthService) CurrentUser(username string) (models.User, error) {
	user, err := authrepo.FindUserByUsername(username)
	if err != nil {
		return models.User{}, err
	}
	user.Password = ""
	return user, nil
}

func (s *AuthService) UpdateBirthYear(username string, birthYear int) (models.User, error) {
	currentYear := time.Now().Year()
	if birthYear < 1900 || birthYear > currentYear {
		return models.User{}, errors.New("birth_year must be between 1900 and the current year")
	}
	user, err := authrepo.UpdateUserBirthYear(username, &birthYear)
	if err != nil {
		return models.User{}, err
	}
	user.Password = ""
	return user, nil
}
