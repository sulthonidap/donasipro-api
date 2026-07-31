package repository

import (
	"clean-api/domain"
	"context"
	"strings"

	"gorm.io/gorm"
)

type inventoryMovementRepository struct {
	db *gorm.DB
}

func NewInventoryMovementRepository(db *gorm.DB) domain.InventoryMovementRepository {
	return &inventoryMovementRepository{db: db}
}

func (r *inventoryMovementRepository) Create(ctx context.Context, m *domain.InventoryMovement) error {
	return r.db.WithContext(ctx).Create(m).Error
}

func (r *inventoryMovementRepository) List(ctx context.Context, filter domain.InventoryMovementFilter) ([]domain.InventoryMovement, error) {
	var movements []domain.InventoryMovement
	query := r.db.WithContext(ctx).
		Preload("Branch").
		Preload("Actor").
		Order("created_at desc")

	if filter.ItemName != "" {
		query = query.Where("LOWER(item_name) LIKE ?", "%"+strings.ToLower(filter.ItemName)+"%")
	}
	if filter.Direction != "" {
		query = query.Where("direction = ?", filter.Direction)
	}
	if filter.Source != "" {
		query = query.Where("source = ?", filter.Source)
	}
	if filter.BranchID != nil {
		query = query.Where("branch_id = ?", *filter.BranchID)
	}
	if filter.DateFrom != nil {
		query = query.Where("created_at >= ?", *filter.DateFrom)
	}
	if filter.DateTo != nil {
		query = query.Where("created_at <= ?", *filter.DateTo)
	}

	err := query.Find(&movements).Error
	return movements, err
}
