package frontend

import (
	"slices"
	"strings"

	"zion-english/internal/auth"
)

type NavAccessMatrixRow struct {
	Title     string
	Path      string
	Superuser bool
	Admin     bool
	Teacher   bool
	Tester    bool
}

func NavAccessMatrixRows() []NavAccessMatrixRow {
	type pathTitle struct {
		path  string
		title string
	}
	items := []pathTitle{
		{path: "/dashboard", title: "Dashboard"},
		{path: "/changelogs", title: "Changelogs"},
	}
	for _, def := range navItemDefs {
		items = append(items, pathTitle{path: def.Path, title: def.Title})
	}
	items = append(items, pathTitle{path: "/students/register", title: "Register student"})

	seen := make(map[string]bool, len(items))
	rows := make([]NavAccessMatrixRow, 0, len(items))
	for _, item := range items {
		if seen[item.path] {
			continue
		}
		seen[item.path] = true
		rows = append(rows, navAccessMatrixRowForPath(item.path, item.title))
	}

	slices.SortFunc(rows, func(a, b NavAccessMatrixRow) int {
		return strings.Compare(strings.ToLower(a.Title), strings.ToLower(b.Title))
	})
	return rows
}

func navAccessMatrixRowForPath(path, title string) NavAccessMatrixRow {
	return NavAccessMatrixRow{
		Title:     title,
		Path:      path,
		Superuser: IsNavAccessible(auth.RoleSuperuser, path),
		Admin:     IsNavAccessible(auth.RoleAdmin, path),
		Teacher:   IsNavAccessible(auth.RoleTeacher, path),
		Tester:    IsNavAccessible(auth.RoleTester, path),
	}
}

func navAccessMatrixRowMatchesIsNavAccessible(row NavAccessMatrixRow) bool {
	return row.Superuser == IsNavAccessible(auth.RoleSuperuser, row.Path) &&
		row.Admin == IsNavAccessible(auth.RoleAdmin, row.Path) &&
		row.Teacher == IsNavAccessible(auth.RoleTeacher, row.Path) &&
		row.Tester == IsNavAccessible(auth.RoleTester, row.Path)
}
