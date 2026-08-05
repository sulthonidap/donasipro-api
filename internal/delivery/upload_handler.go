package delivery

import (
	"clean-api/internal/storage"
	"fmt"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type UploadHandler struct {
	storage *storage.MinioStorage
}

func NewUploadHandler(s *storage.MinioStorage) *UploadHandler {
	return &UploadHandler{storage: s}
}

type PresignUploadInput struct {
	Filename string `json:"filename" binding:"required"`
}

var unsafeFilenameChars = regexp.MustCompile(`[^a-zA-Z0-9._-]`)

// PresignDonationPhoto issues a short-lived MinIO upload URL for a donation
// item photo. The browser uploads the file bytes directly to MinIO using the
// returned URL — this backend never touches the file itself.
func (h *UploadHandler) PresignDonationPhoto(c *gin.Context) {
	if h.storage == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Layanan penyimpanan foto belum dikonfigurasi"})
		return
	}

	var input PresignUploadInput
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ext := filepath.Ext(input.Filename)
	base := strings.TrimSuffix(filepath.Base(input.Filename), ext)
	base = unsafeFilenameChars.ReplaceAllString(base, "_")
	if base == "" {
		base = "foto"
	}
	objectKey := fmt.Sprintf("%s/%d-%s%s", storage.DonationPhotoPrefix(), time.Now().UnixNano(), base, ext)

	uploadURL, publicURL, err := h.storage.PresignUpload(c.Request.Context(), objectKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Gagal menyiapkan upload foto: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"upload_url": uploadURL,
		"photo_url":  publicURL,
	})
}
