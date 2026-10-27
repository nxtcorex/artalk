# Image Upload

Artalk provides an image upload feature with options to limit image size, upload frequency, and more. It supports native uploads to AWS S3 and S3-compatible object storage, as well as integration with UpGit for image hosting services.

You can modify this configuration in the settings interface of the [Dashboard](../frontend/sidebar.md#settings), or configure it through [configuration files](./config.md#image-upload-img_upload) or [environment variables](../env.md#image-upload).

## Configuration File

The complete `img_upload` configuration is as follows:

```yaml
# Image Upload
img_upload:
  enabled: true # Master Switch
  path: ./data/artalk-img/ # Image storage path
  max_size: 5 # Image size limit (Unit: MB)
  public_path: null # Specify the base URL for image links (default is "/static/images/")
  # Use UpGit to upload images to GitHub or image hosting services
  upgit:
    enabled: false # Enable UpGit
    exec: upgit -c <upgit configuration file path> -t /artalk-img
    del_local: true # Delete local images after uploading
  # Native S3 upload; choose either S3 or UpGit
  s3:
    enabled: false
    endpoint: "" # AWS S3 uses its default endpoint; compatible services use a full HTTP(S) endpoint
    region: us-east-1
    bucket: ""
    access_key_id: "" # Leave both keys empty to use the AWS default credential chain
    secret_access_key: ""
    session_token: "" # For temporary credentials
    prefix: artalk-img # Object key prefix; may be empty
    public_path: "" # Required public bucket root or CDN URL; prefix is appended automatically
    path_style: false # Set to true for services such as MinIO
    timeout: 30 # Upload timeout in seconds
```

## Native S3 Uploads

Enable `img_upload.s3.enabled` to upload images through the Artalk backend to object storage using the existing frontend upload API. Image format, size, captcha, and rate limits still apply. S3 mode saves images directly to object storage.

### AWS S3 Configuration

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

### Authentication and Image URLs

Configure authentication using `access_key_id`, `secret_access_key`, and an optional `session_token`. Leave both keys empty to use the AWS SDK default credential chain, including server-side `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` / `AWS_SESSION_TOKEN` environment variables, shared credentials files, or IAM roles. Upload credentials need `s3:PutObject` permission for the target prefix.

`public_path` must be a public HTTP(S) URL pointing to the bucket root, optionally through a CDN. Artalk appends `prefix` and a generated unique filename. For example, `public_path: https://img.example.com` and `prefix: artalk-img` return `https://img.example.com/artalk-img/filename.png`. `img_upload.s3.public_path` controls S3 image links, and `img_upload.public_path` controls local image links.

### S3-Compatible Services

For S3-compatible services such as MinIO and Cloudflare R2, also configure `endpoint`. For example, with MinIO:

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

### Addressing, Region, and Timeout

`path_style: true` uses `endpoint/bucket/key` addressing; the default is virtual-hosted addressing. `region` defaults to `us-east-1`, and `timeout` defaults to 30 seconds and accepts a positive integer in seconds. Choose either S3 or UpGit as the upload method.

### Environment Variables

All fields support `ATK_IMG_UPLOAD_S3_*` environment variables, such as `ATK_IMG_UPLOAD_S3_ENABLED`, `ATK_IMG_UPLOAD_S3_SECRET_ACCESS_KEY`, and `ATK_IMG_UPLOAD_S3_PUBLIC_PATH`. See the [environment variable reference](../env.md#upload) for the complete list.

## Using UpGit to Upload to Image Hosting Services

[UpGit](https://github.com/pluveto/upgit) supports uploading images to various image hosting services or code repositories such as GitHub, Gitee, Tencent Cloud COS, Qiniu Cloud, UpYun, SM.MS, and more.

### Configuring UpGit

First, download UpGit and complete the configuration for your target image hosting service according to the [README.md](https://github.com/pluveto/upgit).

Then, add UpGit to the system's environment variables by adding the following to `~/.bashrc`:

```bash
export PATH=$PATH:/path/to/upgit
```

(Or move it directly to `/usr/bin`)

Finally, fill in the UpGit startup parameters in Artalk's `img_upload.upgit` field:

```yaml
upgit:
  enabled: true # Enable UpGit
  exec: upgit -c <upgit configuration file path> -t /artalk-img
  del_local: true # Delete local images after uploading
```

::: tip Configuration
Starting from version `v2.8.4`, Artalk locates UpGit through the system `PATH`. Add the directory containing UpGit to `PATH`.
:::

### Mounting UpGit with Docker

If you are deploying Artalk with Docker, you can mount the UpGit executable to the container:

```bash
docker run -d --name artalk -v /path/to/upgit:/usr/bin/upgit -v /path/to/artalk:/app/data -p 8080:23366 artalk
```

## Upload Frequency Limit

The frequency limit follows the `captcha` configuration. When the limit is exceeded, a captcha will be prompted.

Refer to: [Backend · Captcha](./captcha.md)

## Path

`img_upload.path` is the "local storage directory" path for uploaded image files. This directory will be mapped by Artalk to be accessible at:

```
http://<backend address>/static/images/
```

## Public Path

The default value for `img.public_path` is: `/static/images/`

When this item is a "relative path", for example: `/static/images/`, the HTML tag for the uploaded image on the frontend will be:

```html
<img src="http://<backend address>/static/images/1.png" />
```

Note: The `<backend address>` is configured in the frontend `conf.server`.

When this item is a "complete URL path", for example: `https://cdn.github.com/img/`, the image tag will be:

```html
<img src="https://cdn.github.com/img/1.png" />
```

Tip: This configuration can be used in scenarios such as load balancing.

## Custom Upload API on the Frontend

The frontend provides the `imgUploader` configuration option, allowing you to customize the API for image upload requests, for example:

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

Refer to: [Frontend Configuration Documentation](../frontend/config.md#imguploader)
