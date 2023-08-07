package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Usdbrl struct {
	Code       string `json:"code"`
	Codein     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
}

type ExchangeRate struct {
	Usdbrl `json:"USDBRL"`
}

type Exchange struct {
	ID           int `gorm:"primaryKey"`
	ExchangeRate `gorm:"ResponseAPI"`
}

func main() {
	http.HandleFunc("/cotacao", handleQuotation)
	log.Println("Server listening on port :8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		panic(err)
	}
}

func handleQuotation(w http.ResponseWriter, r *http.Request) {
	data, err := fetchExchangeRate()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = insertPrice(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	err = json.NewEncoder(w).Encode(data)
	if err != nil {
		panic(err)
	}
}

func fetchExchangeRate() (*ExchangeRate, error) {
	client := http.Client{}
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		panic(err)
	}

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var data ExchangeRate
	err = json.Unmarshal(body, &data)
	if err != nil {
		panic(err)
	}

	return &data, nil
}

func connectToDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("cotacao.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(&Exchange{}, &ExchangeRate{})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func insertPrice(price *ExchangeRate) error {
	db, err := connectToDB()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	db.WithContext(ctx).Create(&Exchange{
		ExchangeRate: *price})

	return nil
}
