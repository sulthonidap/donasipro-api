package usecase

import (
	"clean-api/domain"
	"context"
	"errors"
	"fmt"
	"math"
	"time"
)

type inventoryUsecase struct {
	inventoryRepo  domain.InventoryRepository
	donationRepo   domain.DonationRepository
	masterItemRepo domain.MasterItemRepository
}

func NewInventoryUsecase(inventoryRepo domain.InventoryRepository, donationRepo domain.DonationRepository, masterItemRepo domain.MasterItemRepository) domain.InventoryUsecase {
	return &inventoryUsecase{
		inventoryRepo:  inventoryRepo,
		donationRepo:   donationRepo,
		masterItemRepo: masterItemRepo,
	}
}

func (u *inventoryUsecase) List(ctx context.Context, category string, verifiedOnly bool) ([]domain.Inventory, error) {
	return u.inventoryRepo.List(ctx, category, verifiedOnly)
}

func (u *inventoryUsecase) VerifyPhysical(ctx context.Context, id uint, verifiedByID uint, expiryDate *time.Time) error {
	item, err := u.inventoryRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if item.VerifiedPhysical {
		return errors.New("item is already physically verified")
	}

	err = u.inventoryRepo.VerifyPhysical(ctx, id, verifiedByID, expiryDate)
	if err != nil {
		return err
	}

	if item.DonationID != nil {
		u.maybeCompleteDonationVerification(ctx, *item.DonationID)
	}

	return nil
}

// VerifyPhysicalSplit verifies a pending inventory line whose physical packages
// turned out to have different expiry dates: the first batch reuses the
// original row, remaining batches become new rows sharing the same donation.
func (u *inventoryUsecase) VerifyPhysicalSplit(ctx context.Context, id uint, verifiedByID uint, batches []domain.InventoryBatch) ([]domain.Inventory, error) {
	if len(batches) == 0 {
		return nil, errors.New("at least one batch is required")
	}

	item, err := u.inventoryRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if item.VerifiedPhysical {
		return nil, errors.New("item is already physically verified")
	}

	var total float64
	for _, b := range batches {
		if b.Quantity <= 0 {
			return nil, errors.New("each batch quantity must be greater than 0")
		}
		total += b.Quantity
	}
	if math.Abs(total-item.Quantity) > 0.001 {
		return nil, fmt.Errorf("total batch quantity (%.2f) must match original quantity (%.2f)", total, item.Quantity)
	}

	result := make([]domain.Inventory, 0, len(batches))

	item.Quantity = batches[0].Quantity
	item.ExpiryDate = batches[0].ExpiryDate
	item.VerifiedPhysical = true
	item.VerifiedByID = &verifiedByID
	if err := u.inventoryRepo.Update(ctx, &item); err != nil {
		return nil, err
	}
	result = append(result, item)

	for _, b := range batches[1:] {
		newItem := domain.Inventory{
			DonationID:       item.DonationID,
			BranchID:         item.BranchID,
			ItemName:         item.ItemName,
			Category:         item.Category,
			Quantity:         b.Quantity,
			Unit:             item.Unit,
			VerifiedPhysical: true,
			VerifiedByID:     &verifiedByID,
			DeliveryStatus:   domain.DeliveryUnassigned,
			ExpiryDate:       b.ExpiryDate,
		}
		if err := u.inventoryRepo.Create(ctx, &newItem); err != nil {
			return nil, err
		}
		result = append(result, newItem)
	}

	if item.DonationID != nil {
		u.maybeCompleteDonationVerification(ctx, *item.DonationID)
	}

	return result, nil
}

// maybeCompleteDonationVerification marks the parent donation as verified
// once every inventory line tied to it has been physically verified.
func (u *inventoryUsecase) maybeCompleteDonationVerification(ctx context.Context, donationID uint) {
	donation, err := u.donationRepo.GetByID(ctx, donationID)
	if err != nil {
		return
	}

	allVerified := true
	for _, donationItem := range donation.Items {
		if !donationItem.VerifiedPhysical {
			allVerified = false
			break
		}
	}

	if allVerified {
		_ = u.donationRepo.UpdateStatus(ctx, donationID, domain.StatusVerified)
	}
}

func (u *inventoryUsecase) CreateItemDirectly(ctx context.Context, item *domain.Inventory) (domain.Inventory, error) {
	if item.ItemName == "" {
		return domain.Inventory{}, errors.New("item name is required")
	}
	if item.Category == "" {
		item.Category = "Bahan Pokok & Sembako"
	}

	// Auto-register to Master Item Catalog if item name is not in catalog yet
	if u.masterItemRepo != nil {
		_, err := u.masterItemRepo.GetByName(ctx, item.ItemName)
		if err != nil {
			autoSKU := fmt.Sprintf("SKU-AUTO-%d", time.Now().UnixNano()%1000000)
			_ = u.masterItemRepo.Create(ctx, &domain.MasterItem{
				SKUCode:     autoSKU,
				Name:        item.ItemName,
				Category:    string(item.Category),
				Unit:        item.Unit,
				Description: "Otomatis terdaftar dari input logistik",
			})
		}
	}

	item.VerifiedPhysical = true // directly created items are verified
	item.DeliveryStatus = domain.DeliveryUnassigned

	err := u.inventoryRepo.Create(ctx, item)
	if err != nil {
		return domain.Inventory{}, err
	}

	return *item, nil
}
