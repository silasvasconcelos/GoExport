package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"io"
	"net/http"
	"time"
)

type Coin struct {
	Id         int     `json:"id"`
	Code       string  `json:"code"`
	Codein     string  `json:"codein"`
	Name       string  `json:"name"`
	High       float64 `json:"high,string"`
	Low        float64 `json:"low,string"`
	VarBid     float64 `json:"varBid,string"`
	PctChange  float64 `json:"pctChange,string"`
	Bid        float64 `json:"bid,string"`
	Ask        float64 `json:"ask,string"`
	Timestamp  string  `json:"timestamp"`
	CreateDate string  `json:"create_date"`
}

type Quote struct {
	USDBRL Coin
}

type QuoteResp struct {
	Bid float64 `json:"bid"`
}

type Error struct {
	Error string `json:"error"`
}

const API_URL string = "https://economia.awesomeapi.com.br/json/last/USD-BRL"

var (
	dbCon *sql.DB
)

func main() {
	db, err := sql.Open("sqlite3", "./database.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()
	dbCon = db
	createDatabase(db)

	http.HandleFunc("/cotacao", quoteHandle)

	http.ListenAndServe(":8080", nil)
}

func quoteHandle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	cot, err := getQuote()
	if err != nil {
		showError(err, w)
		return
	}
	err = insertQuote(cot.USDBRL)
	if err != nil {
		showError(err, w)
		return
	}
	res := QuoteResp{
		Bid: cot.USDBRL.Bid,
	}
	jsonRes, err := json.Marshal(res)
	if err != nil {
		showError(err, w)
		return
	}
	fmt.Fprintf(w, string(jsonRes))
}

func getQuote() (*Quote, error) {
	// Not: This was the minimum time that the API returned the information
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*600)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, API_URL, nil)
	if err != nil {
		return nil, err
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return nil, err
	}
	var quote Quote
	err = json.Unmarshal(body, &quote)
	if err != nil {
		return nil, err
	}
	return &quote, nil
}

func createDatabase(db *sql.DB) {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS coins (
    	id INTEGER PRIMARY KEY   AUTOINCREMENT,
		code TEXT,
		codein TEXT,
		name TEXT,
		high DECIMAL,
		low DECIMAL,
		var_bid DECIMAL,
		pct_change BIGINT,
		bid DECIMAL,
		ask DECIMAL,
		timestamp TEXT,
		create_date TEXT
	);
	`)
	if err != nil {
		panic(err)
	}
}

func insertQuote(coin Coin) error {
	sql := "insert into coins (code, codein, name, high, low, var_bid, pct_change, bid, ask, timestamp, create_date) values (?,?,?,?,?,?,?,?,?,?,?);"
	stmt, err := dbCon.Prepare(sql)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*10)
	defer cancel()
	_, err = stmt.ExecContext(ctx, coin.Code, coin.Codein, coin.Name, coin.High, coin.Low, coin.VarBid, coin.PctChange, coin.Bid, coin.Ask, coin.Timestamp, coin.CreateDate)
	if err != nil {
		fmt.Println(err.Error())
		return err
	}
	return nil
}

func showError(err error, w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	resErr, jmErr := json.Marshal(Error{
		Error: err.Error(),
	})
	if jmErr != nil {
		panic(jmErr)
	}
	fmt.Fprintf(w, string(resErr))
}
