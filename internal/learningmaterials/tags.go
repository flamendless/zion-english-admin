package learningmaterials

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"zion-english/internal/database/queries"
	"zion-english/internal/processor"
)

func NextTagColor(index int) string {
	return processor.StudentColorHex(index)
}

func GetOrCreateTag(ctx context.Context, q *queries.Queries, label string, colorIndex int) (int64, error) {
	existing, err := q.GetLearningMaterialTagByLabel(ctx, label)
	if err == nil {
		return existing.ID, nil
	}
	if !errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("lookup tag %q: %w", label, err)
	}

	id, err := q.InsertLearningMaterialTag(ctx, queries.InsertLearningMaterialTagParams{
		Label: label,
		Color: NextTagColor(colorIndex),
	})
	if err != nil {
		return 0, fmt.Errorf("create tag %q: %w", label, err)
	}
	return id, nil
}

func ReplaceMaterialTags(ctx context.Context, q *queries.Queries, materialID int64, labels []string) error {
	normalized := NormalizeTagLabels(labels)
	if len(normalized) < MinTags || len(normalized) > MaxTags {
		return ErrTagCount
	}

	if err := q.DeleteLearningMaterialTagLinks(ctx, materialID); err != nil {
		return fmt.Errorf("clear tag links: %w", err)
	}

	tagCount, err := q.CountLearningMaterialTags(ctx)
	if err != nil {
		return fmt.Errorf("count tags: %w", err)
	}

	for i, label := range normalized {
		tagID, err := GetOrCreateTag(ctx, q, label, int(tagCount)+i)
		if err != nil {
			return err
		}
		if err := q.InsertLearningMaterialTagLink(ctx, queries.InsertLearningMaterialTagLinkParams{
			MaterialID: materialID,
			TagID:      tagID,
		}); err != nil {
			return fmt.Errorf("link tag %q: %w", label, err)
		}
	}
	return nil
}
