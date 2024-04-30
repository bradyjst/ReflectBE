package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/bradyjst/reflectBE/internal/db/"
	mydb "github.com/bradyjst/reflectBE/internal/db/sqlcgen"
	_ "github.com/lib/pq"
)

type Income struct {
	Income1 string `json:"income1"`
	Income2 string `json:"income2"`
	Income3 string `json:"income3"`
}

func NewNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Content-Type")
}

func handleOptions(w http.ResponseWriter, r *http.Request) {
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
}

func main() {
	connStr := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"), os.Getenv("DB_PORT"))

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := mydb.ApplyMigrations(db, connStr); err != nil {
		log.Fatalf("Failed to apply database migrations: %v", err)
	}

	queries := mydb.New(db) // Create an instance of sqlc queries

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, your database is connected!")
	})

	http.HandleFunc("/submit-income", func(w http.ResponseWriter, r *http.Request) {
		enableCors(&w)
		handleOptions(w, r)
		if r.Method == "POST" {
			defer r.Body.Close()
			var income Income
			if err := json.NewDecoder(r.Body).Decode(&income); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			// params := mydb.CreateIncomeParams{
			// 	Income1: NewNullString(income.Income1),
			// 	Income2: NewNullString(income.Income2),
			// 	Income3: NewNullString(income.Income3),
			// }

			err := queries.CreateIncome(context.TODO(), params)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			fmt.Fprintf(w, "Income data saved successfully")
		}
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
