package delivery

import (
	"clean-api/domain"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type DeliveryHandler struct {
	deliveryUsecase domain.DeliveryUsecase
}

func NewDeliveryHandler(du domain.DeliveryUsecase) *DeliveryHandler {
	return &DeliveryHandler{deliveryUsecase: du}
}

type AssignDeliveryInput struct {
	InventoryID      uint   `json:"inventory_id" binding:"required"`
	CourierID        uint   `json:"courier_id" binding:"required"`
	RecipientName    string `json:"recipient_name" binding:"required"`
	RecipientAddress string `json:"recipient_address" binding:"required"`
	RecipientPhone   string `json:"recipient_phone" binding:"required"`
}

func (h *DeliveryHandler) Create(c *gin.Context) {
	var input AssignDeliveryInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	delivery := domain.Delivery{
		InventoryID:      input.InventoryID,
		CourierID:        input.CourierID,
		RecipientName:    input.RecipientName,
		RecipientAddress: input.RecipientAddress,
		RecipientPhone:   input.RecipientPhone,
	}

	res, err := h.deliveryUsecase.Create(c.Request.Context(), &delivery)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, res)
}

func (h *DeliveryHandler) ListAll(c *gin.Context) {
	deliveries, err := h.deliveryUsecase.ListAll(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, deliveries)
}

func (h *DeliveryHandler) ListForCourier(c *gin.Context) {
	// Parse user_id from context
	val, _ := c.Get("user_id")
	var courierID uint
	switch v := val.(type) {
	case float64:
		courierID = uint(v)
	case uint:
		courierID = v
	}

	deliveries, err := h.deliveryUsecase.ListForCourier(c.Request.Context(), courierID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, deliveries)
}

func (h *DeliveryHandler) StartDelivery(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid delivery ID"})
		return
	}

	err = h.deliveryUsecase.StartDelivery(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Delivery started"})
}

func (h *DeliveryHandler) UploadProof(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid delivery ID"})
		return
	}

	// Proof photo is optional — if no file provided, just mark as delivered
	file, fileErr := c.FormFile("proof")
	var proofPath string

	if fileErr == nil && file != nil {
		uploadsDir := "uploads"
		if _, err := os.Stat(uploadsDir); os.IsNotExist(err) {
			os.Mkdir(uploadsDir, 0755)
		}
		filename := strconv.FormatInt(time.Now().UnixNano(), 10) + filepath.Ext(file.Filename)
		dst := filepath.Join(uploadsDir, filename)
		if err := c.SaveUploadedFile(file, dst); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
			return
		}
		proofPath = "/uploads/" + filename
	}

	if err := h.deliveryUsecase.UploadProof(c.Request.Context(), uint(id), proofPath); err != nil {
		if proofPath != "" {
			os.Remove(filepath.Join("uploads", filepath.Base(proofPath)))
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Delivery completed successfully",
		"proof_photo": proofPath,
	})
}

func (h *DeliveryHandler) CompleteWithoutProof(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid delivery ID"})
		return
	}

	if err := h.deliveryUsecase.UploadProof(c.Request.Context(), uint(id), ""); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Delivery completed without proof photo"})
}
