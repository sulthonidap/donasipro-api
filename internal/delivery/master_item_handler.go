package delivery

import (
	"clean-api/domain"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type MasterItemHandler struct {
	masterItemUsecase domain.MasterItemUsecase
}

func NewMasterItemHandler(u domain.MasterItemUsecase) *MasterItemHandler {
	return &MasterItemHandler{masterItemUsecase: u}
}

func (h *MasterItemHandler) List(c *gin.Context) {
	items, err := h.masterItemUsecase.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

func (h *MasterItemHandler) Create(c *gin.Context) {
	var input domain.MasterItem
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	created, err := h.masterItemUsecase.Create(c.Request.Context(), &input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, created)
}

type BulkCreateMasterItemInput struct {
	Items []domain.MasterItem `json:"items"`
}

func (h *MasterItemHandler) BulkCreate(c *gin.Context) {
	var input BulkCreateMasterItemInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(input.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tidak ada barang untuk diimport"})
		return
	}

	result, err := h.masterItemUsecase.BulkCreate(c.Request.Context(), input.Items)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *MasterItemHandler) Update(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	var input domain.MasterItem
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	updated, err := h.masterItemUsecase.Update(c.Request.Context(), uint(id), &input)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}

func (h *MasterItemHandler) Delete(c *gin.Context) {
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ID tidak valid"})
		return
	}

	err = h.masterItemUsecase.Delete(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Barang master berhasil dihapus"})
}
