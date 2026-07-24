package repository

import (
	"clean-api/domain"
	"context"

	"gorm.io/gorm"
)

type masterItemRepository struct {
	db *gorm.DB
}

func NewMasterItemRepository(db *gorm.DB) domain.MasterItemRepository {
	return &masterItemRepository{db: db}
}

func (r *masterItemRepository) Create(ctx context.Context, item *domain.MasterItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *masterItemRepository) Update(ctx context.Context, item *domain.MasterItem) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *masterItemRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&domain.MasterItem{}, id).Error
}

func (r *masterItemRepository) GetByID(ctx context.Context, id uint) (domain.MasterItem, error) {
	var item domain.MasterItem
	err := r.db.WithContext(ctx).First(&item, id).Error
	return item, err
}

func (r *masterItemRepository) GetByName(ctx context.Context, name string) (domain.MasterItem, error) {
	var item domain.MasterItem
	err := r.db.WithContext(ctx).Where("LOWER(name) = LOWER(?)", name).First(&item).Error
	return item, err
}

func (r *masterItemRepository) List(ctx context.Context) ([]domain.MasterItem, error) {
	var items []domain.MasterItem
	err := r.db.WithContext(ctx).Order("created_at desc").Find(&items).Error
	return items, err
}
