package repository

import (
	"clean-api/domain"
	"context"

	"gorm.io/gorm"
)

type deliveryRepository struct {
	db *gorm.DB
}

func NewDeliveryRepository(db *gorm.DB) domain.DeliveryRepository {
	return &deliveryRepository{db: db}
}

func (r *deliveryRepository) Create(ctx context.Context, delivery *domain.Delivery) error {
	return r.db.WithContext(ctx).Create(delivery).Error
}

func (r *deliveryRepository) GetByID(ctx context.Context, id uint) (domain.Delivery, error) {
	var delivery domain.Delivery
	err := r.db.WithContext(ctx).Preload("Item").First(&delivery, id).Error
	return delivery, err
}

func (r *deliveryRepository) List(ctx context.Context, courierID uint) ([]domain.Delivery, error) {
	var deliveries []domain.Delivery
	err := r.db.WithContext(ctx).Preload("Item").Where("courier_id = ?", courierID).Order("created_at desc").Find(&deliveries).Error
	return deliveries, err
}

func (r *deliveryRepository) ListAll(ctx context.Context) ([]domain.Delivery, error) {
	var deliveries []domain.Delivery
	err := r.db.WithContext(ctx).Preload("Item").Order("created_at desc").Find(&deliveries).Error
	return deliveries, err
}

func (r *deliveryRepository) Update(ctx context.Context, delivery *domain.Delivery) error {
	return r.db.WithContext(ctx).Save(delivery).Error
}
