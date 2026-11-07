-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS tbl_nationalities (
	id TEXT PRIMARY KEY NOT NULL,
	country_name TEXT NOT NULL,
	language TEXT
);

CREATE INDEX IF NOT EXISTS idx_nationalities_country_name ON tbl_nationalities(country_name);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_nationalities_country_name;
DROP TABLE IF EXISTS tbl_nationalities;
-- +goose StatementEnd
