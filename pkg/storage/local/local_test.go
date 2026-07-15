package local

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/teamsillybees/initra/pkg/storage"
)

func TestServiceUploadDownloadListAndDelete(t *testing.T) {
	ctx := context.Background()
	service := newTestService(t, storage.LocalConfig{
		RootDir:           t.TempDir(),
		PublicBaseURL:     "http://localhost/files",
		AllowedExtensions: []string{"txt"},
	})

	object, err := service.Upload(ctx, storage.UploadInput{
		Key:         "docs/readme.txt",
		FileName:    "readme.txt",
		Body:        strings.NewReader("hello initra"),
		ContentType: "text/plain",
		Metadata:    map[string]string{"owner": "test"},
	})
	require.NoError(t, err)
	require.Equal(t, "docs/readme.txt", object.Key)
	require.Equal(t, int64(12), object.Size)
	require.Equal(t, "http://localhost/files/docs/readme.txt", object.URL)
	require.Equal(t, map[string]string{"owner": "test"}, object.Metadata)

	content, err := service.DownloadBytes(ctx, storage.DownloadInput{Key: object.Key})
	require.NoError(t, err)
	require.Equal(t, "hello initra", string(content))

	exists, err := service.Exists(ctx, storage.ObjectInput{Key: object.Key})
	require.NoError(t, err)
	require.True(t, exists)

	stat, err := service.Stat(ctx, storage.ObjectInput{Key: object.Key})
	require.NoError(t, err)
	require.Equal(t, object.Key, stat.Key)
	require.Equal(t, object.Size, stat.Size)

	list, err := service.List(ctx, storage.ListInput{Prefix: "docs", MaxKeys: 1})
	require.NoError(t, err)
	require.Len(t, list.Objects, 1)
	require.Equal(t, object.Key, list.Objects[0].Key)

	_, err = service.Upload(ctx, storage.UploadInput{
		Key:  object.Key,
		Body: strings.NewReader("again"),
	})
	require.ErrorIs(t, err, storage.ErrObjectExists)

	require.NoError(t, service.Delete(ctx, storage.DeleteInput{Key: object.Key}))
	exists, err = service.Exists(ctx, storage.ObjectInput{Key: object.Key})
	require.NoError(t, err)
	require.False(t, exists)
}

func TestServiceRejectsDisallowedExtensionAndOversizedFile(t *testing.T) {
	ctx := context.Background()
	service := newTestService(t, storage.LocalConfig{
		RootDir:           t.TempDir(),
		AllowedExtensions: []string{"txt"},
		MaxSize:           4,
	})

	_, err := service.Upload(ctx, storage.UploadInput{
		Key:  "docs/readme.exe",
		Body: strings.NewReader("data"),
	})
	require.ErrorContains(t, err, "不支持")

	_, err = service.Upload(ctx, storage.UploadInput{
		Key:  "docs/readme.txt",
		Body: strings.NewReader("too large"),
	})
	require.ErrorContains(t, err, "文件大小超过限制")
}

func TestServiceMultipartUploadCombinesParts(t *testing.T) {
	ctx := context.Background()
	service := newTestService(t, storage.LocalConfig{RootDir: t.TempDir()})

	upload, err := service.CreateMultipartUpload(ctx, storage.MultipartUploadInput{Key: "large/report.txt"})
	require.NoError(t, err)
	require.NotEmpty(t, upload.UploadID)

	part2, err := service.UploadPart(ctx, storage.UploadPartInput{
		Key:        upload.Key,
		UploadID:   upload.UploadID,
		PartNumber: 2,
		Body:       strings.NewReader(" world"),
	})
	require.NoError(t, err)
	part1, err := service.UploadPart(ctx, storage.UploadPartInput{
		Key:        upload.Key,
		UploadID:   upload.UploadID,
		PartNumber: 1,
		Body:       strings.NewReader("hello"),
	})
	require.NoError(t, err)

	object, err := service.CompleteMultipartUpload(ctx, storage.CompleteMultipartUploadInput{
		Key:      upload.Key,
		UploadID: upload.UploadID,
		Parts:    []storage.UploadedPart{*part2, *part1},
	})
	require.NoError(t, err)
	require.Equal(t, int64(11), object.Size)

	content, err := service.DownloadBytes(ctx, storage.DownloadInput{Key: upload.Key})
	require.NoError(t, err)
	require.Equal(t, "hello world", string(content))
}

func TestServiceRejectsPathTraversal(t *testing.T) {
	ctx := context.Background()
	service := newTestService(t, storage.LocalConfig{RootDir: t.TempDir()})

	_, err := service.Upload(ctx, storage.UploadInput{
		Key:  "../secret.txt",
		Body: strings.NewReader("secret"),
	})
	require.True(t, errors.Is(err, storage.ErrInvalidKey))
}

func TestServiceMultipartRejectsUnsafeBucketAndUploadID(t *testing.T) {
	tests := []struct {
		name     string
		bucket   string
		uploadID string
		execute  func(context.Context, *Service, string, string) error
	}{
		{
			name:     "upload part rejects bucket traversal",
			bucket:   "../..",
			uploadID: "outside",
			execute: func(ctx context.Context, service *Service, bucket string, uploadID string) error {
				_, err := service.UploadPart(ctx, storage.UploadPartInput{
					Bucket: bucket, Key: "result.txt", UploadID: uploadID, PartNumber: 1, Body: strings.NewReader("unsafe"),
				})
				return err
			},
		},
		{
			name:     "upload part rejects upload id traversal",
			uploadID: "../../outside",
			execute: func(ctx context.Context, service *Service, bucket string, uploadID string) error {
				_, err := service.UploadPart(ctx, storage.UploadPartInput{
					Bucket: bucket, Key: "result.txt", UploadID: uploadID, PartNumber: 1, Body: strings.NewReader("unsafe"),
				})
				return err
			},
		},
		{
			name:     "complete rejects bucket traversal",
			bucket:   "../..",
			uploadID: "outside",
			execute: func(ctx context.Context, service *Service, bucket string, uploadID string) error {
				_, err := service.CompleteMultipartUpload(ctx, storage.CompleteMultipartUploadInput{
					Bucket: bucket, Key: "result.txt", UploadID: uploadID, Parts: []storage.UploadedPart{{PartNumber: 1}},
				})
				return err
			},
		},
		{
			name:     "complete rejects upload id traversal",
			uploadID: "../../outside",
			execute: func(ctx context.Context, service *Service, bucket string, uploadID string) error {
				_, err := service.CompleteMultipartUpload(ctx, storage.CompleteMultipartUploadInput{
					Bucket: bucket, Key: "result.txt", UploadID: uploadID, Parts: []storage.UploadedPart{{PartNumber: 1}},
				})
				return err
			},
		},
		{
			name:     "abort rejects bucket traversal",
			bucket:   "../..",
			uploadID: "outside",
			execute: func(ctx context.Context, service *Service, bucket string, uploadID string) error {
				return service.AbortMultipartUpload(ctx, storage.AbortMultipartUploadInput{
					Bucket: bucket, Key: "result.txt", UploadID: uploadID,
				})
			},
		},
		{
			name:     "abort rejects upload id traversal",
			uploadID: "../../outside",
			execute: func(ctx context.Context, service *Service, bucket string, uploadID string) error {
				return service.AbortMultipartUpload(ctx, storage.AbortMultipartUploadInput{
					Bucket: bucket, Key: "result.txt", UploadID: uploadID,
				})
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parent := t.TempDir()
			root := filepath.Join(parent, "storage")
			outside := filepath.Join(parent, "outside")
			require.NoError(t, os.MkdirAll(outside, 0o755))
			require.NoError(t, os.WriteFile(filepath.Join(outside, "sentinel.txt"), []byte("keep"), 0o600))
			require.NoError(t, os.WriteFile(filepath.Join(outside, "1"), []byte("part"), 0o600))
			service := newTestService(t, storage.LocalConfig{RootDir: root})

			err := tt.execute(context.Background(), service, tt.bucket, tt.uploadID)
			require.ErrorIs(t, err, storage.ErrInvalidKey)
			require.FileExists(t, filepath.Join(outside, "sentinel.txt"), "恶意分片参数不得递归删除临时根目录外的文件")
			require.FileExists(t, filepath.Join(outside, "1"), "恶意分片参数不得覆盖临时根目录外的分片")
			require.NoFileExists(t, filepath.Join(root, "result.txt"), "参数校验应在创建合并目标前完成")
		})
	}
}

func TestServiceMultipartRejectsInvalidPartNumberBeforeCreatingTarget(t *testing.T) {
	ctx := context.Background()
	root := t.TempDir()
	service := newTestService(t, storage.LocalConfig{RootDir: root})

	_, err := service.UploadPart(ctx, storage.UploadPartInput{
		Key: "result.txt", UploadID: "valid-upload", PartNumber: 0, Body: strings.NewReader("invalid"),
	})
	require.ErrorIs(t, err, storage.ErrInvalidConfig)

	_, err = service.CompleteMultipartUpload(ctx, storage.CompleteMultipartUploadInput{
		Key: "result.txt", UploadID: "valid-upload", Parts: []storage.UploadedPart{{PartNumber: -1}},
	})
	require.ErrorIs(t, err, storage.ErrInvalidConfig)
	require.NoFileExists(t, filepath.Join(root, "result.txt"), "分片编号应在创建合并目标前校验")
}

func TestNewRejectsMultipartTempDirOutsideStorageRoot(t *testing.T) {
	parent := t.TempDir()
	_, err := New(storage.Config{
		Enabled:  true,
		Provider: storage.ProviderLocal,
		Local: storage.LocalConfig{
			RootDir: filepath.Join(parent, "storage"),
			TempDir: "../outside",
		},
	})
	require.ErrorIs(t, err, storage.ErrInvalidConfig)
	require.NoDirExists(t, filepath.Join(parent, "outside"))
}

func newTestService(t *testing.T, localCfg storage.LocalConfig) *Service {
	t.Helper()
	service, err := New(storage.Config{
		Enabled:  true,
		Provider: storage.ProviderLocal,
		Local:    localCfg,
	})
	require.NoError(t, err)
	return service
}
