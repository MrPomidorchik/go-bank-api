package service

import (
	"bank-api/internal/apperror"
	"bank-api/internal/validator"
	"errors"
	"strings"
	"time"

	"bank-api/internal/models"
	"bank-api/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService interface {
	Register(req models.RegisterRequest) (*models.User, error)
	Login(req models.LoginRequest) (*models.AuthResponse, error)
}

type authService struct {
	userRepo    repository.UserRepository
	jwtSecret   string
	jwtTTLHours int
}

func NewAuthService(
	userRepo repository.UserRepository,
	jwtSecret string,
	jwtTTLHours int,
) AuthService {
	return &authService{
		userRepo:    userRepo,
		jwtSecret:   jwtSecret,
		jwtTTLHours: jwtTTLHours,
	}
}

func (s *authService) Register(req models.RegisterRequest) (*models.User, error) {
	req.Email = strings.TrimSpace(req.Email)
	req.Username = strings.TrimSpace(req.Username)

	if err := validator.Email(req.Email); err != nil {
		return nil, apperror.BadRequest(err.Error())
	}

	if err := validator.Username(req.Username); err != nil {
		return nil, apperror.BadRequest(err.Error())
	}

	if err := validator.Password(req.Password); err != nil {
		return nil, apperror.BadRequest(err.Error())
	}

	emailExists, err := s.userRepo.ExistsByEmail(req.Email)
	if err != nil {
		return nil, err
	}

	if emailExists {
		return nil, apperror.Conflict("email already exists")
	}

	usernameExists, err := s.userRepo.ExistsByUsername(req.Username)
	if err != nil {
		return nil, err
	}

	if usernameExists {
		return nil, apperror.Conflict("username already exists")
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:        req.Email,
		Username:     req.Username,
		PasswordHash: string(passwordHash),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *authService) Login(req models.LoginRequest) (*models.AuthResponse, error) {
	req.Email = strings.TrimSpace(req.Email)

	if err := validator.Email(req.Email); err != nil {
		return nil, apperror.BadRequest(err.Error())
	}

	if req.Password == "" {
		return nil, errors.New("password is required")
	}

	user, err := s.userRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, err
	}

	if user == nil {
		return nil, apperror.Unauthorized("invalid email or password")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, apperror.Unauthorized("invalid email or password")
	}

	token, err := s.generateJWT(user)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{
		Token: token,
	}, nil
}

func (s *authService) generateJWT(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"email":    user.Email,
		"username": user.Username,
		"exp":      time.Now().Add(time.Duration(s.jwtTTLHours) * time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(s.jwtSecret))
}
