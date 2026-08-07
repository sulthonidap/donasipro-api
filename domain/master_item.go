package domain

import (
	"context"
	"time"
)

type MasterItem struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	SKUCode     string    `gorm:"uniqueIndex;not null" json:"sku_code"`
	Name        string    `gorm:"not null" json:"name"`
	Category    string    `gorm:"not null" json:"category"`
	Unit        string    `gorm:"not null" json:"unit"`
	Price       float64   `json:"price"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type MasterItemRepository interface {
	Create(ctx context.Context, item *MasterItem) error
	Update(ctx context.Context, item *MasterItem) error
	Delete(ctx context.Context, id uint) error
	DeleteAll(ctx context.Context) error
	GetByID(ctx context.Context, id uint) (MasterItem, error)
	GetByName(ctx context.Context, name string) (MasterItem, error)
	List(ctx context.Context) ([]MasterItem, error)
}

type MasterItemUsecase interface {
	Create(ctx context.Context, item *MasterItem) (MasterItem, error)
	Update(ctx context.Context, id uint, item *MasterItem) (MasterItem, error)
	Delete(ctx context.Context, id uint) error
	DeleteAll(ctx context.Context) error
	GetByID(ctx context.Context, id uint) (MasterItem, error)
	List(ctx context.Context) ([]MasterItem, error)
	BulkCreate(ctx context.Context, items []MasterItem) (MasterItemBulkResult, error)
}

// MasterItemBulkResult reports what happened for each row of a bulk import
// (e.g. from an uploaded spreadsheet), since some rows may be skipped as duplicates.
type MasterItemBulkResult struct {
	Created []MasterItem     `json:"created"`
	Skipped []MasterItemSkip `json:"skipped"`
}

type MasterItemSkip struct {
	Name   string `json:"name"`
	Reason string `json:"reason"`
}
