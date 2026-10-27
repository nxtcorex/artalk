# 图片上传

Artalk 提供图片上传功能，支持限制图片大小、上传频率等，可原生上传到 AWS S3 或兼容 S3 的对象存储，也能结合 UpGit 将图片上传到图床。

你可以在 [控制中心](../frontend/sidebar.md#设置) 的设置界面修改此配置，也可以通过 [配置文件](./config.md#图片上传-img_upload) 或 [环境变量](../env.md#图片上传) 进行配置。

## 配置文件

完整的 `img_upload` 配置如下：

```yaml
# 图片上传
img_upload:
  enabled: true # 总开关
  path: ./data/artalk-img/ # 图片存放路径
  max_size: 5 # 图片大小限制 (单位：MB)
  public_path: null # 指定图片链接基础路径 (默认为 "/static/images/")
  # 使用 upgit 将图片上传到 GitHub 或图床
  upgit:
    enabled: false # 启用 upgit
    exec: upgit -c <upgit配置文件路径> -t /artalk-img
    del_local: true # 上传后删除本地的图片
  # 原生 S3 上传，S3 与 UpGit 任选一种启用
  s3:
    enabled: false
    endpoint: "" # AWS S3 使用默认端点；兼容服务填写完整 HTTP(S) 端点
    region: us-east-1
    bucket: ""
    access_key_id: "" # 两个密钥均留空时使用 AWS 默认凭据链
    secret_access_key: ""
    session_token: "" # 使用临时凭据时填写
    prefix: artalk-img # 对象键前缀，可留空
    public_path: "" # 必填，存储桶根目录或 CDN 的公开访问 URL，自动追加 prefix
    path_style: false # MinIO 等服务可设为 true
    timeout: 30 # 上传超时，单位：秒
```

## 原生上传到 S3

启用 `img_upload.s3.enabled` 后，图片通过 Artalk 后端上传到对象存储，前端继续使用原有上传接口。图片格式、大小、验证码和频率限制照常生效，S3 模式直接将图片保存到对象存储。

### AWS S3 配置

```yaml
img_upload:
  enabled: true
  max_size: 5
  upgit:
    enabled: false
  s3:
    enabled: true
    region: ap-southeast-1
    bucket: my-images
    prefix: artalk-img
    public_path: https://my-images.s3.ap-southeast-1.amazonaws.com
```

### 认证与图片链接

认证可通过 `access_key_id`、`secret_access_key` 和可选的 `session_token` 配置。两个密钥均留空时使用 AWS SDK 默认凭据链，例如服务端的 `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` / `AWS_SESSION_TOKEN` 环境变量、共享凭据文件或 IAM 角色。上传凭据需要对目标前缀具有 `s3:PutObject` 权限。

`public_path` 必须是指向存储桶根目录的公开 HTTP(S) URL，可以使用 CDN 域名；Artalk 会在后面追加 `prefix` 和生成的唯一文件名。例如 `public_path: https://img.example.com`、`prefix: artalk-img` 会返回 `https://img.example.com/artalk-img/文件名.png`。`img_upload.s3.public_path` 用于 S3 图片链接，`img_upload.public_path` 用于本地图片链接。

### 兼容 S3 的服务

MinIO、Cloudflare R2 等兼容 S3 的服务可额外配置 `endpoint`。例如 MinIO：

```yaml
img_upload:
  s3:
    enabled: true
    endpoint: https://minio.example.com
    region: us-east-1
    bucket: my-images
    access_key_id: YOUR_ACCESS_KEY_ID
    secret_access_key: YOUR_SECRET_ACCESS_KEY
    prefix: artalk-img
    public_path: https://minio.example.com/my-images
    path_style: true
    timeout: 30
```

### 寻址、区域与超时

`path_style: true` 使用 `endpoint/bucket/key` 形式，默认使用虚拟主机风格寻址。`region` 默认为 `us-east-1`，`timeout` 默认为 30 秒，可填写正整数秒数。S3 与 UpGit 任选一种启用即可。

### 环境变量

以上字段也支持 `ATK_IMG_UPLOAD_S3_*` 环境变量，例如 `ATK_IMG_UPLOAD_S3_ENABLED`、`ATK_IMG_UPLOAD_S3_SECRET_ACCESS_KEY` 和 `ATK_IMG_UPLOAD_S3_PUBLIC_PATH`，完整列表见[环境变量文档](../env.md#图片上传)。

## 使用 UpGit 上传到图床

[UpGit](https://github.com/pluveto/upgit) 支持将图片上传到 Github、Gitee、腾讯云 COS、七牛云、又拍云、SM.MS 等图床或代码仓库。

### 配置 UpGit

首先，根据 [README.md](https://github.com/pluveto/upgit) 的说明，下载 UpGit 并完成你需要上传的目标图床的配置。

然后，将 UpGit 加入系统的环境变量中，在 `~/.bashrc` 加入：

```bash
export PATH=$PATH:/path/to/upgit
```

（或者直接移入 `/usr/bin`）

最后，在 Artalk 的 `img_upload.upgit` 字段填入 UpGit 启动参数：

```yaml
upgit:
  enabled: true # 启用 upgit
  exec: upgit -c <upgit配置文件路径> -t /artalk-img
  del_local: true # 上传后删除本地的图片
```

::: tip 配置提示
从 `v2.8.4` 版本开始，Artalk 通过系统 `PATH` 查找 UpGit，请将 UpGit 所在目录加入 `PATH`。
:::

### Docker 挂载 UpGit

如果你使用 Docker 部署 Artalk，可以将 UpGit 可执行文件挂载到容器中：

```bash
docker run -d --name artalk -v /path/to/upgit:/usr/bin/upgit -v /path/to/artalk:/app/data -p 8080:23366 artalk
```

## 上传频率限制

频率限制跟随 `captcha` 验证码配置，当超出限制将弹出验证码。

可参考：[后端 · 验证码](./captcha.md)

## path

`img_upload.path` 为上传的图片文件「本地存放目录」路径，该目录会被 Artalk 映射到可访问的：

```
http://<后端地址>/static/images/
```

## public_path

`img.public_path` 为空的默认值为：`/static/images/`

当该项为「相对路径」时，例如：`/static/images/` 前端上传图片得到的 HTML 标签将为：

```html
<img src="http://<后端地址>/static/images/1.png" />
```

注：这里的 `<后端地址>` 是前端 `conf.server` 配置。

当该项为「完整 URL 路径」时，例如：`https://cdn.github.com/img/` 时，图片标签将为：

```html
<img src="https://cdn.github.com/img/1.png" />
```

提示：这个配置可以结合负载均衡等场景使用。

## 在前端自定义上传 API

前端提供了配置项 `imgUploader`，你可以自定义前端图片上传时请求的 API，例如：

```js
Artalk.init({
  imgUploader: async (file) => {
    const form = new FormData()
    form.set('file', file)

    const imgUrl = await fetch('https://api.example.org/upload', {
      method: 'POST',
      body: form,
    })

    return imgUrl
  },
})
```

参考：[前端配置文档](../frontend/config.md#imguploader)
