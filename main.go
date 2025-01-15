package main

import (
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type Men struct {
	Rank    *int   `json:"rank"` // Using a pointer allows for null ranks in the database
	Name    string `json:"name"`
	Country string `json:"country"`
}

// User struct for storing user credentials
type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Claims struct for JWT claims
type Claims struct {
	Username string `json:"username"`
	jwt.StandardClaims
}

var jwtKey = []byte("secret_key")

func main() {
	// connection to databse
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// check if table exists during setup
	// Efficiently check if table exists
	var dbExists bool
	var usersTableExists bool

	err = db.QueryRow("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'men')").Scan(&dbExists)
	if err != nil {
		log.Fatal(err)
	}

	err = db.QueryRow("SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_name = 'users')").Scan(&usersTableExists)
	if err != nil {
		log.Fatal(err)
	}

	if !dbExists { // create table if it does not exist
		createMenTable(db)
	}

	if !usersTableExists { // create the 'users' table if it doesn exist
		createUsersTable(db)
	}

	// JWT Key management
	encodedKey := os.Getenv("JWT_KEY") // Should consider Google Cloud KMS in production

	if encodedKey == "" {
		// generate and store key if not found
		jwtKey := generateJWTKey()
		encodingKey := base64.StdEncoding.EncodeToString(jwtKey)                  // encoding for storage
		fmt.Println("Generated and stored JWT Key (Base64 Encoded)", encodingKey) // for DEMO only - No log in production
	} else {
		// retrieve and decode the key
		var err error
		jwtKey, err = base64.StdEncoding.DecodeString(encodedKey)
		if err != nil {
			log.Fatal("Error decoding JWT Key", err)
		}
	}

	// create rounter
	router := mux.NewRouter()
	router.Use(jsonContentTypeMiddleware)

	router.HandleFunc("/login", loginHandler(db)).Methods("POST") // New login route
	router.Use(authMiddleware)

	router.HandleFunc("/men", getMen(db)).Methods("GET")
	router.HandleFunc("/men/{rank}", getMenByRank(db)).Methods("GET")
	router.HandleFunc("/men", createUser(db)).Methods("POST")
	router.HandleFunc("/men/{rank}", updateUser(db)).Methods("PUT")
	router.HandleFunc("/men/{rank}", deleteUser(db)).Methods("DELETE")

	// start server
	if err := http.ListenAndServe(":8000", router); err != nil {
		log.Fatal(err) // log and exit on server startup error
	}

	defer func() { // check if database connection fails to close
		if err := db.Close(); err != nil {
			log.Printf("Error closing database connection: %w", err)
		}
	}()

}
func jsonContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func generateJWTKey() []byte {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		log.Fatal("Error generating KWT Key", err)
		return nil
	}
	return key
}

func createUsersTable(db *sql.DB) {
	_, err := db.Exec(`
			CREATE TABLE users (
				username TEXT PRIMARY KEY,
				password TEXT NOT NULL)
		
		`)
	if err != nil {
		log.Fatal(err)
	}
}

func createMenTable(db *sql.DB) {
	_, err := db.Exec("CREATE TABLE men (rank INT PRIMARY KEY, name TEXT, country TEXT)")
	if err != nil {
		log.Fatal(err)
	}
}
