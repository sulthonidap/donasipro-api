package usecase

import (
	"clean-api/domain"
	"context"
	"errors"
)

type branchRequestUsecase struct {
	reqRepo       domain.BranchRequestRepository
	deliveryRepo  domain.DeliveryRepository
	inventoryRepo domain.InventoryRepository
	branchRepo    domain.BranchRepository
}

func NewBranchRequestUsecase(
	reqRepo domain.BranchRequestRepository,
	deliveryRepo domain.DeliveryRepository,
	inventoryRepo domain.InventoryRepository,
	branchRepo domain.BranchRepository,
) domain.BranchRequestUsecase {
	return &branchRequestUsecase{
		reqRepo:       reqRepo,
		deliveryRepo:  deliveryRepo,
		inventoryRepo: inventoryRepo,
		branchRepo:    branchRepo,
	}
}

func (u *branchRequestUsecase) CreateRequest(ctx context.Context, branchID, requestedBy uint, itemName string, category domain.InventoryCategory, quantity float64, unit, purpose string) (domain.BranchRequest, error) {
	if itemName == "" || quantity <= 0 {
		return domain.BranchRequest{}, errors.New("nama barang dan jumlah wajib diisi")
	}

	req := domain.BranchRequest{
		BranchID:    branchID,
		RequestedBy: requestedBy,
		ItemName:    itemName,
		Category:    category,
		Quantity:    quantity,
		Unit:        unit,
		Purpose:     purpose,
		Status:      domain.BranchReqPending,
	}

	err := u.reqRepo.Create(ctx, &req)
	if err != nil {
		return domain.BranchRequest{}, err
	}

	return u.reqRepo.GetByID(ctx, req.ID)
}

func (u *branchRequestUsecase) ListRequests(ctx context.Context, branchID *uint) ([]domain.BranchRequest, error) {
	return u.reqRepo.List(ctx, branchID)
}

func (u *branchRequestUsecase) ApproveAndAssignCourier(ctx context.Context, requestID uint, courierID uint) error {
	req, err := u.reqRepo.GetByID(ctx, requestID)
	if err != nil {
		return errors.New("pengajuan cabang tidak ditemukan")
	}

	if req.Status != domain.BranchReqPending {
		return errors.New("hanya pengajuan status pending yang dapat disetujui")
	}

	// 1. Create temporary/permanent inventory item at Pusat for delivery assignment
	invItem := domain.Inventory{
		ItemName:         req.ItemName,
		Category:         req.Category,
		Quantity:         req.Quantity,
		Unit:             req.Unit,
		VerifiedPhysical: true,
		DeliveryStatus:   domain.DeliveryAssigned,
	}

	err = u.inventoryRepo.Create(ctx, &invItem)
	if err != nil {
		return err
	}

	// 2. Create delivery task targeting the requesting Branch
	branchName := "Cabang Regional"
	branchAddress := "Alamat Cabang"
	branchPhone := "-"
	if req.Branch != nil {
		branchName = req.Branch.Name
		branchAddress = req.Branch.Address
		branchPhone = req.Branch.Phone
	}

	delivery := domain.Delivery{
		InventoryID:      invItem.ID,
		RecipientName:    branchName,
		RecipientAddress: branchAddress,
		RecipientPhone:   branchPhone,
		CourierID:        courierID,
		Status:           domain.StatusDeliveryPending,
	}

	err = u.deliveryRepo.Create(ctx, &delivery)
	if err != nil {
		return err
	}

	// 3. Update branch request status to approved
	return u.reqRepo.UpdateStatus(ctx, requestID, domain.BranchReqApproved, &courierID)
}

func (u *branchRequestUsecase) RejectRequest(ctx context.Context, requestID uint) error {
	req, err := u.reqRepo.GetByID(ctx, requestID)
	if err != nil {
		return errors.New("pengajuan cabang tidak ditemukan")
	}

	if req.Status != domain.BranchReqPending {
		return errors.New("hanya pengajuan status pending yang dapat ditolak")
	}

	return u.reqRepo.UpdateStatus(ctx, requestID, domain.BranchReqRejected, nil)
}

func (u *branchRequestUsecase) CompleteRequest(ctx context.Context, requestID uint) error {
	req, err := u.reqRepo.GetByID(ctx, requestID)
	if err != nil {
		return errors.New("pengajuan cabang tidak ditemukan")
	}

	// Create inventory item assigned to Branch
	branchInv := domain.Inventory{
		BranchID:         &req.BranchID,
		ItemName:         req.ItemName,
		Category:         req.Category,
		Quantity:         req.Quantity,
		Unit:             req.Unit,
		VerifiedPhysical: true,
		DeliveryStatus:   domain.DeliveryDelivered,
	}

	err = u.inventoryRepo.Create(ctx, &branchInv)
	if err != nil {
		return err
	}

	return u.reqRepo.UpdateStatus(ctx, requestID, domain.BranchReqCompleted, nil)
}
