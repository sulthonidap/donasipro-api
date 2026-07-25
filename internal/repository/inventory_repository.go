package repository

import (
	"clean-api/domain"
	"context"
	"time"

	"gorm.io/gorm"
)

type inventoryRepository struct {
	db *gorm.DB
}

func NewInventoryRepository(db *gorm.DB) domain.InventoryRepository {
	return &inventoryRepository{db: db}
}

func (r *inventoryRepository) Create(ctx context.Context, item *domain.Inventory) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *inventoryRepository) GetByID(ctx context.Context, id uint) (domain.Inventory, error) {
	var item domain.Inventory
	err := r.db.WithContext(ctx).First(&item, id).Error
	return item, err
}

func (r *inventoryRepository) List(ctx context.Context, category string, verifiedOnly bool) ([]domain.Inventory, error) {
	var items []domain.Inventory
	query := r.db.WithContext(ctx).Preload("Branch")

	if category != "" {
		query = query.Where("category = ?", category)
	}

	if verifiedOnly {
		query = query.Where("verified_physical = ?", true)
	}

	err := query.Order("created_at desc").Find(&items).Error
	return items, err
}

func (r *inventoryRepository) Update(ctx context.Context, item *domain.Inventory) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *inventoryRepository) VerifyPhysical(ctx context.Context, id uint, verifiedByID uint, expiryDate *time.Time) error {
	updates := map[string]interface{}{
		"verified_physical": true,
		"verified_by_id":    verifiedByID,
	}
	if expiryDate != nil {
		updates["expiry_date"] = expiryDate
	}
	return r.db.WithContext(ctx).Model(&domain.Inventory{}).Where("id = ?", id).Updates(updates).Error
}

func (r *inventoryRepository) UpdateDeliveryStatus(ctx context.Context, id uint, status domain.DeliveryStatus) error {
	return r.db.WithContext(ctx).Model(&domain.Inventory{}).Where("id = ?", id).Update("delivery_status", status).Error
}

// FindByNameInPusat returns all inventory items (no branch) that match the given item name (Gudang Pusat)
func (r *inventoryRepository) FindByNameInPusat(ctx context.Context, itemName string) ([]domain.Inventory, error) {
	var items []domain.Inventory
	err := r.db.WithContext(ctx).
		Where("LOWER(item_name) = LOWER(?) AND branch_id IS NULL", itemName).
		Find(&items).Error
	return items, err
}

// DeductQuantity reduces an inventory item's quantity by the given amount (partial deduct possible)
func (r *inventoryRepository) DeductQuantity(ctx context.Context, id uint, amount float64) error {
	return r.db.WithContext(ctx).
		Model(&domain.Inventory{}).
		Where("id = ?", id).
		Update("quantity", r.db.Raw("quantity - ?", amount)).Error
}
