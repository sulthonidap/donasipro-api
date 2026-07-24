package domain

import (
	"context"
	"time"

	"gorm.io/gorm"
)

type Branch struct {
	ID        uint           `gorm:"primaryKey" json:"id"`
	Name      string         `json:"name"`
	Address   string         `json:"address"`
	Phone     string         `json:"phone"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type BranchRepository interface {
	Create(ctx context.Context, branch *Branch) error
	GetByID(ctx context.Context, id uint) (Branch, error)
	ListAll(ctx context.Context) ([]Branch, error)
	Update(ctx context.Context, branch *Branch) error
	Delete(ctx context.Context, id uint) error
}

type BranchUsecase interface {
	CreateBranch(ctx context.Context, name, address, phone string) (Branch, error)
	GetByID(ctx context.Context, id uint) (Branch, error)
	ListBranches(ctx context.Context) ([]Branch, error)
	UpdateBranch(ctx context.Context, id uint, name, address, phone string) (Branch, error)
	DeleteBranch(ctx context.Context, id uint) error
}
