package storage

import (
	"context"
	"fmt"
	"log"
	"mime/multipart"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Storage struct {
	client    *minio.Client
	bucket    string
	publicURL string
}

// NewStorage — подключается к MinIO
func NewStorage(endpoint, accessKey, secretKey, bucket string, useSSL bool, region string, publicURL string) (*Storage, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
		Region: region,
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
		client:    client,
		bucket:    bucket,
		publicURL: publicURL,
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

// PresignedPostData — то, что отдадим фронту
type PresignedPostData struct {
	Url    string            `json:"url"`
	Fields map[string]string `json:"fields"`
}

func (s *Storage) GeneratePresignedPost(ctx context.Context, objectName string, expires time.Duration) (*PresignedPostData, error) {
	// 1. Создаём политику
	policy := minio.NewPostPolicy()

	// 2. Куда грузить
	_ = policy.SetBucket(s.bucket)
	_ = policy.SetKey(objectName)

	// 3. Время жизни (например, 5 минут)
	_ = policy.SetExpires(time.Now().Add(expires))

	// 4. ЖЁСТКИЕ ЛИМИТЫ (безопасность)
	// Разрешаем только картинки
	_ = policy.SetContentTypeStartsWith("image/")
	// Размер от 1 байта до 5 MB
	_ = policy.SetContentLengthRange(1, 5*1024*1024)

	// 5. Генерируем URL и поля формы
	url, formData, err := s.client.PresignedPostPolicy(ctx, policy)
	if err != nil {
		return nil, fmt.Errorf("presigned post: %w", err)
	}

	internalHost := s.client.EndpointURL().Host

	//  Собираем внутренний URL, который сгенерировал MinIO
	internalURL := url.String()

	publicHost := strings.TrimPrefix(s.publicURL, "http://")
	publicHost = strings.TrimPrefix(publicHost, "https://")

	// Делаем подмену хоста в строке
	publicURLStr := strings.Replace(internalURL, internalHost, publicHost, 1)

	return &PresignedPostData{
		Url:    publicURLStr,
		Fields: formData,
	}, nil
}

func (s *Storage) StatObject(ctx context.Context, objectName string) (minio.ObjectInfo, error) {
	return s.client.StatObject(ctx, s.bucket, objectName, minio.StatObjectOptions{})
}

func (s *Storage) DeleteObject(ctx context.Context, objectName string) error {
	return s.client.RemoveObject(ctx, s.bucket, objectName, minio.RemoveObjectOptions{})
}

// Для формирования URL (если нужно)
func (s *Storage) GetEndpoint() string { return s.client.EndpointURL().Host }
func (s *Storage) GetBucket() string   { return s.bucket }
