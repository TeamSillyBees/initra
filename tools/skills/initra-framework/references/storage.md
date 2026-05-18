# storage

当业务需要文件上传/下载、本地存储、对象存储、预签名 URL、STS、删除、元信息、列表、复制、移动或分片上传时，使用 `github.com/teamsillybees/initra/pkg/storage` 和 `github.com/teamsillybees/initra/pkg/storage/provider`。

## 标准装配

在业务 `Config` 中组合 storage 配置，并在 boot 层注册 provider：

```go
Storage platformstorage.Config `mapstructure:"storage"`

storageprovider.Register(injector, cfg.Storage)
```

provider 根据 `storage.provider` 选择 local、Aliyun OSS、Tencent COS、AWS S3 或 S3 compatible 实现。

## 业务用法

业务模块依赖 `storage.Service` 或更窄的模块内接口：

```go
type fileStorage interface {
	Upload(ctx context.Context, input storage.UploadInput) (*storage.Object, error)
	DownloadBytes(ctx context.Context, input storage.DownloadInput) ([]byte, error)
	Delete(ctx context.Context, input storage.DeleteInput) error
	Stat(ctx context.Context, input storage.ObjectInput) (*storage.Object, error)
}
```

在存储适配边界使用 `storage.UploadInput`、`DownloadInput`、`ObjectInput`、`PresignInput` 和 `STSTokenInput`。

## 错误映射

将 storage 错误映射为应用错误：

```go
switch {
case storage.IsNotFound(err):
	return apperrors.Wrap(err, apperrors.CodeNotFound, "文件不存在")
case errors.Is(err, storage.ErrInvalidKey), errors.Is(err, storage.ErrInvalidConfig):
	return apperrors.Wrap(err, apperrors.CodeBadRequest, "文件请求无效")
default:
	return apperrors.Wrap(err, apperrors.CodeInternalError, "存储操作失败")
}
```

## 禁止写法

- 不要在业务模块中 import 云厂商 SDK。
- 不要在业务模块中 import `pkg/storage/aliyunoss`、`awss3`、`tencentcos`、`s3compat` 或 `local`。
- 不要把本地文件系统路径暴露为稳定业务 ID。
- 不要记录凭证、带敏感 query 的签名 URL 或原始文件内容。

## 检查清单

- storage 是否通过 `storage.Config` 配置？
- boot 是否只调用一次 `storageprovider.Register`？
- 业务代码是否依赖 `storage.Service` 或窄接口？
- provider-specific 细节是否仅存在于配置中？
- storage 错误是否映射为统一应用错误？
