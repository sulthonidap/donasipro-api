package domain

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type Role string

const (
	RoleSuperAdmin       Role = "superadmin"
	RoleDonationReceiver Role = "donation_receiver"
	RoleLogistics        Role = "logistics"
	RoleFinance          Role = "finance"
	RoleDelivery         Role = "delivery"
	RoleBranchStaff      Role = "branch_staff"
)

type User struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `json:"name"`
	Email     string         `gorm:"unique" json:"email"`
	Password  string         `json:"-"`
	Role      Role           `json:"role"`
	Status    string         `json:"status" gorm:"default:'active'"`
	BranchID  *uint          `json:"branch_id"`
	Branch    *Branch        `json:"branch,omitempty" gorm:"foreignKey:BranchID"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type UserRepository interface {
	GetByEmail(ctx context.Context, email string) (User, error)
	GetByID(ctx context.Context, id uint) (User, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	Count(ctx context.Context) (int64, error)
	ListByRole(ctx context.Context, role Role) ([]User, error)
	ListAll(ctx context.Context) ([]User, error)
}

type UserUsecase interface {
	Login(ctx context.Context, email, password string) (string, User, error)
	GetByID(ctx context.Context, id uint) (User, error)
	Setup(ctx context.Context) (string, error)
	ConfigureUserRole(ctx context.Context, name, email, password string, role Role, branchID *uint) (User, error)
	ListByRole(ctx context.Context, role Role) ([]User, error)
	ListAll(ctx context.Context) ([]User, error)
}

