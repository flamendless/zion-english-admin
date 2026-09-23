package storage

import (
	"context"
	"fmt"
	"sync"

	"zion-english/internal/conf"
)

var (
	defaultStorage Storage
	initOnce       sync.Once
	initErr        error
)

func Init(cfg *conf.Config) error {
	initOnce.Do(func() {
		defaultStorage, initErr = NewFromConfig(cfg)
		if initErr != nil {
			return
		}
		if err := defaultStorage.EnsureDirs(); err != nil {
			initErr = err
		}
	})
	return initErr
}

func Default() Storage {
	if defaultStorage == nil {
		panic("storage not initialized")
	}
	return defaultStorage
}

func NewFromConfig(cfg *conf.Config) (Storage, error) {
	if cfg.R2Enabled() {
		r2, err := NewR2Storage(
			cfg.Storage.R2.AccountID,
			cfg.Storage.R2.AccessKeyID,
			cfg.Storage.R2.SecretAccessKey,
			cfg.Storage.R2.Bucket,
		)
		if err != nil {
			return nil, fmt.Errorf("create r2 storage: %w", err)
		}
		if err := ValidateR2Bucket(context.Background(), cfg); err != nil {
			return nil, fmt.Errorf("validate r2 bucket: %w", err)
		}
		return r2, nil
	}
	return NewLocalStorage(), nil
}
