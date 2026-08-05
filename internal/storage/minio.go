package storage

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// donationPhotoPrefix is the only prefix given a public-read bucket policy —
// everything else in the bucket stays private by default.
const donationPhotoPrefix = "donations"

type MinioStorage struct {
	client     *minio.Client
	bucket     string
	publicBase string
}

func NewMinioStorage() (*MinioStorage, error) {
	endpoint := os.Getenv("MINIO_ENDPOINT")
	accessKey := os.Getenv("MINIO_ACCESS_KEY")
	secretKey := os.Getenv("MINIO_SECRET_KEY")
	bucket := os.Getenv("MINIO_BUCKET")
	useSSL := os.Getenv("MINIO_USE_SSL") == "true"

	if endpoint == "" || accessKey == "" || secretKey == "" || bucket == "" {
		return nil, fmt.Errorf("MinIO belum dikonfigurasi (MINIO_ENDPOINT/MINIO_ACCESS_KEY/MINIO_SECRET_KEY/MINIO_BUCKET)")
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	scheme := "http"
	if useSSL {
		scheme = "https"
	}

	s := &MinioStorage{
		client:     client,
		bucket:     bucket,
		publicBase: fmt.Sprintf("%s://%s/%s", scheme, endpoint, bucket),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.ensureBucket(ctx); err != nil {
		log.Println("Peringatan: gagal menyiapkan bucket MinIO:", err)
	}

	return s, nil
}

func (s *MinioStorage) ensureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return err
	}
	if !exists {
		if err := s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
			return err
		}
	}

	policy := fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [{
			"Effect": "Allow",
			"Principal": {"AWS": ["*"]},
			"Action": ["s3:GetObject"],
			"Resource": ["arn:aws:s3:::%s/%s/*"]
		}]
	}`, s.bucket, donationPhotoPrefix)
	return s.client.SetBucketPolicy(ctx, s.bucket, policy)
}

// PresignUpload returns a short-lived signed PUT URL the browser can upload
// directly to, plus the public URL the object will be reachable at afterwards.
func (s *MinioStorage) PresignUpload(ctx context.Context, objectKey string) (uploadURL string, publicURL string, err error) {
	u, err := s.client.PresignedPutObject(ctx, s.bucket, objectKey, 10*time.Minute)
	if err != nil {
		return "", "", err
	}
	return u.String(), fmt.Sprintf("%s/%s", s.publicBase, objectKey), nil
}

// DonationPhotoPrefix exposes the prefix so handlers can build object keys
// consistently with the bucket policy set up above.
func DonationPhotoPrefix() string {
	return donationPhotoPrefix
}
