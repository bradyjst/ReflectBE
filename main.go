package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	// Import your sqlc generated package

	_ "github.com/lib/pq"
)

type Income struct {
	Income1 string `json:"income1"`
	Income2 string `json:"income2"`
	Income3 string `json:"income3"`
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

	queries := db.New(db) // Create an instance of sqlc queries

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

			err := queries.CreateIncome(context.TODO(), db.CreateIncomeParams{
				Income1: income.Income1,
				Income2: income.Income2,
				Income3: income.Income3,
			})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			fmt.Fprintf(w, "Income data saved successfully")
		}
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}
