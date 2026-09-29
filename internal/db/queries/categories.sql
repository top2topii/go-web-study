-- name: CreateCategory :one
INSERT INTO categories (parent_id, name, slug)
VALUES (?, ?, ?)
RETURNING *;

-- name: ListCategories :many
SELECT * FROM categories
ORDER BY COALESCE(parent_id, 0), name;
