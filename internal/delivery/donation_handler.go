package delivery

import (
	"clean-api/domain"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type DonationHandler struct {
	donationUsecase domain.DonationUsecase
}

func NewDonationHandler(du domain.DonationUsecase) *DonationHandler {
	return &DonationHandler{donationUsecase: du}
}

type DonationItemInput struct {
	ItemName   string                   `json:"item_name" binding:"required"`
	Category   domain.InventoryCategory `json:"category" binding:"required"`
	Quantity   float64                  `json:"quantity" binding:"required"`
	Unit       string                   `json:"unit" binding:"required"`
	ExpiryDate *string                  `json:"expiry_date"` // donor-claimed, still re-verified by logistics
	PhotoURL   string                   `json:"photo_url"`
}

type SubmitDonationInput struct {
	DonorName     string              `json:"donor_name" binding:"required"`
	DonorPhone    string              `json:"donor_phone" binding:"required"`
	DonorAddress  string              `json:"donor_address" binding:"required"`
	DonationType  domain.DonationType `json:"donation_type" binding:"required"`
	Amount        float64             `json:"amount"`          // for funds
	PaymentMethod domain.PaymentMethod `json:"payment_method"` // cash | transfer | va_bca
	Items         []DonationItemInput  `json:"items"`          // for goods
}

func (h *DonationHandler) Submit(c *gin.Context) {
	var input SubmitDonationInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse receiver ID from token context
	val, _ := c.Get("user_id")
	var receiverID uint
	switch v := val.(type) {
	case float64:
		receiverID = uint(v)
	case uint:
		receiverID = v
	}

	donation := domain.Donation{
		DonorName:     input.DonorName,
		DonorPhone:    input.DonorPhone,
		DonorAddress:  input.DonorAddress,
		DonationType:  input.DonationType,
		Amount:        input.Amount,
		PaymentMethod: input.PaymentMethod,
		ReceiverID:    receiverID,
	}

	if input.DonationType == domain.TypeGoods {
		donation.Items = make([]domain.Inventory, len(input.Items))
		for i, item := range input.Items {
			donation.Items[i] = domain.Inventory{
				ItemName:   item.ItemName,
				Category:   item.Category,
				Quantity:   item.Quantity,
				Unit:       item.Unit,
				ExpiryDate: parseExpiryDate(item.ExpiryDate),
				PhotoURL:   item.PhotoURL,
			}
		}
	}

	res, err := h.donationUsecase.Submit(c.Request.Context(), &donation)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, res)
}

func (h *DonationHandler) List(c *gin.Context) {
	donations, err := h.donationUsecase.List(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, donations)
}

func (h *DonationHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid donation ID"})
		return
	}

	donation, err := h.donationUsecase.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Donation not found"})
		return
	}

	c.JSON(http.StatusOK, donation)
}

func (h *DonationHandler) Print(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid donation ID"})
		return
	}

	donation, err := h.donationUsecase.GetByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Donation not found"})
		return
	}

	// Returns HTML formatted receipt or JSON formatted metadata
	// The client blueprint says: "GET /api/v1/donations/:id/print - Mengambil data terformat untuk kebutuhan cetak nota"
	// So we return JSON representing the invoice/receipt content, and the frontend React DocumentPrinter component will render it beautifully for print.
	c.JSON(http.StatusOK, gin.H{
		"receipt_no":   "RCP-" + strconv.FormatUint(id, 10),
		"print_date":   donation.CreatedAt.Format("2006-01-02 15:04:05"),
		"donation":     donation,
		"signature_by": "Petugas Penerima Donasi",
	})
}

func (h *DonationHandler) VerifyFund(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid donation ID"})
		return
	}

	err = h.donationUsecase.VerifyFund(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Fund donation verified successfully"})
}
