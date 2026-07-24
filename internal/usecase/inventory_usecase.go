package usecase

import (
	"clean-api/domain"
	"context"
	"errors"
	"fmt"
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

func (u *inventoryUsecase) VerifyPhysical(ctx context.Context, id uint, verifiedByID uint) error {
	item, err := u.inventoryRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if item.VerifiedPhysical {
		return errors.New("item is already physically verified")
	}

	err = u.inventoryRepo.VerifyPhysical(ctx, id, verifiedByID)
	if err != nil {
		return err
	}

	// Check if all items in the parent donation are verified
	if item.DonationID != nil {
		donation, err := u.donationRepo.GetByID(ctx, *item.DonationID)
		if err == nil {
			allVerified := true
			for _, donationItem := range donation.Items {
				// If it's the current item we just verified, skip check
				if donationItem.ID == id {
					continue
				}
				if !donationItem.VerifiedPhysical {
					allVerified = false
					break
				}
			}

			if allVerified {
				_ = u.donationRepo.UpdateStatus(ctx, *item.DonationID, domain.StatusVerified)
			}
		}
	}

	return nil
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
