package usecase

import (
	"clean-api/domain"
	"context"
	"errors"
	"fmt"
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

func (u *branchRequestUsecase) CreateRequest(
	ctx context.Context,
	branchID, requestedBy uint,
	itemName string,
	category domain.InventoryCategory,
	routineQuota, remainingStock, quantity float64,
	unit, monthPeriod, applicantName, purpose string,
) (domain.BranchRequest, error) {
	if itemName == "" || quantity <= 0 {
		return domain.BranchRequest{}, errors.New("nama barang dan jumlah wajib diisi")
	}

	req := domain.BranchRequest{
		BranchID:       branchID,
		RequestedBy:    requestedBy,
		ItemName:       itemName,
		Category:       category,
		RoutineQuota:   routineQuota,
		RemainingStock: remainingStock,
		Quantity:       quantity,
		Unit:           unit,
		MonthPeriod:    monthPeriod,
		ApplicantName:  applicantName,
		Purpose:        purpose,
		Status:         domain.BranchReqPending,
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

func (u *branchRequestUsecase) ApproveAndAssignCourier(ctx context.Context, requestID uint, courierID uint, approverName string) error {
	req, err := u.reqRepo.GetByID(ctx, requestID)
	if err != nil {
		return errors.New("pengajuan cabang tidak ditemukan")
	}

	if req.Status != domain.BranchReqPending {
		return errors.New("hanya pengajuan status pending yang dapat disetujui")
	}

	// 1. Find matching stock in Gudang Pusat
	pusatItems, err := u.inventoryRepo.FindByNameInPusat(ctx, req.ItemName)
	if err != nil {
		return errors.New("gagal memeriksa stok gudang pusat")
	}

	// Calculate total available Pusat stock
	var totalPusatStock float64
	for _, item := range pusatItems {
		totalPusatStock += item.Quantity
	}

	if totalPusatStock < req.Quantity {
		return fmt.Errorf("stok gudang pusat tidak mencukupi: tersedia %.0f %s, dibutuhkan %.0f %s",
			totalPusatStock, req.Unit, req.Quantity, req.Unit)
	}

	// 2. Deduct from Pusat stock (FIFO: deduct from first matching items)
	remaining := req.Quantity
	var usedItemID uint
	for _, item := range pusatItems {
		if remaining <= 0 {
			break
		}
		deduct := remaining
		if item.Quantity < deduct {
			deduct = item.Quantity
		}
		if err := u.inventoryRepo.DeductQuantity(ctx, item.ID, deduct); err != nil {
			return errors.New("gagal mengurangi stok gudang pusat")
		}
		remaining -= deduct
		usedItemID = item.ID
	}

	// 3. Create delivery task using the (last) Pusat inventory item
	branchName := ""
	branchAddress := ""
	branchPhone := ""
	if req.Branch != nil {
		branchName = req.Branch.Name
		branchAddress = req.Branch.Address
		branchPhone = req.Branch.Phone
	}

	delivery := domain.Delivery{
		InventoryID:     usedItemID,
		BranchRequestID: &requestID,
		RecipientName:   branchName,
		RecipientAddress: branchAddress,
		RecipientPhone:  branchPhone,
		CourierID:       courierID,
		Status:          domain.StatusDeliveryPending,
	}

	err = u.deliveryRepo.Create(ctx, &delivery)
	if err != nil {
		return err
	}

	if approverName != "" {
		_ = u.reqRepo.UpdateApprover(ctx, requestID, approverName)
	}

	// 4. Update branch request status to approved
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
