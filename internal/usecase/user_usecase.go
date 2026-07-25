package usecase

import (
	"clean-api/domain"
	"context"
	"errors"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type userUsecase struct {
	userRepo domain.UserRepository
}

func NewUserUsecase(userRepo domain.UserRepository) domain.UserUsecase {
	return &userUsecase{userRepo: userRepo}
}

func (u *userUsecase) Login(ctx context.Context, email, password string) (string, domain.User, error) {
	user, err := u.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return "", domain.User{}, errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return "", domain.User{}, errors.New("invalid email or password")
	}

	// Generate JWT
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": user.ID,
		"role":    user.Role,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	})

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "yoursupersecretjwtkeyforcleanapp2026"
	}

	tokenString, err := token.SignedString([]byte(jwtSecret))
	if err != nil {
		return "", domain.User{}, err
	}

	return tokenString, user, nil
}

func (u *userUsecase) GetByID(ctx context.Context, id uint) (domain.User, error) {
	return u.userRepo.GetByID(ctx, id)
}

func (u *userUsecase) Setup(ctx context.Context) (string, error) {
	count, err := u.userRepo.Count(ctx)
	if err != nil {
		return "", err
	}

	if count > 0 {
		return "", errors.New("setup already completed")
	}

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
	user := domain.User{
		Name:     "Super Admin",
		Email:    "admin@cleanapp.com",
		Password: string(hashedPassword),
		Role:     domain.RoleSuperAdmin,
		Status:   "active",
	}

	err = u.userRepo.Create(ctx, &user)
	if err != nil {
		return "", err
	}

	return "Super Admin created. Email: admin@cleanapp.com, Pass: admin123", nil
}

func (u *userUsecase) ConfigureUserRole(ctx context.Context, name, email, password string, role domain.Role, branchID *uint) (domain.User, error) {
	existingUser, err := u.userRepo.GetByEmail(ctx, email)
	if err == nil && existingUser.ID > 0 {
		// Existing user -> UPDATE mode
		existingUser.Name = name
		existingUser.Role = role
		existingUser.BranchID = branchID

		if password != "" && password != "keep_existing_password_123" {
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
			if err != nil {
				return domain.User{}, err
			}
			existingUser.Password = string(hashedPassword)
		}

		err = u.userRepo.Update(ctx, &existingUser)
		if err != nil {
			return domain.User{}, err
		}
		return existingUser, nil
	}

	// New user -> CREATE mode
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return domain.User{}, err
	}

	user := domain.User{
		Name:     name,
		Email:    email,
		Password: string(hashedPassword),
		Role:     role,
		BranchID: branchID,
		Status:   "active",
	}

	err = u.userRepo.Create(ctx, &user)
	if err != nil {
		return domain.User{}, err
	}

	return user, nil
}

func (u *userUsecase) Delete(ctx context.Context, id uint) error {
	return u.userRepo.Delete(ctx, id)
}

func (u *userUsecase) ListByRole(ctx context.Context, role domain.Role) ([]domain.User, error) {
	return u.userRepo.ListByRole(ctx, role)
}

func (u *userUsecase) ListAll(ctx context.Context) ([]domain.User, error) {
	return u.userRepo.ListAll(ctx)
}

