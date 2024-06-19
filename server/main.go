package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "modernc.org/sqlite"
)

const (
	apiURL     = "https://economia.awesomeapi.com.br/json/last/USD-BRL"
	dbTimeout  = 10 * time.Millisecond
	apiTimeout = 200 * time.Millisecond
	serverAddr = ":8080"
)

type Cotacao struct {
	USD struct {
		Bid string `json:"bid"`
	} `json:"USDBRL"`
}

func main() {
	db, err := sql.Open("sqlite", "./cotacoes.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := createTableIfNotExists(db); err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		if err := processaCotacao(w, r, db); err != nil {
			log.Println(err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
	})

	log.Printf("Server is running on port %s\n", serverAddr)
	log.Fatal(http.ListenAndServe(serverAddr, nil))
}

func createTableIfNotExists(db *sql.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS cotacao (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        bid TEXT,
        timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
    )`)
	return err
}

func processaCotacao(w http.ResponseWriter, r *http.Request, db *sql.DB) error {
	ctx, cancel := context.WithTimeout(r.Context(), apiTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL, nil)
	if err != nil {
		return err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var cotacao Cotacao
	if err := json.NewDecoder(resp.Body).Decode(&cotacao); err != nil {
		return err
	}

	bid := cotacao.USD.Bid

	if err := saveCotacaoToDB(db, bid); err != nil {
		return err
	}

	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(map[string]string{"bid": bid})
}

func saveCotacaoToDB(db *sql.DB, bid string) error {
	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, "INSERT INTO cotacao (bid) VALUES (?)", bid)
	if err != nil {
		return err
	}

	return tx.Commit()
}
