package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type R2Storage struct {
	client *s3.Client
	bucket string
}

func NewR2Storage(accountID, accessKeyID, secretAccessKey, bucket string) (*R2Storage, error) {
	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)

	resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:               endpoint,
			HostnameImmutable: true,
		}, nil
	})

	cfg := aws.Config{
		Region: "auto",
		Credentials: aws.NewCredentialsCache(credentials.NewStaticCredentialsProvider(
			accessKeyID,
			secretAccessKey,
			"",
		)),
		EndpointResolverWithOptions: resolver,
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return &R2Storage{
		client: client,
		bucket: bucket,
	}, nil
}

func (s *R2Storage) Backend() string {
	return "r2"
}

func (s *R2Storage) EnsureDirs() error {
	return nil
}

func (s *R2Storage) objectKey(category Category, filename string) (string, error) {
	rel := SanitizeRelativePath(filename)
	if rel == "" {
		rel = SanitizeFilename(filename)
	}
	if rel == "" || rel == "." {
		return "", ErrObjectNotFound
	}
	return ObjectKey(category, rel), nil
}

func (s *R2Storage) Put(ctx context.Context, category Category, filename string, body io.Reader, contentType string) error {
	key, err := s.objectKey(category, filename)
	if err != nil {
		return err
	}
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	input := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        body,
		ContentType: aws.String(contentType),
	}
	if length, ok := contentLengthForPut(body); ok {
		input.ContentLength = aws.Int64(length)
	}
	_, err = s.client.PutObject(ctx, input)
	return err
}

func contentLengthForPut(body io.Reader) (int64, bool) {
	switch r := body.(type) {
	case *bytes.Reader:
		return int64(r.Len()), true
	case *bytes.Buffer:
		return int64(r.Len()), true
	case *strings.Reader:
		return int64(r.Len()), true
	}
	rs, ok := body.(io.ReadSeeker)
	if !ok {
		return 0, false
	}
	cur, err := rs.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, false
	}
	end, err := rs.Seek(0, io.SeekEnd)
	if err != nil {
		return 0, false
	}
	if _, err := rs.Seek(cur, io.SeekStart); err != nil {
		return 0, false
	}
	return end - cur, true
}

func (s *R2Storage) Get(ctx context.Context, category Category, filename string) (*Object, error) {
	key, err := s.objectKey(category, filename)
	if err != nil {
		return nil, err
	}
	out, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if isS3NotFound(err) {
			return nil, ErrObjectNotFound
		}
		return nil, err
	}
	contentType := ""
	if out.ContentType != nil {
		contentType = *out.ContentType
	}
	size := int64(0)
	if out.ContentLength != nil {
		size = *out.ContentLength
	}
	return &Object{
		Reader:      out.Body,
		ContentType: contentType,
		Size:        size,
	}, nil
}

func (s *R2Storage) Delete(ctx context.Context, category Category, filename string) error {
	key, err := s.objectKey(category, filename)
	if err != nil {
		return err
	}
	_, err = s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}

func (s *R2Storage) Exists(ctx context.Context, category Category, filename string) (bool, error) {
	key, err := s.objectKey(category, filename)
	if err != nil {
		return false, err
	}
	_, err = s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		if isS3NotFound(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func isS3NotFound(err error) bool {
	var notFound *types.NotFound
	var noSuchKey *types.NoSuchKey
	return errors.As(err, &notFound) || errors.As(err, &noSuchKey)
}
