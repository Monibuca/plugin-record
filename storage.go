package record

import (
	"context"
	"sync"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
)

var uploadSemaphore = make(chan struct{}, 8)
var once sync.Once

type StorageConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
}

func (r *Recorder) UploadFile(filePath string, fileName string) {
	// 使用信号量控制并发数
	uploadSemaphore <- struct{}{}
	defer func() { <-uploadSemaphore }()
	// 判断Storage是否配置，未配置就不上传
	if r.Storage.Endpoint == "" || r.Storage.SecretKey == "" || r.Storage.AccessKey == "" || r.Storage.Bucket == "" {
		r.Info("Minio Storage Config Not Configured")
		return
	}

	ctx := context.Background()

	endpoint := r.Storage.Endpoint
	accessKeyID := r.Storage.AccessKey
	secretAccessKey := r.Storage.SecretKey
	bucketName := r.Storage.Bucket
	useSSL := r.Storage.UseSSL
	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})

	if err != nil {
		r.Error("create minioClient error:", zap.Error(err))
	}

	// Make a new bucket called testbucket.
	location := "us-east-1"

	// 检查Bucket是否存在，不存在就创建Bucket
	exists, err := minioClient.BucketExists(ctx, bucketName)
	if err != nil {
		r.Error("Failed to check bucket existence:", zap.Error(err))
		return
	}

	if !exists {
		err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: location})
		if err != nil {
			r.Error("Create Bucket Error:", zap.Error(err))
			return
		}
		r.Info("Successfully created Bucket:", zap.String("bucket", bucketName))
	} else {
		r.Info("Bucket already exists:", zap.String("bucket", bucketName))
	}

	// Change the value of filePath if the file is in another location
	objectName := fileName
	fileFullPath := filePath + "/" + objectName
	r.Info("Prepare Upload  Path:  fileName:", zap.String("objectName", objectName))
	contentType := "application/octet-stream"

	// Upload the test file with FPutObject
	info, err := minioClient.FPutObject(ctx, bucketName, objectName, fileFullPath, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		r.Error("Minio PutObject Error:", zap.Error(err))
	}

	r.Info("Successfully uploaded of size ", zap.String("Key", info.Key), zap.Int64("Size", info.Size))

	r.RemoveRecordById()

	// Remove the file after upload
	// 使用定时删除几天前的数据，减少并发录制时写入+删除的磁盘I/O
	// err = os.Remove(fileFullPath)
	// if err != nil {
	// 	r.Error("Remove file Error:", zap.Error(err))
	// }
	// r.Info("Successfully Removed of size ", zap.String("fileFullPath", fileFullPath))
}
