-- name: CreateCustomer :one
INSERT INTO customers (email, name, phone, status, created_at)
VALUES (?, ?, ?, ?, ?)
RETURNING *;

-- name: CountCustomers :one
SELECT count(*) FROM customers;
