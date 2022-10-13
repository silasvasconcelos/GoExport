package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

const API_URL = "http://localhost:8080/cotacao"
const OUTPUT_FILE = "cotacao.txt"

type Coin struct {
	Bid float64 `json:"bid"`
}

func main() {
	// Note: This was the minimum time that the API returned the information
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*600)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, API_URL, nil)
	if err != nil {
		panic(err)
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		panic(err)
	}
	defer res.Body.Close()
	var quote Coin
	err = json.Unmarshal(body, &quote)
	if err != nil {
		panic(err)
	}

	f, err := os.Create(OUTPUT_FILE)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	result := fmt.Sprintf("Dólar: %v", quote.Bid)
	io.WriteString(f, result)
	fmt.Println(result)
}
