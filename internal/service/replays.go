package service

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/rankguessr/api/internal/config"
	"github.com/rankguessr/api/pkg/utils"
)

type Replays interface {
	AnonymizeAndPresign(ctx context.Context, data []byte, filename string, ttl time.Duration) (string, error)
	CreateAndPresign(ctx context.Context, data []byte, size int64, filename string, ttl time.Duration) (string, error)
}

type replays struct {
	cfg *config.Config
	s3  *minio.Client
}

func NewReplays(cfg *config.Config, s3 *minio.Client) Replays {
	return &replays{
		cfg: cfg,
		s3:  s3,
	}
}

func (s *replays) AnonymizeAndPresign(ctx context.Context, data []byte, filename string, ttl time.Duration) (string, error) {
	anonymized, err := utils.AnonymizeReplay(data)
	if err != nil {
		return "", err
	}

	return s.CreateAndPresign(ctx, anonymized, int64(len(anonymized)), filename, ttl)
}

func (s *replays) CreateAndPresign(ctx context.Context, data []byte, size int64, filename string, ttl time.Duration) (string, error) {
	location := fmt.Sprintf("replays/%s", filename)
	_, err := s.s3.PutObject(ctx, s.cfg.S3BucketName, location, bytes.NewReader(data), size, minio.PutObjectOptions{})
	if err != nil {
		return "", err
	}

	u, err := s.s3.PresignedGetObject(ctx, s.cfg.S3BucketName, location, ttl, nil)
	if err != nil {
		return "", err
	}

	return u.String(), nil
}
