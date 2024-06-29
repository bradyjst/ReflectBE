package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"

	migratedb "github.com/bradyjst/reflectBE/internal/db"
	mydb "github.com/bradyjst/reflectBE/internal/db/sqlcgen"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

type Finance struct {
	UserID      int32     `json:"user_id"`
	Type        string    `json:"type"` // 'income' or 'expense'
	Source      string    `json:"source"`
	Amount      string    `json:"amount"`
	Date        time.Time `json:"date,omitempty"` // Optional, defaults to CURRENT_TIMESTAMP if not provided
	Description string    `json:"description,omitempty"`
}

type User struct {
	Username string `json:"username"`
	Password string `json:"password_hash"`
	Email    string `json:"email"`
}

var jwtKey = []byte("my_secret_key")

func generateJWT(username string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &jwt.StandardClaims{
		Subject:   username,
		ExpiresAt: expirationTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil

}

func validateJWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Missing Authorization Header", http.StatusUnauthorized)
			return
		}

		tokenString := authHeader[len("Bearer "):]

		claims := &jwt.StandardClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "username", claims.Subject)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func NewNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

func NewNullTime(t time.Time) sql.NullTime {
	if t.IsZero() {
		return sql.NullTime{Valid: false} // Return a NullTime with Valid set to false if time is zero
	}
	return sql.NullTime{Time: t, Valid: true} // Otherwise, return the time with Valid set to true
}

func enableCors(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func registerHandler(queries *mydb.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		defer r.Body.Close()
		var user User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		hashedPassword, err := hashPassword(user.Password)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		err = queries.CreateUser(context.TODO(), mydb.CreateUserParams{
			Username:     user.Username,
			PasswordHash: hashedPassword,
			Email:        user.Email,
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		fmt.Fprintf(w, "User account created!")
	}
}

func loginHandler(queries *mydb.Queries) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		defer r.Body.Close()
		var user User
		if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
			log.Printf("Error decoding request body: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		dbUser, err := queries.GetUserByUsername(context.TODO(), user.Username)
		if err != nil {
			log.Printf("Error fetching user: %v", err)
			http.Error(w, "Invalid username or password", http.StatusUnauthorized)
			return
		}

		if !checkPasswordHash(user.Password, dbUser.PasswordHash) {
			log.Printf("Invalid password for user: %s", user.Username)
			http.Error(w, "Invalid username r password", http.StatusUnauthorized)
		}

		token, err := generateJWT(user.Username)
		if err != nil {
			log.Printf("Error generating token: %v", err)
			http.Error(w, "Error generating token", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"token": token})

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

	if err := migratedb.ApplyMigrations(db, connStr); err != nil {
		log.Fatalf("Failed to apply database migrations: %v", err)
	}

	queries := mydb.New(db) // Create an instance of sqlc queries

	r := mux.NewRouter()

	r.HandleFunc("/register", registerHandler(queries)).Methods("POST", "OPTIONS")
	r.HandleFunc("/login", loginHandler(queries)).Methods("POST", "OPTIONS")

	protected := r.PathPrefix("/api").Subrouter()
	protected.Use(validateJWT)

	r.HandleFunc("/submit-finance", func(w http.ResponseWriter, r *http.Request) {
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
		params := mydb.CreateFinanceParams{
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

		fmt.Fprintf(w, "Income data saved successfully")
	}).Methods("POST", "OPTIONS")

	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, your database is connected!")
	}).Methods("GET")

	http.Handle("/", enableCors(r))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
