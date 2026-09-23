package storage

import (
	"context"
	"io"
	"os"
	"path/filepath"
)

type LocalStorage struct {
	root string
}

func NewLocalStorage() *LocalStorage {
	return &LocalStorage{root: ""}
}

func (s *LocalStorage) Backend() string {
	return "local"
}

func (s *LocalStorage) categoryRoot(category Category) (string, bool) {
	dir, ok := localCategoryDir(category)
	if !ok {
		return "", false
	}
	if s.root == "" {
		return dir, true
	}
	return filepath.Join(s.root, dir), true
}

func (s *LocalStorage) EnsureDirs() error {
	for _, category := range []Category{
		CategoryAvatars,
		CategoryTeacherDocuments,
		CategoryIntroVideos,
		CategoryReports,
	} {
		dir, ok := s.categoryRoot(category)
		if !ok {
			continue
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
		if category == CategoryTeacherDocuments {
			for _, subdir := range []string{"document", "resume"} {
				if err := os.MkdirAll(filepath.Join(dir, subdir), 0755); err != nil {
					return err
				}
			}
		}
	}
	if s.root == "" {
		if err := os.MkdirAll("tmp", 0755); err != nil {
			return err
		}
	}
	return nil
}

func (s *LocalStorage) Put(ctx context.Context, category Category, filename string, body io.Reader, contentType string) error {
	path, err := s.localPath(category, filename)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	out, err := os.Create(path)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, body); err != nil {
		out.Close()
		_ = os.Remove(path)
		return err
	}
	return out.Close()
}

func (s *LocalStorage) Get(ctx context.Context, category Category, filename string) (*Object, error) {
	path, err := s.localPath(category, filename)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrObjectNotFound
		}
		return nil, err
	}
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrObjectNotFound
		}
		return nil, err
	}
	return &Object{
		Reader: f,
		Size:   info.Size(),
	}, nil
}

func (s *LocalStorage) Delete(ctx context.Context, category Category, filename string) error {
	path, err := s.localPath(category, filename)
	if err != nil {
		return err
	}
	err = os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

func (s *LocalStorage) Exists(ctx context.Context, category Category, filename string) (bool, error) {
	path, err := s.localPath(category, filename)
	if err != nil {
		return false, err
	}
	_, err = os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (s *LocalStorage) localPath(category Category, filename string) (string, error) {
	rel := SanitizeRelativePath(filename)
	if rel == "" {
		rel = SanitizeFilename(filename)
	}
	if rel == "" || rel == "." {
		return "", ErrObjectNotFound
	}
	dir, ok := s.categoryRoot(category)
	if !ok {
		return "", ErrObjectNotFound
	}
	return filepath.Join(dir, rel), nil
}

func localCategoryDir(category Category) (string, bool) {
	switch category {
	case CategoryAvatars:
		return "data/avatars", true
	case CategoryTeacherDocuments:
		return "data/teacher-documents", true
	case CategoryIntroVideos:
		return "data/teacher-intro-videos", true
	case CategoryReports:
		return "tmp", true
	case CategoryBackups:
		return "", false
	default:
		return "", false
	}
}
