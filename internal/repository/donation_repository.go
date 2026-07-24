package repository

import (
	"clean-api/domain"
	"context"

	"gorm.io/gorm"
)

type donationRepository struct {
	db *gorm.DB
}

func NewDonationRepository(db *gorm.DB) domain.DonationRepository {
	return &donationRepository{db: db}
}

func (r *donationRepository) Create(ctx context.Context, donation *domain.Donation) error {
	return r.db.WithContext(ctx).Create(donation).Error
}

func (r *donationRepository) GetByID(ctx context.Context, id uint) (domain.Donation, error) {
	var donation domain.Donation
	err := r.db.WithContext(ctx).Preload("Items").Preload("Receiver").First(&donation, id).Error
	return donation, err
}

func (r *donationRepository) List(ctx context.Context) ([]domain.Donation, error) {
	var donations []domain.Donation
	err := r.db.WithContext(ctx).Preload("Items").Preload("Receiver").Order("created_at desc").Find(&donations).Error
	return donations, err
}

func (r *donationRepository) UpdateStatus(ctx context.Context, id uint, status domain.DonationStatus) error {
	return r.db.WithContext(ctx).Model(&domain.Donation{}).Where("id = ?", id).Update("status", status).Error
}
