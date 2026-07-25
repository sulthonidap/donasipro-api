package repository

import (
	"clean-api/domain"
	"context"

	"gorm.io/gorm"
)

type branchRequestRepository struct {
	db *gorm.DB
}

func NewBranchRequestRepository(db *gorm.DB) domain.BranchRequestRepository {
	return &branchRequestRepository{db: db}
}

func (r *branchRequestRepository) Create(ctx context.Context, req *domain.BranchRequest) error {
	return r.db.WithContext(ctx).Create(req).Error
}

func (r *branchRequestRepository) GetByID(ctx context.Context, id uint) (domain.BranchRequest, error) {
	var req domain.BranchRequest
	err := r.db.WithContext(ctx).
		Preload("Branch").
		Preload("User").
		Preload("Courier").
		First(&req, id).Error
	return req, err
}

func (r *branchRequestRepository) List(ctx context.Context, branchID *uint) ([]domain.BranchRequest, error) {
	var requests []domain.BranchRequest
	query := r.db.WithContext(ctx).
		Preload("Branch").
		Preload("User").
		Preload("Courier").
		Order("created_at desc")

	if branchID != nil && *branchID > 0 {
		query = query.Where("branch_id = ?", *branchID)
	}

	err := query.Find(&requests).Error
	return requests, err
}

func (r *branchRequestRepository) UpdateStatus(ctx context.Context, id uint, status domain.BranchRequestStatus, courierID *uint) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if courierID != nil {
		updates["courier_id"] = courierID
	}
	return r.db.WithContext(ctx).Model(&domain.BranchRequest{}).Where("id = ?", id).Updates(updates).Error
}

func (r *branchRequestRepository) UpdateApprover(ctx context.Context, id uint, approverName string) error {
	return r.db.WithContext(ctx).Model(&domain.BranchRequest{}).Where("id = ?", id).Update("approver_name", approverName).Error
}
