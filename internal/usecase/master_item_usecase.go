package usecase

import (
	"clean-api/domain"
	"context"
	"errors"
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

func (u *masterItemUsecase) GetByID(ctx context.Context, id uint) (domain.MasterItem, error) {
	return u.masterItemRepo.GetByID(ctx, id)
}

func (u *masterItemUsecase) List(ctx context.Context) ([]domain.MasterItem, error) {
	return u.masterItemRepo.List(ctx)
}
