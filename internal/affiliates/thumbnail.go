package affiliates

import (
	"context"
	"strings"
	"time"

	"zion-english/internal/logs"

	"go.uber.org/zap"
)

func FetchThumbnailFromProductURL(ctx context.Context, productURL string) (string, error) {
	start := time.Now()
	productURL = strings.TrimSpace(productURL)
	if productURL == "" {
		return "", ErrInvalidProductURL
	}
	if !isValidHTTPSURL(productURL) {
		return "", ErrInvalidProductURL
	}
	if _, err := validateShopeeURL(productURL); err != nil {
		return "", err
	}

	pageHTML, err := fetchPageHTML(ctx, productURL)
	if err != nil {
		logs.Log().Warn("affiliate thumbnail fetch",
			zap.String("product_url", productURL),
			zap.Duration("duration", time.Since(start)),
			zap.Error(err),
		)
		return "", ErrFetchBlocked
	}

	thumb := extractMetaImage(pageHTML)
	if thumb == "" {
		return "", ErrThumbnailNotFound
	}
	logs.Log().Info("affiliate thumbnail fetch",
		zap.String("product_url", productURL),
		zap.Duration("duration", time.Since(start)),
		zap.Bool("ok", true),
	)
	return thumb, nil
}
