-- name: GetAllNationalities :many
SELECT
	id,
	country_name,
	language,
	flag_path
FROM tbl_nationalities
ORDER BY country_name ASC;

-- name: GetNationalityByID :one
SELECT
	id,
	country_name,
	language,
	flag_path
FROM tbl_nationalities
WHERE id = ?;
