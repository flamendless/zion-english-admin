-- name: GetAffiliatedProductShopByBrandName :one
SELECT
	id,
	brand_name
FROM tbl_affiliated_product_shops
WHERE brand_name = ?;

-- name: InsertAffiliatedProductShop :one
INSERT INTO tbl_affiliated_product_shops (brand_name)
VALUES (?)
RETURNING id;

-- name: GetAffiliatedProductShopByID :one
SELECT
	id,
	brand_name
FROM tbl_affiliated_product_shops
WHERE id = ?;
