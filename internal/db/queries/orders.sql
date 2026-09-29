-- name: CreateOrder :one
INSERT INTO orders (customer_id, status, total, created_at, updated_at)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: CreateOrderItem :exec
INSERT INTO order_items (order_id, product_id, quantity, unit_price)
VALUES (?, ?, ?, ?);

-- name: ListRecentOrders :many
SELECT o.id, o.status, o.total, o.created_at,
       c.id AS customer_id, c.name AS customer_name
FROM orders o
JOIN customers c ON c.id = o.customer_id
ORDER BY o.created_at DESC
LIMIT ?;

-- Daily revenue excluding cancelled/refunded orders (dashboard chart).
-- name: RevenueByDay :many
SELECT CAST(date(created_at) AS TEXT)      AS day,
       count(*)                            AS orders,
       CAST(COALESCE(sum(total), 0) AS INTEGER) AS revenue
FROM orders
WHERE status NOT IN ('cancelled', 'refunded')
  AND created_at >= sqlc.arg(since)
GROUP BY day
ORDER BY day;
