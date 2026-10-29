package storage

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Storage struct {
	client *minio.Client
	bucket string
}

// NewStorage — подключается к MinIO
func NewStorage(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*Storage, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к MinIO: %w", err)
	}

	// Проверяем, существует ли бакет (если нет — создаем)
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("ошибка проверки бакета: %w", err)
	}
	if !exists {
		err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("ошибка создания бакета: %w", err)
		}
		log.Println("✅ Бакет", bucket, "создан")
	}

	return &Storage{
		client: client,
		bucket: bucket,
	}, nil
}

// UploadFile — загружает файл в MinIO и возвращает его URL
func (s *Storage) UploadFile(ctx context.Context, file multipart.File, filename string, size int64) (string, error) {
	// Загружаем файл
	_, err := s.client.PutObject(ctx, s.bucket, filename, file, size, minio.PutObjectOptions{})
	if err != nil {
		return "", fmt.Errorf("ошибка загрузки файла: %w", err)
	}

	// Возвращаем URL файла (публичную ссылку)
	// Пока делаем бакет публичным для чтения, чтобы фронт мог видеть картинку
	url := fmt.Sprintf("http://%s/%s/%s", s.client.EndpointURL().Host, s.bucket, filename)
	return url, nil
}
