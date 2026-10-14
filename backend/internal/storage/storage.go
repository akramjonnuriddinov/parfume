// Package storage keeps uploaded files either on the local disk (development)
// or in an S3-compatible bucket (production, e.g. Laravel Cloud / R2).
package storage

import (
	"context"
	"errors"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// ErrNotFound is returned by Open when the file does not exist.
var ErrNotFound = errors.New("file not found")

// Object is an opened file ready to be streamed to a client.
type Object struct {
	io.ReadCloser
	ContentType string
	Size        int64
}

type Storage interface {
	// Put stores r under key (a slash-separated path like "2026/10/x.jpg").
	Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error
	// Open returns the stored file or ErrNotFound.
	Open(ctx context.Context, key string) (*Object, error)
}

// S3Config is read from the standard AWS_* variables that Laravel Cloud
// injects when a bucket is attached to the environment.
type S3Config struct {
	Bucket    string
	Endpoint  string
	Region    string
	AccessKey string
	SecretKey string
}

// New returns a bucket store when cfg.Bucket is set, otherwise a disk store.
func New(cfg S3Config, dir string) (Storage, error) {
	if cfg.Bucket == "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, err
		}
		return Disk{Dir: dir}, nil
	}
	u, err := url.Parse(cfg.Endpoint)
	if err != nil || u.Host == "" {
		return nil, errors.New("AWS_ENDPOINT_URL must be a URL like https://<id>.r2.cloudflarestorage.com")
	}
	client, err := minio.New(u.Host, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: u.Scheme != "http",
		Region: cfg.Region,
	})
	if err != nil {
		return nil, err
	}
	return &S3{client: client, bucket: cfg.Bucket}, nil
}

// cleanKey rejects absolute paths and "..", so a key can never escape the store.
func cleanKey(key string) (string, bool) {
	key = strings.TrimPrefix(filepath.ToSlash(filepath.Clean("/"+key)), "/")
	return key, key != "" && key != "."
}

type Disk struct{ Dir string }

func (d Disk) Put(_ context.Context, key string, r io.Reader, _ int64, _ string) error {
	key, ok := cleanKey(key)
	if !ok {
		return errors.New("invalid key")
	}
	path := filepath.Join(d.Dir, filepath.FromSlash(key))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, r); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

func (d Disk) Open(_ context.Context, key string) (*Object, error) {
	key, ok := cleanKey(key)
	if !ok {
		return nil, ErrNotFound
	}
	f, err := os.Open(filepath.Join(d.Dir, filepath.FromSlash(key)))
	if err != nil {
		return nil, ErrNotFound
	}
	st, err := f.Stat()
	if err != nil || st.IsDir() {
		f.Close()
		return nil, ErrNotFound
	}
	return &Object{ReadCloser: f, Size: st.Size()}, nil
}

type S3 struct {
	client *minio.Client
	bucket string
}

func (s *S3) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) error {
	key, ok := cleanKey(key)
	if !ok {
		return errors.New("invalid key")
	}
	_, err := s.client.PutObject(ctx, s.bucket, key, r, size, minio.PutObjectOptions{
		ContentType:  contentType,
		CacheControl: "public, max-age=604800",
	})
	return err
}

func (s *S3) Open(ctx context.Context, key string) (*Object, error) {
	key, ok := cleanKey(key)
	if !ok {
		return nil, ErrNotFound
	}
	obj, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	// GetObject is lazy: Stat performs the request and reports a missing key.
	st, err := obj.Stat()
	if err != nil {
		obj.Close()
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &Object{ReadCloser: obj, ContentType: st.ContentType, Size: st.Size}, nil
}
