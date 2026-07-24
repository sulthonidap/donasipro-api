package domain

import (
	"context"
	"time"
)

type InventoryCategory string
type DeliveryStatus string

const (
	CategoryBahanPokok  InventoryCategory = "Bahan Pokok"
	CategoryMakananAnak InventoryCategory = "Makanan Anak"

	DeliveryUnassigned DeliveryStatus = "unassigned"
	DeliveryAssigned   DeliveryStatus = "assigned"
	DeliveryDelivered  DeliveryStatus = "delivered"
)

type Inventory struct {
	ID               uint              `gorm:"primaryKey" json:"id"`
	DonationID       *uint             `json:"donation_id,omitempty"`
	BranchID         *uint             `json:"branch_id,omitempty"`
	Branch           *Branch           `json:"branch,omitempty" gorm:"foreignKey:BranchID"`
	ItemName         string            `json:"item_name"`
	Category         InventoryCategory `json:"category"`
	Quantity         float64           `json:"quantity"`
	Unit             string            `json:"unit"`
	VerifiedPhysical bool              `json:"verified_physical" gorm:"default:false"`
	VerifiedByID     *uint             `json:"verified_by_id,omitempty"`
	DeliveryStatus   DeliveryStatus    `json:"delivery_status" gorm:"default:'unassigned'"`
	ExpiryDate       *time.Time        `json:"expiry_date,omitempty"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}

type InventoryRepository interface {
	Create(ctx context.Context, item *Inventory) error
	GetByID(ctx context.Context, id uint) (Inventory, error)
	List(ctx context.Context, category string, verifiedOnly bool) ([]Inventory, error)
	Update(ctx context.Context, item *Inventory) error
	VerifyPhysical(ctx context.Context, id uint, verifiedByID uint) error
	UpdateDeliveryStatus(ctx context.Context, id uint, status DeliveryStatus) error
}

type InventoryUsecase interface {
	List(ctx context.Context, category string, verifiedOnly bool) ([]Inventory, error)
	VerifyPhysical(ctx context.Context, id uint, verifiedByID uint) error
	CreateItemDirectly(ctx context.Context, item *Inventory) (Inventory, error)
}
