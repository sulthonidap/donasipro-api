package domain

import (
	"context"
	"time"
)

type DonationType string
type DonationStatus string
type PaymentMethod string

const (
	TypeGoods DonationType = "goods"
	TypeFunds DonationType = "funds"

	StatusPending  DonationStatus = "pending_verification"
	StatusVerified DonationStatus = "verified"
	StatusRejected DonationStatus = "rejected"

	PaymentCash     PaymentMethod = "cash"
	PaymentTransfer PaymentMethod = "transfer"
	PaymentVABCA    PaymentMethod = "va_bca"
)

type Donation struct {
	ID           uint           `gorm:"primaryKey" json:"id"`
	DonorName    string         `json:"donor_name"`
	DonorPhone   string         `json:"donor_phone"`
	DonorAddress string         `json:"donor_address"`
	DonationType   DonationType  `json:"donation_type"`
	Amount         float64       `json:"amount"`          // for funds
	PaymentMethod  PaymentMethod `json:"payment_method"`  // cash | transfer | va_bca
	ItemsDesc      string        `json:"items_desc"`      // summary/desc
	Status         DonationStatus `json:"status" gorm:"default:'pending_verification'"`
	ReceiverID   uint           `json:"receiver_id"`
	Receiver     *User          `json:"receiver,omitempty" gorm:"foreignKey:ReceiverID"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`

	// Relational data / embedded items
	Items []Inventory `json:"items,omitempty" gorm:"foreignKey:DonationID"`
}

type DonationRepository interface {
	Create(ctx context.Context, donation *Donation) error
	GetByID(ctx context.Context, id uint) (Donation, error)
	List(ctx context.Context) ([]Donation, error)
	UpdateStatus(ctx context.Context, id uint, status DonationStatus) error
}

type DonationUsecase interface {
	Submit(ctx context.Context, donation *Donation) (Donation, error)
	GetByID(ctx context.Context, id uint) (Donation, error)
	List(ctx context.Context) ([]Donation, error)
	VerifyFund(ctx context.Context, id uint) error
}
