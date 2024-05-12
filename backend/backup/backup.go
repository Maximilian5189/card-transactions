package backup

import (
	"backend/logger"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

type Backup struct {
	accessKeyID     string
	secretAccessKey string
	bucket          string
	logger          logger.Logger
	sess            *session.Session
}

func New(logger logger.Logger) (Backup, error) {
	accessKeyID := "4baf8cecf0035daea9f6254c896ef9a4"
	secretAccessKey := "87042dac0e01e62a44dbd0b4e4398e3e4ded1a40245982791664f3b24f0cf597"

	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String("ENAM"),
		Credentials:      credentials.NewStaticCredentials(accessKeyID, secretAccessKey, ""),
		Endpoint:         aws.String("https://a4c4c2120db7203352114a34675017e9.r2.cloudflarestorage.com"),
		S3ForcePathStyle: aws.Bool(true),
	})
	if err != nil {
		return Backup{}, err
	}

	bucket := "transactions-backup"
	return Backup{accessKeyID, secretAccessKey, bucket, logger, sess}, nil
}

func (b *Backup) Upload(filepath string) {
	file, err := os.Open(filepath)
	if err != nil {
		b.logger.Error(fmt.Sprintf("Failed to open file %q, %v", filepath, err))
		return
	}
	defer file.Close()

	f := strings.Split(filepath, "/")
	filename := f[len(f)-1]
	t := time.Now()
	key := t.Format("2006_01_02") + "_" + filename

	uploader := s3manager.NewUploader(b.sess)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(key),
		Body:   file,
	})
	if err != nil {
		b.logger.Error(fmt.Sprintf("Failed to upload file, %v", err))
		return
	}
}

func (b *Backup) Download(filepath string) {
	downloader := s3manager.NewDownloader(b.sess)
	file, err := os.Create(filepath)
	if err != nil {
		b.logger.Error(fmt.Sprintf("Failed to create file %q, %v", file.Name(), err))
		return
	}

	f := strings.Split(filepath, "/")
	filename := f[len(f)-1]
	t := time.Now()
	key := t.Format("2006_01_02") + "_" + filename

	_, err = downloader.Download(file,
		&s3.GetObjectInput{
			Bucket: aws.String(b.bucket),
			Key:    aws.String(key),
		})
	if err != nil {
		b.logger.Error(fmt.Sprintf("Failed to download file, %v", err))
		return
	}
}
