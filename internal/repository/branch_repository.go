package repository

import (
	"clean-api/domain"
	"context"

	"gorm.io/gorm"
)

type branchRepository struct {
	db *gorm.DB
}

func NewBranchRepository(db *gorm.DB) domain.BranchRepository {
	return &branchRepository{db: db}
}

func (r *branchRepository) Create(ctx context.Context, branch *domain.Branch) error {
	return r.db.WithContext(ctx).Create(branch).Error
}

func (r *branchRepository) GetByID(ctx context.Context, id uint) (domain.Branch, error) {
	var branch domain.Branch
	err := r.db.WithContext(ctx).First(&branch, id).Error
	return branch, err
}

func (r *branchRepository) ListAll(ctx context.Context) ([]domain.Branch, error) {
	var branches []domain.Branch
	err := r.db.WithContext(ctx).Order("name asc").Find(&branches).Error
	return branches, err
}

func (r *branchRepository) Update(ctx context.Context, branch *domain.Branch) error {
	return r.db.WithContext(ctx).Save(branch).Error
}

func (r *branchRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.Branch{}, id).Error
}
