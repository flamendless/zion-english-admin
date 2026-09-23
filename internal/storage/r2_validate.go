package storage

import (
	"context"

	"github.com/cloudflare/cloudflare-go/v7"
	"github.com/cloudflare/cloudflare-go/v7/option"
	cfr2 "github.com/cloudflare/cloudflare-go/v7/r2"

	"zion-english/internal/conf"
)

func ValidateR2Bucket(ctx context.Context, cfg *conf.Config) error {
	r2 := cfg.Storage.R2
	if r2.APIToken == "" {
		return nil
	}
	client := cloudflare.NewClient(option.WithAPIToken(r2.APIToken))
	_, err := client.R2.Buckets.Get(ctx, r2.Bucket, cfr2.BucketGetParams{
		AccountID: cloudflare.F(r2.AccountID),
	})
	return err
}
