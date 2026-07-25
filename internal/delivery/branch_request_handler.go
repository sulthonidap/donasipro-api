package delivery

import (
	"clean-api/domain"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type BranchRequestHandler struct {
	reqUsecase domain.BranchRequestUsecase
}

func NewBranchRequestHandler(ru domain.BranchRequestUsecase) *BranchRequestHandler {
	return &BranchRequestHandler{reqUsecase: ru}
}

type CreateBranchReqInput struct {
	ItemName       string                   `json:"item_name" binding:"required"`
	Category       domain.InventoryCategory `json:"category" binding:"required"`
	RoutineQuota   float64                  `json:"routine_quota"`
	RemainingStock float64                  `json:"remaining_stock"`
	Quantity       float64                  `json:"quantity" binding:"required"`
	Unit           string                   `json:"unit" binding:"required"`
	MonthPeriod    string                   `json:"month_period"`
	ApplicantName  string                   `json:"applicant_name"`
	Purpose        string                   `json:"purpose"`
}

type ApproveBranchReqInput struct {
	CourierID    uint   `json:"courier_id" binding:"required"`
	ApproverName string `json:"approver_name"`
}

func (h *BranchRequestHandler) Create(c *gin.Context) {
	var input CreateBranchReqInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userIDVal, exists := c.Get("user_id")
	if !exists {
		userIDVal, exists = c.Get("userID")
	}
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var requestedBy uint
	switch v := userIDVal.(type) {
	case float64:
		requestedBy = uint(v)
	case uint:
		requestedBy = v
	case int:
		requestedBy = uint(v)
	}

	// Retrieve user's branch_id if available, or allow explicit branch_id
	var branchID uint = 1
	if bVal, ok := c.Get("branchID"); ok && bVal != nil {
		if bUint, ok := bVal.(*uint); ok && bUint != nil {
			branchID = *bUint
		}
	}

	req, err := h.reqUsecase.CreateRequest(
		c.Request.Context(),
		branchID,
		requestedBy,
		input.ItemName,
		input.Category,
		input.RoutineQuota,
		input.RemainingStock,
		input.Quantity,
		input.Unit,
		input.MonthPeriod,
		input.ApplicantName,
		input.Purpose,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, req)
}

func (h *BranchRequestHandler) List(c *gin.Context) {
	var branchIDPtr *uint
	if bStr := c.Query("branch_id"); bStr != "" {
		if bId, err := strconv.ParseUint(bStr, 10, 32); err == nil && bId > 0 {
			bVal := uint(bId)
			branchIDPtr = &bVal
		}
	}

	requests, err := h.reqUsecase.ListRequests(c.Request.Context(), branchIDPtr)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, requests)
}

func (h *BranchRequestHandler) Approve(c *gin.Context) {
	idParam := c.Param("id")
	reqID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID"})
		return
	}

	var input ApproveBranchReqInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Kurir wajib dipilih untuk pengiriman"})
		return
	}

	err = h.reqUsecase.ApproveAndAssignCourier(c.Request.Context(), uint(reqID), input.CourierID, input.ApproverName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pengajuan disetujui & tugas kurir berhasil dibuat"})
}

func (h *BranchRequestHandler) Reject(c *gin.Context) {
	idParam := c.Param("id")
	reqID, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request ID"})
		return
	}

	err = h.reqUsecase.RejectRequest(c.Request.Context(), uint(reqID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Pengajuan cabang ditolak"})
}
