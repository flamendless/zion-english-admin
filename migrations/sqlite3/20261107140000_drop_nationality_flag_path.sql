-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tbl_nationalities_new (
	id TEXT PRIMARY KEY NOT NULL,
	country_name TEXT NOT NULL,
	language TEXT
);

INSERT INTO tbl_nationalities_new (id, country_name, language)
SELECT id, country_name, language FROM tbl_nationalities;

DROP TABLE tbl_nationalities;
ALTER TABLE tbl_nationalities_new RENAME TO tbl_nationalities;
CREATE INDEX IF NOT EXISTS idx_nationalities_country_name ON tbl_nationalities(country_name);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tbl_nationalities_old (
	id TEXT PRIMARY KEY NOT NULL,
	country_name TEXT NOT NULL,
	language TEXT,
	flag_path TEXT NOT NULL DEFAULT ''
);

INSERT INTO tbl_nationalities_old (id, country_name, language, flag_path)
SELECT id, country_name, language, '' FROM tbl_nationalities;

DROP TABLE tbl_nationalities;
ALTER TABLE tbl_nationalities_old RENAME TO tbl_nationalities;
CREATE INDEX IF NOT EXISTS idx_nationalities_country_name ON tbl_nationalities(country_name);
-- +goose StatementEnd
