// Command gen_nationalities downloads country metadata and writes a Goose seed migration.
// Flag SVG markup lives in internal/nationalities/flags.go (regenerate via tools/embed_flags).
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type countryAPI struct {
	CCA2      string            `json:"cca2"`
	Name      countryNames      `json:"name"`
	Languages map[string]string `json:"languages"`
}

type countryNames struct {
	Common   string `json:"common"`
	Official string `json:"official"`
}

func main() {
	repoRoot, err := findRepoRoot()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get("https://cdn.jsdelivr.net/npm/world-countries@5.1.0/countries.json")
	if err != nil {
		fmt.Fprintln(os.Stderr, "fetch countries:", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintln(os.Stderr, "fetch countries: status", resp.StatusCode)
		os.Exit(1)
	}

	var countries []countryAPI
	if err := json.NewDecoder(resp.Body).Decode(&countries); err != nil {
		fmt.Fprintln(os.Stderr, "decode countries:", err)
		os.Exit(1)
	}

	type row struct {
		id, name, language string
	}
	var rows []row

	for _, c := range countries {
		id := strings.ToLower(strings.TrimSpace(c.CCA2))
		if id == "" {
			continue
		}
		name := strings.TrimSpace(c.Name.Common)
		if name == "" {
			name = strings.TrimSpace(c.Name.Official)
		}
		if name == "" {
			continue
		}
		language := ""
		for _, lang := range c.Languages {
			language = lang
			break
		}
		rows = append(rows, row{id: id, name: name, language: language})
	}

	sort.Slice(rows, func(i, j int) bool { return rows[i].name < rows[j].name })

	migrationPath := filepath.Join(repoRoot, "migrations", "sqlite3", "20261107120200_seed_nationalities.sql")
	f, err := os.Create(migrationPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "create migration:", err)
		os.Exit(1)
	}
	defer f.Close()

	fmt.Fprintln(f, "-- +goose Up")
	fmt.Fprintln(f, "-- +goose StatementBegin")
	for _, r := range rows {
		fmt.Fprintf(f,
			"INSERT INTO tbl_nationalities (id, country_name, language) VALUES ('%s', %s, %s);\n",
			sqlEscape(r.id),
			sqlString(r.name),
			sqlString(r.language),
		)
	}
	fmt.Fprintln(f, "-- +goose StatementEnd")
	fmt.Fprintln(f, "")
	fmt.Fprintln(f, "-- +goose Down")
	fmt.Fprintln(f, "-- +goose StatementBegin")
	fmt.Fprintln(f, "DELETE FROM tbl_nationalities;")
	fmt.Fprintln(f, "-- +goose StatementEnd")

	fmt.Printf("Wrote %d nationalities to %s\n", len(rows), migrationPath)
}

func findRepoRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod not found")
		}
		dir = parent
	}
}

func sqlEscape(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

func sqlString(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "NULL"
	}
	return "'" + sqlEscape(s) + "'"
}
