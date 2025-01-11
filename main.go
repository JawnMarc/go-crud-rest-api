package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type Men struct {
	Rank    *int   `json:"rank"`
	Name    string `json:"name"`
	Country string `json:"country"`
}

// get all men
func getMen(db *sql.DB) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// query database
		rows, err := db.Query("SELECT rank, name, country FROM men")
		if err != nil {
			// log.Fatal(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rows.Close() // close rows when function ends

		var men []Men // slice of men

		for rows.Next() { // loop through rows
			var man Men // instance
			// scan rows into men instance
			err := rows.Scan(&man.Rank, &man.Name, &man.Country)
			if err != nil {
				// log.Fatal(err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			men = append(men, man)
		}
		// check for errors
		if err := rows.Err(); err != nil {
			// log.Fatal(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		// encode
		json.NewEncoder(w).Encode(men)
	}
}

func getMenByRank(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rankStr := mux.Vars(r)["rank"]

		rank, err := strconv.Atoi(rankStr) //convert rank to int
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		row := db.QueryRow("SELECT rank, name, country FROM men WHERE rank = $1", rank)

		var men Men
		if err := row.Scan(&men.Rank, &men.Name, &men.Country); err != nil {
			if err == sql.ErrNoRows {
				http.Error(w, "Men not found", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(men)
	}
}

// // get men by user id
// func getMenByID(db *sql.DB) http.HandlerFunc {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		id := mux.Vars(r)["id"]
// 		// query database
// 		row := db.QueryRow("SELECT rank, name, country FROM men WHERE id =$1", id)
// 		var men Men // instance
// 		// scan row into
// 		err := row.Scan(&men.Rank, &men.Name, &men.Country)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		// encode
// 		json.NewEncoder(w).Encode(men)
// 	}
// }

// create new user
func createUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var men Men
		err := json.NewDecoder(r.Body).Decode(&men)
		if err != nil {
			// log.Fatal(err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		// insert into database
		_, err = db.Exec("INSERT INTO men (rank, name, country) VALUES ($1, $2, $3)", men.Rank, men.Name, men.Country)
		if err != nil {
			// log.Fatal(err)
			// http.Error(w, err.Error(), http.StatusBadRequest)
			http.Error(w, err.Error(), http.StatusConflict)
			return
		}
		json.NewEncoder(w).Encode(men)
	}
}

// update user
func updateUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rankStr := mux.Vars(r)["rank"]

		rank, err := strconv.Atoi(rankStr)
		if err != nil {
			http.Error(w, "Invalid rank parameter", http.StatusBadRequest)
			return
		}

		var men Men

		err = json.NewDecoder(r.Body).Decode(&men)
		if err != nil {
			// log.Fatal(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		_, err = db.Exec("UPDATE men SET name = $1, country = $2 WHERE rank = $3", men.Name, men.Country, rank)
		if err != nil {
			// log.Fatal(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		json.NewEncoder(w).Encode(men)
	}
}

// delete user
func deleteUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rankStr := mux.Vars(r)["rank"]
		rank, err := strconv.Atoi(rankStr)
		if err != nil {
			http.Error(w, "Invalid rank parameter", http.StatusBadRequest)
			return
		}

		_, err = db.Exec("DELETE FROM men WHERE rank = $1", rank)
		if err != nil {
			// log.Fatal(err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		json.NewEncoder(w).Encode(rank)

	}
}

func main() {
	// connection to databse
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// create table if it does not exist
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS men(rank int, name text, country text)")
	if err != nil {
		log.Fatal(err)
	}

	// create rounter
	router := mux.NewRouter()
	router.Use(jsonContentTypeMiddleware)

	router.HandleFunc("/men", getMen(db)).Methods("GET")
	// router.HandleFunc("/men/{id}", getMenByID(db)).Methods("GET")
	router.HandleFunc("/men/{rank}", getMenByRank(db)).Methods("GET")
	router.HandleFunc("/men", createUser(db)).Methods("POST")
	router.HandleFunc("/men/{rank}", updateUser(db)).Methods("PUT")
	router.HandleFunc("/men/{rank}", deleteUser(db)).Methods("DELETE")

	// start server
	log.Fatal(http.ListenAndServe(":8000", router))

}
func jsonContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}
