package utils

import (
	"math/rand"
	"path/filepath"
	"time"
	"zion-english/internal/constants"
)

const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

func RandomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = letters[rng.Intn(len(letters))]
	}
	return string(b)
}

func SanitizeFilename(name string) string {
	name = filepath.Base(name)
	name = constants.ReSafeName.ReplaceAllString(name, "_")
	if len(name) > 50 {
		name = name[:50]
	}
	if name == "" {
		name = "file"
	}
	return name
}
