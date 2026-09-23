package storage

import (
	"context"
	"errors"

	"zion-english/internal/constants"
)

func TeacherDocumentKey(docType constants.TeacherDocumentType, basename string) string {
	base := SanitizeFilename(basename)
	subdir := docType.StorageSubdir()
	if subdir == "" {
		return base
	}
	return subdir + "/" + base
}

func GetTeacherDocument(ctx context.Context, store Storage, docType constants.TeacherDocumentType, basename string) (*Object, error) {
	key := TeacherDocumentKey(docType, basename)
	obj, err := store.Get(ctx, CategoryTeacherDocuments, key)
	if err == nil {
		return obj, nil
	}
	if !errors.Is(err, ErrObjectNotFound) {
		return nil, err
	}
	return store.Get(ctx, CategoryTeacherDocuments, SanitizeFilename(basename))
}

func DeleteTeacherDocument(ctx context.Context, store Storage, docType constants.TeacherDocumentType, basename string) error {
	key := TeacherDocumentKey(docType, basename)
	if err := store.Delete(ctx, CategoryTeacherDocuments, key); err != nil {
		return err
	}
	flat := SanitizeFilename(basename)
	if flat != key {
		_ = store.Delete(ctx, CategoryTeacherDocuments, flat)
	}
	return nil
}
