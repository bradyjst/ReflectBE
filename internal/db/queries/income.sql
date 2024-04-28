-- db/queries/income.sql
-- name: CreateIncome :exec
INSERT INTO incomes (income1, income2, income3) VALUES ($1, $2, $3);

-- name: ListIncomes :many
SELECT * FROM incomes;
