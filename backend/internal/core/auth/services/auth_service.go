package services

import (
	"be-simpletracker/internal/clock"
	"context"
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

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (AuthResult, error) {
	username, err := NormalizeUsername(input.Username)
	if err != nil {
		return AuthResult{}, err
	}
	input.Username = username
	if err := ValidatePassword(input.Password); err != nil {
		return AuthResult{}, err
	}
	_, err = authrepo.FindUserByUsername(ctx, input.Username)
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
	if err := authrepo.CreateUser(ctx, &user); err != nil {
		return AuthResult{}, fmt.Errorf("%w: %v", ErrUserCreation, err)
	}

	token, err := s.generateToken(user)
	if err != nil {
		return AuthResult{}, fmt.Errorf("%w: %v", ErrTokenGeneration, err)
	}

	user.Password = ""
	return AuthResult{Token: token, User: user}, nil
}

// LoginFailure is an ErrInvalidCredentials with what the audit log needs: the account, when
// the username exists (0 otherwise; the attempted name is never recorded), and why.
type LoginFailure struct {
	UserID uint
	Reason string // unknown_user, wrong_password, no_password, disabled
}

func (f *LoginFailure) Error() string { return ErrInvalidCredentials.Error() + " (" + f.Reason + ")" }
func (f *LoginFailure) Unwrap() error { return ErrInvalidCredentials }

func (s *AuthService) Login(ctx context.Context, input LoginInput) (AuthResult, error) {
	user, err := authrepo.FindUserByUsername(ctx, input.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			_ = bcrypt.CompareHashAndPassword(dummyPasswordHash, []byte(input.Password))
			return AuthResult{}, &LoginFailure{Reason: "unknown_user"}
		}
		return AuthResult{}, err
	}
	// An account created by the admin has no password until its setup link is used. Compare
	// against the dummy hash anyway so it answers in the same time as any other failure.
	if user.Password == "" {
		_ = bcrypt.CompareHashAndPassword(dummyPasswordHash, []byte(input.Password))
		return AuthResult{}, &LoginFailure{UserID: user.ID, Reason: "no_password"}
	}
	// bcrypt only reads the first 72 bytes, so a longer input could match on its prefix. No
	// stored password is longer (they're validated, and bcrypt refuses to hash more).
	if len(input.Password) > PasswordMaxBytes {
		_ = bcrypt.CompareHashAndPassword(dummyPasswordHash, []byte(input.Password[:PasswordMaxBytes]))
		return AuthResult{}, &LoginFailure{UserID: user.ID, Reason: "wrong_password"}
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		return AuthResult{}, &LoginFailure{UserID: user.ID, Reason: "wrong_password"}
	}
	// Checked after the password so a disabled account answers exactly like a wrong password.
	if user.DisabledAt != nil {
		return AuthResult{}, &LoginFailure{UserID: user.ID, Reason: "disabled"}
	}

	token, err := s.generateToken(user)
	if err != nil {
		return AuthResult{}, fmt.Errorf("%w: %v", ErrTokenGeneration, err)
	}

	user.Password = ""
	return AuthResult{Token: token, User: user}, nil
}

func (s *AuthService) CurrentUser(ctx context.Context, userID uint) (models.User, error) {
	user, err := authrepo.FindUserByID(ctx, userID)
	if err != nil {
		return models.User{}, err
	}
	user.Password = ""
	return user, nil
}

func (s *AuthService) UpdateBirthYear(ctx context.Context, userID uint, birthYear int) (models.User, error) {
	currentYear := clock.Now().Year()
	if birthYear < 1900 || birthYear > currentYear {
		return models.User{}, ErrInvalidBirthYear
	}
	user, err := authrepo.UpdateUserBirthYear(ctx, userID, &birthYear)
	if err != nil {
		return models.User{}, err
	}
	user.Password = ""
	return user, nil
}
