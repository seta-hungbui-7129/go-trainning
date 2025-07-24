package services

import (
	"errors"
	"fmt"

	"github.com/google/uuid"
	"seta-training/internal/models"
	"seta-training/internal/repositories"
	"seta-training/pkg/auth"
)

type UserService struct {
	userRepo   *repositories.UserRepository
	jwtManager *auth.JWTManager
}

func NewUserService(userRepo *repositories.UserRepository, jwtManager *auth.JWTManager) *UserService {
	return &UserService{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

type CreateUserInput struct {
	Username string          `json:"username" binding:"required,min=3,max=50"`
	Email    string          `json:"email" binding:"required,email"`
	Password string          `json:"password" binding:"required,min=6"`
	Role     models.UserRole `json:"role" binding:"required,oneof=manager member"`
}

type LoginInput struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	User  *models.User `json:"user"`
	Token string       `json:"token"`
}

func (s *UserService) CreateUser(input *CreateUserInput) (*models.User, error) {
	// Check if email already exists
	if exists, err := s.userRepo.EmailExists(input.Email); err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	} else if exists {
		return nil, errors.New("email already exists")
	}

	// Check if username already exists
	if exists, err := s.userRepo.UsernameExists(input.Username); err != nil {
		return nil, fmt.Errorf("failed to check username existence: %w", err)
	} else if exists {
		return nil, errors.New("username already exists")
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		Username:     input.Username,
		Email:        input.Email,
		PasswordHash: hashedPassword,
		Role:         input.Role,
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user, nil
}

func (s *UserService) Login(input *LoginInput) (*LoginResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(input.Email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Check password
	if err := auth.CheckPassword(user.PasswordHash, input.Password); err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Generate JWT token
	token, err := s.jwtManager.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &LoginResponse{
		User:  user,
		Token: token,
	}, nil
}

func (s *UserService) GetUserByID(id uuid.UUID) (*models.User, error) {
	return s.userRepo.GetByID(id)
}

func (s *UserService) GetAllUsers() ([]models.User, error) {
	return s.userRepo.GetAll()
}

func (s *UserService) ValidateToken(tokenString string) (*auth.Claims, error) {
	return s.jwtManager.ValidateToken(tokenString)
}
