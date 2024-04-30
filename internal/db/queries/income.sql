-- name: GetIncomes :many
SELECT * FROM incomes;

-- name: CreateIncome :exec
INSERT INTO incomes (income1, income2, income3) VALUES ($1, $2, $3);
