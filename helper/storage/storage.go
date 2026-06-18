package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"time"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"

	"github.com/yerobalg/wealthpulse-service/helper/cryptolib"
	"github.com/yerobalg/wealthpulse-service/helper/logger"
)

type Interface interface {
	Upload(ctx context.Context, param UploadParam) (string, error)
	Delete(ctx context.Context, objectName string) error
	GetObjectSignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error)
	GetUploadSignedURL(ctx context.Context, objectName, contentType string, expiry time.Duration) (string, error)
	ObjectExists(ctx context.Context, objectName string) (bool, error)
}

type UploadParam struct {
	ObjectName  string
	File        io.Reader
	ContentType string
}

type Storage struct {
	client     *storage.Client
	bucketName string
	log        logger.Interface
}

type InitParam struct {
	EncryptedCredentialJSON string
	EncryptionKey           string
	BucketName              string
	Log                     logger.Interface
}

func Init(ctx context.Context, param InitParam) (*Storage, error) {
	credentialJSON, err := cryptolib.Decrypt([]byte(param.EncryptionKey), param.EncryptedCredentialJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt GCS credentials: %w", err)
	}

	client, err := storage.NewClient(ctx, option.WithCredentialsJSON(credentialJSON))
	if err != nil {
		return nil, fmt.Errorf("failed to create GCS client: %w", err)
	}

	return &Storage{
		client:     client,
		bucketName: param.BucketName,
		log:        param.Log,
	}, nil
}

func (s *Storage) Upload(ctx context.Context, param UploadParam) (string, error) {
	writer := s.client.Bucket(s.bucketName).Object(param.ObjectName).NewWriter(ctx)
	writer.ContentType = param.ContentType

	if _, err := io.Copy(writer, param.File); err != nil {
		return "", fmt.Errorf("failed to copy file to GCS: %w", err)
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close GCS writer: %w", err)
	}

	url := fmt.Sprintf("https://storage.googleapis.com/%s/%s", s.bucketName, param.ObjectName)

	s.log.Info(ctx, fmt.Sprintf("uploaded object %s to bucket %s", param.ObjectName, s.bucketName))

	return url, nil
}

func (s *Storage) Delete(ctx context.Context, objectName string) error {
	if err := s.client.Bucket(s.bucketName).Object(objectName).Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete object %s: %w", objectName, err)
	}

	s.log.Info(ctx, fmt.Sprintf("deleted object %s from bucket %s", objectName, s.bucketName))

	return nil
}

func (s *Storage) GetObjectSignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	url, err := s.client.Bucket(s.bucketName).SignedURL(objectName, &storage.SignedURLOptions{
		Method:  "GET",
		Expires: time.Now().Add(expiry),
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate object signed URL: %w", err)
	}

	return url, nil
}

func (s *Storage) GetUploadSignedURL(ctx context.Context, objectName, contentType string, expiry time.Duration) (string, error) {
	url, err := s.client.Bucket(s.bucketName).SignedURL(objectName, &storage.SignedURLOptions{
		Method:      "PUT",
		Expires:     time.Now().Add(expiry),
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate upload signed URL: %w", err)
	}

	return url, nil
}

func (s *Storage) ObjectExists(ctx context.Context, objectName string) (bool, error) {
	_, err := s.client.Bucket(s.bucketName).Object(objectName).Attrs(ctx)
	if errors.Is(err, storage.ErrObjectNotExist) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("failed to check object existence: %w", err)
	}
	return true, nil
}

func (s *Storage) Close() error {
	return s.client.Close()
}
