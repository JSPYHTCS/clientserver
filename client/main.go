package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

const (
	serverURL    = "http://localhost:8080/cotacao"
	clientTimeout = 300 * time.Millisecond
	outputFile   = "cotacao.txt"
)

type CotacaoResponse struct {
	Bid string `json:"bid"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), clientTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", serverURL, nil)
	if err != nil {
		log.Fatal("Error creating request:", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal("Error making request:", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("Unexpected status code: %d", resp.StatusCode)
	}

	var cotacao CotacaoResponse
	if err := json.NewDecoder(resp.Body).Decode(&cotacao); err != nil {
		log.Fatal("Error decoding response:", err)
	}

	if err := saveCotacaoToFile(cotacao.Bid); err != nil {
		log.Fatal("Error saving to file:", err)
	}

	log.Println("Cotação salva em", outputFile)
}

func saveCotacaoToFile(bid string) error {
	data := []byte("Dólar: " + bid)
	if err := ioutil.WriteFile(outputFile, data, 0644); err != nil {
		return err
	}
	return nil
}
