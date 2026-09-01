package utils

import (
	"cmp"
	"slices"
	"strings"
)

func SortKey(by string, order SortOrder) string {
	return by + "_" + string(order)
}

func SortSlice[T any](items []T, order SortOrder, lessAsc func(a, b T) int) {
	slices.SortFunc(items, func(a, b T) int {
		cmpVal := lessAsc(a, b)
		if order == SortOrderDesc {
			return -cmpVal
		}
		return cmpVal
	})
}

func CompareStrings(a, b string) int {
	return cmp.Compare(strings.ToLower(a), strings.ToLower(b))
}

func CompareFloat64(a, b float64) int {
	return cmp.Compare(a, b)
}

func CompareInt64(a, b int64) int {
	return cmp.Compare(a, b)
}

func CompareBool(a, b bool) int {
	if a == b {
		return 0
	}
	if a {
		return 1
	}
	return -1
}
