package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	_ "github.com/lib/pq"
)

type Income struct {
	Income1 string `json:"income1"`
	Income2 string `json:"income2"`
	Income3 string `json:"income3"`
}

func main() {
	connStr := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"), os.Getenv("DB_PORT"))

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, your database is connected!")
	})

	http.HandleFunc("/submit-income", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Unsupported method", http.StatusMethodNotAllowed)
			return
		}

		var income Income
		if err := json.NewDecoder(r.Body).Decode(&income); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		if err := insertIncome(db, income); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Fprintf(w, "Income data saved successfully")
	})

	log.Fatal(http.ListenAndServe(":8080", nil))
}

func insertIncome(db *sql.DB, income Income) error {
	_, err := db.Exec("INSERT INTO incomes (income1, income2, income3) VALUES ($1, $2, $3)", income.Income1, income.Income2, income.Income3)
	return err
}
