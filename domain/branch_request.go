package domain

import (
	"context"
	"time"
)

type BranchRequestStatus string

const (
	BranchReqPending    BranchRequestStatus = "pending"
	BranchReqApproved   BranchRequestStatus = "approved"
	BranchReqRejected   BranchRequestStatus = "rejected"
	BranchReqDelivering BranchRequestStatus = "delivering"
	BranchReqCompleted  BranchRequestStatus = "completed"
)

type BranchRequest struct {
	ID          uint                `gorm:"primaryKey" json:"id"`
	BranchID    uint                `json:"branch_id"`
	Branch      *Branch             `json:"branch,omitempty" gorm:"foreignKey:BranchID"`
	RequestedBy uint                `json:"requested_by"`
	User        *User               `json:"user,omitempty" gorm:"foreignKey:RequestedBy"`
	ItemName    string              `json:"item_name"`
	Category    InventoryCategory   `json:"category"`
	Quantity    float64             `json:"quantity"`
	Unit        string              `json:"unit"`
	Purpose     string              `json:"purpose"`
	Status      BranchRequestStatus `json:"status" gorm:"default:'pending'"`
	CourierID   *uint               `json:"courier_id,omitempty"`
	Courier     *User               `json:"courier,omitempty" gorm:"foreignKey:CourierID"`
	CreatedAt   time.Time           `json:"created_at"`
	UpdatedAt   time.Time           `json:"updated_at"`
}

type BranchRequestRepository interface {
	Create(ctx context.Context, req *BranchRequest) error
	GetByID(ctx context.Context, id uint) (BranchRequest, error)
	List(ctx context.Context, branchID *uint) ([]BranchRequest, error)
	UpdateStatus(ctx context.Context, id uint, status BranchRequestStatus, courierID *uint) error
}

type BranchRequestUsecase interface {
	CreateRequest(ctx context.Context, branchID, requestedBy uint, itemName string, category InventoryCategory, quantity float64, unit, purpose string) (BranchRequest, error)
	ListRequests(ctx context.Context, branchID *uint) ([]BranchRequest, error)
	ApproveAndAssignCourier(ctx context.Context, requestID uint, courierID uint) error
	RejectRequest(ctx context.Context, requestID uint) error
	CompleteRequest(ctx context.Context, requestID uint) error
}
