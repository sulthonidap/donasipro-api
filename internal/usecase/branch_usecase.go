package usecase

import (
	"clean-api/domain"
	"context"
	"errors"
)

type branchUsecase struct {
	branchRepo domain.BranchRepository
}

func NewBranchUsecase(branchRepo domain.BranchRepository) domain.BranchUsecase {
	return &branchUsecase{branchRepo: branchRepo}
}

func (u *branchUsecase) CreateBranch(ctx context.Context, name, address, phone string) (domain.Branch, error) {
	if name == "" {
		return domain.Branch{}, errors.New("nama cabang wajib diisi")
	}

	branch := domain.Branch{
		Name:    name,
		Address: address,
		Phone:   phone,
	}

	err := u.branchRepo.Create(ctx, &branch)
	return branch, err
}

func (u *branchUsecase) GetByID(ctx context.Context, id uint) (domain.Branch, error) {
	return u.branchRepo.GetByID(ctx, id)
}

func (u *branchUsecase) ListBranches(ctx context.Context) ([]domain.Branch, error) {
	return u.branchRepo.ListAll(ctx)
}

func (u *branchUsecase) UpdateBranch(ctx context.Context, id uint, name, address, phone string) (domain.Branch, error) {
	branch, err := u.branchRepo.GetByID(ctx, id)
	if err != nil {
		return domain.Branch{}, errors.New("cabang tidak ditemukan")
	}

	if name != "" {
		branch.Name = name
	}
	branch.Address = address
	branch.Phone = phone

	err = u.branchRepo.Update(ctx, &branch)
	return branch, err
}

func (u *branchUsecase) DeleteBranch(ctx context.Context, id uint) error {
	return u.branchRepo.Delete(ctx, id)
}
