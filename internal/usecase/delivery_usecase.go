package usecase

import (
	"clean-api/domain"
	"context"
	"errors"
	"time"
)

type deliveryUsecase struct {
	deliveryRepo  domain.DeliveryRepository
	inventoryRepo domain.InventoryRepository
	branchReqRepo domain.BranchRequestRepository
}

func NewDeliveryUsecase(deliveryRepo domain.DeliveryRepository, inventoryRepo domain.InventoryRepository, branchReqRepo domain.BranchRequestRepository) domain.DeliveryUsecase {
	return &deliveryUsecase{
		deliveryRepo:  deliveryRepo,
		inventoryRepo: inventoryRepo,
		branchReqRepo: branchReqRepo,
	}
}

func (u *deliveryUsecase) Create(ctx context.Context, delivery *domain.Delivery) (domain.Delivery, error) {
	if delivery.InventoryID == 0 {
		return domain.Delivery{}, errors.New("inventory item ID is required")
	}
	if delivery.CourierID == 0 {
		return domain.Delivery{}, errors.New("courier ID is required")
	}

	item, err := u.inventoryRepo.GetByID(ctx, delivery.InventoryID)
	if err != nil {
		return domain.Delivery{}, errors.New("inventory item not found")
	}

	if !item.VerifiedPhysical {
		return domain.Delivery{}, errors.New("item must be physically verified before delivery")
	}

	if item.DeliveryStatus != domain.DeliveryUnassigned {
		return domain.Delivery{}, errors.New("item is already assigned or delivered")
	}

	delivery.Status = domain.StatusDeliveryPending

	err = u.deliveryRepo.Create(ctx, delivery)
	if err != nil {
		return domain.Delivery{}, err
	}

	// Update delivery status of inventory item to assigned
	err = u.inventoryRepo.UpdateDeliveryStatus(ctx, delivery.InventoryID, domain.DeliveryAssigned)
	if err != nil {
		return domain.Delivery{}, err
	}

	return *delivery, nil
}

func (u *deliveryUsecase) GetByID(ctx context.Context, id uint) (domain.Delivery, error) {
	return u.deliveryRepo.GetByID(ctx, id)
}

func (u *deliveryUsecase) ListForCourier(ctx context.Context, courierID uint) ([]domain.Delivery, error) {
	return u.deliveryRepo.List(ctx, courierID)
}

func (u *deliveryUsecase) ListAll(ctx context.Context) ([]domain.Delivery, error) {
	return u.deliveryRepo.ListAll(ctx)
}

func (u *deliveryUsecase) StartDelivery(ctx context.Context, id uint) error {
	delivery, err := u.deliveryRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if delivery.Status != domain.StatusDeliveryPending {
		return errors.New("delivery can only be started from pending status")
	}

	delivery.Status = domain.StatusDeliveryOngoing
	return u.deliveryRepo.Update(ctx, &delivery)
}

func (u *deliveryUsecase) UploadProof(ctx context.Context, id uint, proofFilename string) error {
	delivery, err := u.deliveryRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if delivery.Status != domain.StatusDeliveryOngoing {
		return errors.New("delivery must be ongoing to upload proof")
	}

	now := time.Now()
	delivery.Status = domain.StatusDeliveryDelivered
	delivery.ProofPhoto = proofFilename
	delivery.DeliveredAt = &now

	err = u.deliveryRepo.Update(ctx, &delivery)
	if err != nil {
		return err
	}

	// Update source inventory item delivery status to delivered
	_ = u.inventoryRepo.UpdateDeliveryStatus(ctx, delivery.InventoryID, domain.DeliveryDelivered)

	// If this delivery is linked to a branch request, create inventory at branch and complete request
	if u.branchReqRepo != nil && delivery.BranchRequestID != nil {
		req, reqErr := u.branchReqRepo.GetByID(ctx, *delivery.BranchRequestID)
		if reqErr == nil && req.Status == domain.BranchReqApproved {
			// Create inventory item at the destination branch
			branchInv := domain.Inventory{
				BranchID:         &req.BranchID,
				ItemName:         req.ItemName,
				Category:         req.Category,
				Quantity:         req.Quantity,
				Unit:             req.Unit,
				VerifiedPhysical: true,
				DeliveryStatus:   domain.DeliveryDelivered,
			}
			_ = u.inventoryRepo.Create(ctx, &branchInv)

			// Mark branch request as completed
			_ = u.branchReqRepo.UpdateStatus(ctx, req.ID, domain.BranchReqCompleted, nil)
		}
	}

	return nil
}
