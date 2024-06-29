package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/bradyjst/reflectBE/internal/db"
	"github.com/bradyjst/reflectBE/internal/handlers"
	"github.com/bradyjst/reflectBE/internal/middleware"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

func main() {
	connStr := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		os.Getenv("DB_HOST"), os.Getenv("DB_USER"), os.Getenv("DB_PASSWORD"), os.Getenv("DB_NAME"), os.Getenv("DB_PORT"))

	dbConn, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer dbConn.Close()

	if err := db.ApplyMigrations(dbConn, connStr); err != nil {
		log.Fatalf("Failed to apply database migrations: %v", err)
	}

	queries := db.NewQueries(dbConn) // Create an instance of sqlc queries

	r := mux.NewRouter()

	r.HandleFunc("/register", handlers.RegisterHandler(queries)).Methods("POST", "OPTIONS")
	r.HandleFunc("/login", handlers.LoginHandler(queries)).Methods("POST", "OPTIONS")

	protected := r.PathPrefix("/api").Subrouter()
	protected.Use(middleware.ValidateJWT)

	r.HandleFunc("/submit-finance", handlers.SubmitFinanceHandler(queries)).Methods("POST", "OPTIONS")

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, your database is connected!")
	}).Methods("GET")

	http.Handle("/", middleware.EnableCors(r))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
