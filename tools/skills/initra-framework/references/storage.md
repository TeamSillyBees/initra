# 文件与对象存储

## 注册

```go
storageprovider.Register(injector, cfg.Storage)
```

业务模块依赖 `storage.Service`，不要 import 具体 provider。

## 使用

```go
object, err := s.storage.Upload(ctx, storage.UploadInput{
	Key:         key,
	FileName:    filename,
	Body:        body,
	Size:        size,
	ContentType: contentType,
	Overwrite:   false,
})
```

下载、删除、公开 URL 和对象元信息都通过同一个 service 接口完成。云 provider 支持预签名 URL；local provider 没有可验证的签名和过期端点，因此 `PresignUpload`、`PresignDownload` 明确返回 `storage.ErrUnsupported`。

## Provider

配置切换 provider：`local`、`aliyun_oss`、`tencent_cos`、`aws_s3`、`s3_compatible`。业务代码不应感知差异。

## 高级能力

- 分片上传：在需要时判断实现是否满足 `storage.MultipartService`。
- STS：在 boot 层调用 `storageprovider.NewSTS(ctx, cfg.Storage)` 创建并注册独立的 `storage.STSService`；当前支持 AWS S3、阿里云 OSS 和腾讯云 COS，local 与 S3 compatible 返回 `ErrUnsupported`。

## 禁止

- 不要在业务模块直接使用 OSS/COS/S3 SDK。
- 不要把 access key、secret、STS token 写入日志或错误详情。
