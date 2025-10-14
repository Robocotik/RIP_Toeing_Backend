package repository

import (
	"fmt"
	"mime/multipart"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"context"
)

func (r *Repository) UploadRumbImageToMinio(rumbID int, file *multipart.FileHeader, fileName string) error {
	ctx := context.Background()

	// Настройка MinIO клиента
	minioClient, err := minio.New("localhost:9000", &minio.Options{
		Creds:  credentials.NewStaticV4("admin", "admin12345", ""),
		Secure: false,
	})
	if err != nil {
		return err
	}

	bucketName := "rumbs"

	// Проверка и создание бакета
	_, err = minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		return err
	}

	// Удаляем старое изображение, если есть
	rumb, err := r.GetRumbByID(rumbID)
	if err != nil {
		return err
	}
	if rumb.Image != "" {
		_ = minioClient.RemoveObject(ctx, bucketName, rumb.Image, minio.RemoveObjectOptions{})
	}

	// Загружаем новый файл с оригинальным именем
	srcFile, err := file.Open()
	if err != nil {
		return err
	}
	defer srcFile.Close()

	_, err = minioClient.PutObject(ctx, bucketName, fileName, srcFile, file.Size, minio.PutObjectOptions{
		ContentType: file.Header.Get("Content-Type"),
	})
	if err != nil {
		return fmt.Errorf("failed to upload to minio: %v", err)
	}

	return nil
}
