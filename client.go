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

func main() {
	client := http.Client{}
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080/cotacao", nil)
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

	var data interface{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile("cotacao.txt", []byte(fmt.Sprintf("Dólar: %s\n", data)), 0644)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Cotação do Dólar: %s\n", data)
}
