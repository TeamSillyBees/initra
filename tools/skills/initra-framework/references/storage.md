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

下载、删除、预签名 URL、公开 URL 和对象元信息都通过同一个 service 接口完成。

## Provider

配置切换 provider：`local`、`aliyun_oss`、`tencent_cos`、`aws_s3`、`s3_compatible`。业务代码不应感知差异。

## 高级能力

- 分片上传：在需要时判断实现是否满足 `storage.MultipartService`。
- STS：在需要临时授权时判断实现是否满足 `storage.STSService`。

## 禁止

- 不要在业务模块直接使用 OSS/COS/S3 SDK。
- 不要把 access key、secret、STS token 写入日志或错误详情。
