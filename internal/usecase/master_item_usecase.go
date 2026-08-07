package usecase

import (
	"clean-api/domain"
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
)

type masterItemUsecase struct {
	masterItemRepo domain.MasterItemRepository
}

func NewMasterItemUsecase(repo domain.MasterItemRepository) domain.MasterItemUsecase {
	return &masterItemUsecase{masterItemRepo: repo}
}

func (u *masterItemUsecase) Create(ctx context.Context, item *domain.MasterItem) (domain.MasterItem, error) {
	if item.Name == "" {
		return domain.MasterItem{}, errors.New("nama barang wajib diisi")
	}
	if item.SKUCode == "" {
		return domain.MasterItem{}, errors.New("kode SKU wajib diisi")
	}
	if item.Category == "" {
		item.Category = "Bahan Pokok & Sembako"
	}
	if item.Unit == "" {
		item.Unit = "Pcs"
	}

	err := u.masterItemRepo.Create(ctx, item)
	if err != nil {
		return domain.MasterItem{}, err
	}
	return *item, nil
}

func (u *masterItemUsecase) Update(ctx context.Context, id uint, item *domain.MasterItem) (domain.MasterItem, error) {
	existing, err := u.masterItemRepo.GetByID(ctx, id)
	if err != nil {
		return domain.MasterItem{}, errors.New("barang master tidak ditemukan")
	}

	if item.SKUCode != "" {
		existing.SKUCode = item.SKUCode
	}
	if item.Name != "" {
		existing.Name = item.Name
	}
	if item.Category != "" {
		existing.Category = item.Category
	}
	if item.Unit != "" {
		existing.Unit = item.Unit
	}
	existing.Description = item.Description

	err = u.masterItemRepo.Update(ctx, &existing)
	if err != nil {
		return domain.MasterItem{}, err
	}
	return existing, nil
}

func (u *masterItemUsecase) Delete(ctx context.Context, id uint) error {
	return u.masterItemRepo.Delete(ctx, id)
}

func (u *masterItemUsecase) DeleteAll(ctx context.Context) error {
	return u.masterItemRepo.DeleteAll(ctx)
}

func (u *masterItemUsecase) GetByID(ctx context.Context, id uint) (domain.MasterItem, error) {
	return u.masterItemRepo.GetByID(ctx, id)
}

func (u *masterItemUsecase) List(ctx context.Context) ([]domain.MasterItem, error) {
	return u.masterItemRepo.List(ctx)
}

func (u *masterItemUsecase) BulkCreate(ctx context.Context, items []domain.MasterItem) (domain.MasterItemBulkResult, error) {
	result := domain.MasterItemBulkResult{}

	existing, err := u.masterItemRepo.List(ctx)
	if err != nil {
		return result, err
	}
	usedSKUs := make(map[string]bool, len(existing))
	for _, e := range existing {
		usedSKUs[e.SKUCode] = true
	}

	for i, item := range items {
		name := strings.TrimSpace(item.Name)
		if name == "" {
			result.Skipped = append(result.Skipped, domain.MasterItemSkip{Name: item.Name, Reason: "Nama barang kosong"})
			continue
		}
		if item.Category == "" {
			result.Skipped = append(result.Skipped, domain.MasterItemSkip{Name: name, Reason: "Kategori belum dipilih"})
			continue
		}
		if _, err := u.masterItemRepo.GetByName(ctx, name); err == nil {
			result.Skipped = append(result.Skipped, domain.MasterItemSkip{Name: name, Reason: "Sudah ada di katalog"})
			continue
		}

		unit := item.Unit
		if unit == "" {
			unit = "Pcs"
		}

		// SKU comes from the imported file (e.g. "No Barang") when provided;
		// since that source data isn't guaranteed unique, collisions get a
		// numeric suffix rather than silently dropping the product.
		baseSKU := strings.TrimSpace(item.SKUCode)
		if baseSKU == "" {
			baseSKU = fmt.Sprintf("SKU-BULK-%d-%d", time.Now().UnixNano()%1000000, i)
		}
		sku := baseSKU
		for suffix := 2; usedSKUs[sku]; suffix++ {
			sku = fmt.Sprintf("%s-%d", baseSKU, suffix)
		}
		usedSKUs[sku] = true

		newItem := domain.MasterItem{
			SKUCode:     sku,
			Name:        name,
			Category:    item.Category,
			Unit:        unit,
			Price:       item.Price,
			Description: item.Description,
		}
		if err := u.masterItemRepo.Create(ctx, &newItem); err != nil {
			result.Skipped = append(result.Skipped, domain.MasterItemSkip{Name: name, Reason: "Gagal disimpan: " + err.Error()})
			continue
		}
		result.Created = append(result.Created, newItem)
	}

	return result, nil
}
