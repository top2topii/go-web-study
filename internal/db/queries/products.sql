-- name: CreateProduct :one
INSERT INTO products (category_id, sku, name, description, price, stock, status)
VALUES (?, ?, ?, ?, ?, ?, ?)
RETURNING *;

-- name: GetProduct :one
SELECT * FROM products
WHERE id = ?;

-- name: ListProducts :many
SELECT p.id, p.sku, p.name, p.price, p.stock, p.status, p.category_id,
       c.name AS category_name, p.created_at, p.updated_at
FROM products p
LEFT JOIN categories c ON c.id = p.category_id
ORDER BY p.id
LIMIT ? OFFSET ?;

-- name: CountProducts :one
SELECT count(*) FROM products;
