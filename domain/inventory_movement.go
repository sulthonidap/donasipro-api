package domain

import (
	"context"
	"time"
)

type MovementDirection string
type MovementSource string

const (
	MovementIn  MovementDirection = "in"
	MovementOut MovementDirection = "out"

	SourceDonation      MovementSource = "donation"
	SourceManual        MovementSource = "manual"
	SourceBranchRequest MovementSource = "branch_request"
	SourceDelivery      MovementSource = "delivery"
)

// InventoryMovement is an append-only ledger entry recording a single stock
// in/out event. Inventory.Quantity itself is a mutable snapshot with no
// history, so every place that changes it must also write one of these.
type InventoryMovement struct {
	ID              uint              `gorm:"primaryKey" json:"id"`
	InventoryID     *uint             `json:"inventory_id,omitempty"`
	ItemName        string            `json:"item_name"`
	Category        InventoryCategory `json:"category"`
	Unit            string            `json:"unit"`
	Direction       MovementDirection `json:"direction"`
	Quantity        float64           `json:"quantity"`
	Source          MovementSource    `json:"source"`
	BranchID        *uint             `json:"branch_id,omitempty"`
	Branch          *Branch           `json:"branch,omitempty" gorm:"foreignKey:BranchID"`
	DonationID      *uint             `json:"donation_id,omitempty"`
	BranchRequestID *uint             `json:"branch_request_id,omitempty"`
	DeliveryID      *uint             `json:"delivery_id,omitempty"`
	ActorID         *uint             `json:"actor_id,omitempty"`
	Actor           *User             `json:"actor,omitempty" gorm:"foreignKey:ActorID"`
	Note            string            `json:"note"`
	CreatedAt       time.Time         `json:"created_at"`
}

type InventoryMovementFilter struct {
	ItemName  string
	Direction MovementDirection
	Source    MovementSource
	BranchID  *uint
	DateFrom  *time.Time
	DateTo    *time.Time
}

type InventoryMovementRepository interface {
	Create(ctx context.Context, m *InventoryMovement) error
	List(ctx context.Context, filter InventoryMovementFilter) ([]InventoryMovement, error)
}

type InventoryMovementUsecase interface {
	List(ctx context.Context, filter InventoryMovementFilter) ([]InventoryMovement, error)
}
