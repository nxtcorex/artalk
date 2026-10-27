package imgupload

import (
	"bytes"
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/artalkjs/artalk/v2/internal/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// UploadS3 uploads a validated image directly to S3 without creating a local file.
// PublicPath points to the bucket root; Prefix is included in the returned URL.
func UploadS3(ctx context.Context, conf config.S3Conf, filename, contentType string, data []byte) (string, error) {
	publicURL, err := parseS3URL(conf.PublicPath, "public_path")
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(conf.Bucket) == "" {
		return "", fmt.Errorf("img_upload.s3.bucket is required")
	}
	if (conf.AccessKeyID == "") != (conf.SecretAccessKey == "") ||
		(conf.SessionToken != "" && conf.AccessKeyID == "") {
		return "", fmt.Errorf("img_upload.s3 credentials require both access_key_id and secret_access_key")
	}
	if conf.Endpoint != "" {
		if _, err := parseS3URL(conf.Endpoint, "endpoint"); err != nil {
			return "", err
		}
	}

	prefix := strings.Trim(conf.Prefix, "/")
	for _, segment := range strings.Split(prefix, "/") {
		if segment == "." || segment == ".." {
			return "", fmt.Errorf("img_upload.s3.prefix must not contain dot path segments")
		}
	}
	key := filename
	if prefix != "" {
		key = prefix + "/" + filename
	}
	publicURL.Path = strings.TrimRight(publicURL.Path, "/") + "/" + key
	publicURL.RawPath = ""

	timeout := time.Duration(conf.Timeout) * time.Second
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	region := conf.Region
	if region == "" {
		region = "us-east-1"
	}
	options := []func(*awsconfig.LoadOptions) error{
		awsconfig.WithRegion(region),
		awsconfig.WithRetryMaxAttempts(2),
	}
	if conf.AccessKeyID != "" {
		options = append(options, awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(conf.AccessKeyID, conf.SecretAccessKey, conf.SessionToken),
		))
	}
	awsConf, err := awsconfig.LoadDefaultConfig(ctx, options...)
	if err != nil {
		return "", fmt.Errorf("load S3 configuration: %w", err)
	}
	client := s3.NewFromConfig(awsConf, func(o *s3.Options) {
		o.UsePathStyle = conf.PathStyle
		if conf.Endpoint != "" {
			o.BaseEndpoint = aws.String(conf.Endpoint)
		}
		// Avoid optional streaming checksums unsupported by some S3-compatible services.
		o.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
	})
	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(conf.Bucket),
		Key:           aws.String(key),
		Body:          bytes.NewReader(data),
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(int64(len(data))),
	})
	if err != nil {
		return "", fmt.Errorf("upload image to S3: %w", err)
	}
	return publicURL.String(), nil
}

func parseS3URL(value, field string) (*url.URL, error) {
	u, err := url.Parse(value)
	if err != nil || (u.Scheme != "https" && u.Scheme != "http") || u.Hostname() == "" ||
		u.User != nil || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" {
		return nil, fmt.Errorf("img_upload.s3.%s must be an absolute HTTP(S) URL without credentials, query or fragment", field)
	}
	return u, nil
}
