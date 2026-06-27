package storage

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioConfig struct {
	Endpoint   string
	AccessKey  string
	SecretKey  string
	BucketName string
	UseSSL     bool
}

type minioStorage struct {
	client     *minio.Client
	bucketName string
}

func NewMinioStorage(cfg MinioConfig) (Storage, error) {
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("Failed to initialize minio client: %w", err)
	}

	storage := &minioStorage{
		client:     minioClient,
		bucketName: cfg.BucketName,
	}

	ctx := context.Background()
	exists, err := minioClient.BucketExists(ctx, cfg.BucketName)
	if err != nil {
		return nil, fmt.Errorf("failed to check if bucket exists: %w", err)
	}

	if !exists {
		err = minioClient.MakeBucket(ctx, cfg.BucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create minio bucket: %w", err)
		}
	}
	return storage, nil
}

func (m *minioStorage) Upload(ctx context.Context, key string, reader io.Reader, contentType string) error {
	_, err := m.client.PutObject(ctx, m.bucketName, key, reader, -1, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return fmt.Errorf("minio upload error for key %s: %w", key, err)
	}
	return nil
}

func (m *minioStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	object, err := m.client.GetObject(ctx, m.bucketName, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("minio download error for key %s: %w", key, err)
	}

	_, err = object.Stat()
	if err != nil {
		object.Close()
		return nil, fmt.Errorf("failed to find object %s: %w", key, err)
	}

	return object, nil
}

func (m *minioStorage) Delete(ctx context.Context, key string) error {
	err := m.client.RemoveObject(ctx, m.bucketName, key, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete minio object %s: %w", key, err)
	}
	return nil
}
