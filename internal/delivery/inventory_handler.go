package delivery

import (
	"clean-api/domain"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type InventoryHandler struct {
	inventoryUsecase domain.InventoryUsecase
}

func NewInventoryHandler(iu domain.InventoryUsecase) *InventoryHandler {
	return &InventoryHandler{inventoryUsecase: iu}
}

func (h *InventoryHandler) List(c *gin.Context) {
	category := c.Query("category") // Bahan Pokok or Makanan Anak
	verifiedOnlyStr := c.Query("verified_only")

	verifiedOnly := false
	if verifiedOnlyStr == "true" {
		verifiedOnly = true
	}

	items, err := h.inventoryUsecase.List(c.Request.Context(), category, verifiedOnly)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, items)
}

type VerifyPhysicalInput struct {
	ExpiryDate *string `json:"expiry_date"`
	PhotoURL   *string `json:"photo_url"`
}

func parseExpiryDate(raw *string) *time.Time {
	if raw == nil || *raw == "" {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339, *raw)
	if err != nil {
		parsed, err = time.Parse("2006-01-02", *raw)
	}
	if err != nil {
		return nil
	}
	return &parsed
}

func userIDFromContext(c *gin.Context) uint {
	val, _ := c.Get("user_id")
	switch v := val.(type) {
	case float64:
		return uint(v)
	case uint:
		return v
	}
	return 0
}

func (h *InventoryHandler) VerifyPhysical(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid inventory ID"})
		return
	}

	var input VerifyPhysicalInput
	_ = c.ShouldBindJSON(&input)

	expTime := parseExpiryDate(input.ExpiryDate)
	verifiedByID := userIDFromContext(c)

	err = h.inventoryUsecase.VerifyPhysical(c.Request.Context(), uint(id), verifiedByID, expTime, input.PhotoURL)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Item verified successfully"})
}

type VerifyPhysicalSplitInput struct {
	Batches []struct {
		Quantity   float64 `json:"quantity" binding:"required"`
		ExpiryDate *string `json:"expiry_date"`
		PhotoURL   string  `json:"photo_url"`
	} `json:"batches" binding:"required,min=1"`
}

func (h *InventoryHandler) VerifyPhysicalSplit(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid inventory ID"})
		return
	}

	var input VerifyPhysicalSplitInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak valid: " + err.Error()})
		return
	}

	batches := make([]domain.InventoryBatch, len(input.Batches))
	for i, b := range input.Batches {
		batches[i] = domain.InventoryBatch{
			Quantity:   b.Quantity,
			ExpiryDate: parseExpiryDate(b.ExpiryDate),
			PhotoURL:   b.PhotoURL,
		}
	}

	result, err := h.inventoryUsecase.VerifyPhysicalSplit(c.Request.Context(), uint(id), userIDFromContext(c), batches)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

type CreateInventoryInput struct {
	ItemName   string                   `json:"item_name" binding:"required"`
	Category   domain.InventoryCategory `json:"category"`
	Quantity   float64                  `json:"quantity"`
	Unit       string                   `json:"unit"`
	BranchID   *uint                    `json:"branch_id"`
	ExpiryDate *string                  `json:"expiry_date"`
}

func (h *InventoryHandler) CreateDirectly(c *gin.Context) {
	var input CreateInventoryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Data tidak valid: " + err.Error()})
		return
	}

	if input.Quantity <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Kuantitas barang harus lebih dari 0"})
		return
	}

	if input.Category == "" {
		input.Category = "Bahan Pokok & Sembako"
	}
	if input.Unit == "" {
		input.Unit = "Pcs"
	}

	var expTime *time.Time
	if input.ExpiryDate != nil && *input.ExpiryDate != "" {
		parsed, err := time.Parse(time.RFC3339, *input.ExpiryDate)
		if err != nil {
			parsed, err = time.Parse("2006-01-02", *input.ExpiryDate)
		}
		if err == nil {
			expTime = &parsed
		}
	}

	item := domain.Inventory{
		ItemName:   input.ItemName,
		Category:   input.Category,
		Quantity:   input.Quantity,
		Unit:       input.Unit,
		BranchID:   input.BranchID,
		ExpiryDate: expTime,
	}

	res, err := h.inventoryUsecase.CreateItemDirectly(c.Request.Context(), &item, userIDFromContext(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, res)
}
