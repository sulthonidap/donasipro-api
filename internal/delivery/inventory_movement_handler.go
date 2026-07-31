package delivery

import (
	"clean-api/domain"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type InventoryMovementHandler struct {
	movementUsecase domain.InventoryMovementUsecase
}

func NewInventoryMovementHandler(u domain.InventoryMovementUsecase) *InventoryMovementHandler {
	return &InventoryMovementHandler{movementUsecase: u}
}

func (h *InventoryMovementHandler) List(c *gin.Context) {
	filter := domain.InventoryMovementFilter{
		ItemName:  c.Query("item_name"),
		Direction: domain.MovementDirection(c.Query("direction")),
		Source:    domain.MovementSource(c.Query("source")),
	}

	if bStr := c.Query("branch_id"); bStr != "" {
		if bID, err := strconv.ParseUint(bStr, 10, 32); err == nil && bID > 0 {
			bVal := uint(bID)
			filter.BranchID = &bVal
		}
	}
	if dStr := c.Query("date_from"); dStr != "" {
		if d, err := time.Parse("2006-01-02", dStr); err == nil {
			filter.DateFrom = &d
		}
	}
	if dStr := c.Query("date_to"); dStr != "" {
		if d, err := time.Parse("2006-01-02", dStr); err == nil {
			endOfDay := d.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			filter.DateTo = &endOfDay
		}
	}

	movements, err := h.movementUsecase.List(c.Request.Context(), filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, movements)
}
