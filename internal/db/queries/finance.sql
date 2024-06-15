-- name: GetFinanceForUser :many
SELECT * FROM finances
WHERE type = 'income' AND user_id = $1;

-- name: CreateFinance :exec
INSERT INTO finances (user_id, type, source, amount, date, description)
VALUES ($1, 'income', $2, $3, $4, $5);

