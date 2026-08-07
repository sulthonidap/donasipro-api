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
	PhotoURL         string            `json:"photo_url,omitempty"`
	VerifiedPhysical bool              `json:"verified_physical" gorm:"default:false"`
	VerifiedByID     *uint             `json:"verified_by_id,omitempty"`
	VerifiedAt       *time.Time        `json:"verified_at,omitempty"`
	DeliveryStatus   DeliveryStatus    `json:"delivery_status" gorm:"default:'unassigned'"`
	ExpiryDate       *time.Time        `json:"expiry_date,omitempty"`
	CreatedAt        time.Time         `json:"created_at"`
	UpdatedAt        time.Time         `json:"updated_at"`
}

// InventoryBatch represents one split of a pending inventory line during
// physical verification, when packages in the same donation line turn out
// to have different expiry dates.
type InventoryBatch struct {
	Quantity   float64
	ExpiryDate *time.Time
	PhotoURL   string
}

type InventoryRepository interface {
	Create(ctx context.Context, item *Inventory) error
	GetByID(ctx context.Context, id uint) (Inventory, error)
	List(ctx context.Context, category string, verifiedOnly bool) ([]Inventory, error)
	Update(ctx context.Context, item *Inventory) error
	VerifyPhysical(ctx context.Context, id uint, verifiedByID uint, expiryDate *time.Time, photoURL *string) error
	UpdateDeliveryStatus(ctx context.Context, id uint, status DeliveryStatus) error
	FindByNameInPusat(ctx context.Context, itemName string) ([]Inventory, error)
	DeductQuantity(ctx context.Context, id uint, amount float64) error
}

type InventoryUsecase interface {
	List(ctx context.Context, category string, verifiedOnly bool) ([]Inventory, error)
	VerifyPhysical(ctx context.Context, id uint, verifiedByID uint, expiryDate *time.Time, photoURL *string) error
	VerifyPhysicalSplit(ctx context.Context, id uint, verifiedByID uint, batches []InventoryBatch) ([]Inventory, error)
	CreateItemDirectly(ctx context.Context, item *Inventory, actorID uint) (Inventory, error)
}
