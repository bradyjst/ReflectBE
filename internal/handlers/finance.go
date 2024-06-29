package handlers

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/bradyjst/reflectBE/internal/db/"
)

type Finance struct {
	UserID      int32     `json:"user_id"`
	Type        string    `json:"type"`
	Source      string    `json:"source"`
	Amount      string    `json:"amount"`
	Date        time.Time `json:"date,omitempty"`
	Description string    `json:"description,omitempty"`
}

func SubmitFinanceHandler(queries *db.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		defer r.Body.Close()
		var finance Finance
		if err := json.NewDecoder(r.Body).Decode(&finance); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		userID := 1 // Placeholder for now, assuming user is authenticated
		params := db.CreateFinanceParams{
			UserID:      int32(userID),
			Source:      NewNullString(finance.Source),
			Amount:      NewNullString(finance.Amount),
			Date:        NewNullTime(finance.Date),
			Description: NewNullString(finance.Description),
		}
		err := queries.CreateFinance(context.TODO(), params)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "Income data saved successfully")
	}
}

func NewNullTime(t time.Time) sql.NullTime {
	if t.IsZero() {
		return sql.NullTime{Valid: false} // Return a NullTime with Valid set to false if time is zero
	}
	return sql.NullTime{Time: t, Valid: true} // Otherwise, return the time with Valid set to true
}

func NewNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}
