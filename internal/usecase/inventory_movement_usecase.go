package usecase

import (
	"clean-api/domain"
	"context"
)

type inventoryMovementUsecase struct {
	movementRepo domain.InventoryMovementRepository
}

func NewInventoryMovementUsecase(movementRepo domain.InventoryMovementRepository) domain.InventoryMovementUsecase {
	return &inventoryMovementUsecase{movementRepo: movementRepo}
}

func (u *inventoryMovementUsecase) List(ctx context.Context, filter domain.InventoryMovementFilter) ([]domain.InventoryMovement, error) {
	return u.movementRepo.List(ctx, filter)
}
