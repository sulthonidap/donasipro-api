package usecase

import (
	"clean-api/domain"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

type donationUsecase struct {
	donationRepo   domain.DonationRepository
	masterItemRepo domain.MasterItemRepository
}

func NewDonationUsecase(donationRepo domain.DonationRepository, masterItemRepo domain.MasterItemRepository) domain.DonationUsecase {
	return &donationUsecase{
		donationRepo:   donationRepo,
		masterItemRepo: masterItemRepo,
	}
}

func (u *donationUsecase) Submit(ctx context.Context, donation *domain.Donation) (domain.Donation, error) {
	if donation.DonorName == "" {
		return domain.Donation{}, errors.New("donor name is required")
	}

	donation.Status = domain.StatusPending

	if donation.DonationType == domain.TypeGoods {
		if len(donation.Items) == 0 {
			return domain.Donation{}, errors.New("goods donation must contain at least one item")
		}

		var descParts []string
		for i := range donation.Items {
			donation.Items[i].VerifiedPhysical = false
			donation.Items[i].DeliveryStatus = domain.DeliveryUnassigned
			descParts = append(descParts, fmt.Sprintf("%.1f %s %s", donation.Items[i].Quantity, donation.Items[i].Unit, donation.Items[i].ItemName))

			// Auto-register to Master Item Catalog if item name is not in catalog yet
			if u.masterItemRepo != nil && donation.Items[i].ItemName != "" {
				_, err := u.masterItemRepo.GetByName(ctx, donation.Items[i].ItemName)
				if err != nil {
					autoSKU := fmt.Sprintf("SKU-AUTO-%d", time.Now().UnixNano()%1000000)
					cat := string(donation.Items[i].Category)
					if cat == "" {
						cat = "Bahan Pokok & Sembako"
					}
					_ = u.masterItemRepo.Create(ctx, &domain.MasterItem{
						SKUCode:     autoSKU,
						Name:        donation.Items[i].ItemName,
						Category:    cat,
						Unit:        donation.Items[i].Unit,
						Description: "Otomatis terdaftar dari input donasi barang",
					})
				}
			}
		}
		donation.ItemsDesc = strings.Join(descParts, ", ")
		donation.Amount = 0
	} else if donation.DonationType == domain.TypeFunds {
		if donation.Amount <= 0 {
			return domain.Donation{}, errors.New("funds donation must have an amount greater than 0")
		}
		donation.Items = nil
		donation.ItemsDesc = fmt.Sprintf("Donasi Dana: Rp. %.0f", donation.Amount)
	} else {
		return domain.Donation{}, errors.New("invalid donation type")
	}

	err := u.donationRepo.Create(ctx, donation)
	if err != nil {
		return domain.Donation{}, err
	}

	return *donation, nil
}

func (u *donationUsecase) GetByID(ctx context.Context, id uint) (domain.Donation, error) {
	return u.donationRepo.GetByID(ctx, id)
}

func (u *donationUsecase) List(ctx context.Context) ([]domain.Donation, error) {
	return u.donationRepo.List(ctx)
}

func (u *donationUsecase) VerifyFund(ctx context.Context, id uint) error {
	donation, err := u.donationRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if donation.DonationType != domain.TypeFunds {
		return errors.New("only funds donations can be verified by finance")
	}

	return u.donationRepo.UpdateStatus(ctx, id, domain.StatusVerified)
}
