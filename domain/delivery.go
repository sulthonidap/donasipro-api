package domain

import (
	"context"
	"time"
)

type DeliveryStatusState string

const (
	StatusDeliveryPending   DeliveryStatusState = "pending"
	StatusDeliveryOngoing   DeliveryStatusState = "ongoing"
	StatusDeliveryDelivered DeliveryStatusState = "delivered"
)

type Delivery struct {
	ID               uint                `gorm:"primaryKey" json:"id"`
	InventoryID      uint                `json:"inventory_id"`
	RecipientName    string              `json:"recipient_name"`
	RecipientAddress string              `json:"recipient_address"`
	RecipientPhone   string              `json:"recipient_phone"`
	CourierID        uint                `json:"courier_id"`
	Status           DeliveryStatusState `json:"status" gorm:"default:'pending'"`
	ProofPhoto       string              `json:"proof_photo"`
	DeliveredAt      *time.Time          `json:"delivered_at,omitempty"`
	CreatedAt        time.Time           `json:"created_at"`
	UpdatedAt        time.Time           `json:"updated_at"`

	// Relational data / embedded fields
	Item Inventory `json:"item,omitempty" gorm:"foreignKey:InventoryID"`
}

type DeliveryRepository interface {
	Create(ctx context.Context, delivery *Delivery) error
	GetByID(ctx context.Context, id uint) (Delivery, error)
	List(ctx context.Context, courierID uint) ([]Delivery, error)
	ListAll(ctx context.Context) ([]Delivery, error)
	Update(ctx context.Context, delivery *Delivery) error
}

type DeliveryUsecase interface {
	Create(ctx context.Context, delivery *Delivery) (Delivery, error)
	GetByID(ctx context.Context, id uint) (Delivery, error)
	ListForCourier(ctx context.Context, courierID uint) ([]Delivery, error)
	ListAll(ctx context.Context) ([]Delivery, error)
	StartDelivery(ctx context.Context, id uint) error
	UploadProof(ctx context.Context, id uint, proofFilename string) error
}
