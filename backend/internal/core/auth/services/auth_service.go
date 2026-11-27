package services

import (
	"be-simpletracker/internal/clock"
	"errors"
	"fmt"

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
	ErrInvalidBirthYear   = errors.New("birth_year must be between 1900 and the current year")
)

var dummyPasswordHash = func() []byte {
	hash, err := bcrypt.GenerateFromPassword([]byte("invalid-login-password"), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	return hash
}()

// TokenGenerator signs a session token for a stored user (see auth.GenerateToken).
type TokenGenerator func(models.User) (string, error)

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

	token, err := s.generateToken(user)
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
	// Checked after the password so a disabled account answers exactly like a wrong password.
	if user.DisabledAt != nil {
		return AuthResult{}, ErrInvalidCredentials
	}

	token, err := s.generateToken(user)
	if err != nil {
		return AuthResult{}, fmt.Errorf("%w: %v", ErrTokenGeneration, err)
	}

	user.Password = ""
	return AuthResult{Token: token, User: user}, nil
}

func (s *AuthService) CurrentUser(userID uint) (models.User, error) {
	user, err := authrepo.FindUserByID(userID)
	if err != nil {
		return models.User{}, err
	}
	user.Password = ""
	return user, nil
}

func (s *AuthService) UpdateBirthYear(userID uint, birthYear int) (models.User, error) {
	currentYear := clock.Now().Year()
	if birthYear < 1900 || birthYear > currentYear {
		return models.User{}, ErrInvalidBirthYear
	}
	user, err := authrepo.UpdateUserBirthYear(userID, &birthYear)
	if err != nil {
		return models.User{}, err
	}
	user.Password = ""
	return user, nil
}
