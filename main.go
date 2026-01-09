package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Message struct {
	APIName string
	Body    string
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Erro: Número de argumentos inválido.\n")
		os.Exit(1)
	}
	cep := os.Args[1]

	c := make(chan Message, 2)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Goroutine 1: BrasilAPI
	go func() {
		url := "https://brasilapi.com.br/api/cep/v1/" + cep
		requestAPI(ctx, url, "BrasilAPI", c)
	}()

	// Goroutine 2: ViaCEP
	go func() {
		url := "http://viacep.com.br/ws/" + cep + "/json/"
		requestAPI(ctx, url, "ViaCEP", c)
	}()

	select {
	case msg := <-c:
		fmt.Printf("--- Resposta Recebida ---\n")
		fmt.Printf("Fonte: %s\n", msg.APIName)
		fmt.Printf("Dados: %s\n", msg.Body)

	case <-ctx.Done():
		fmt.Fprintf(os.Stderr, "Erro: Timeout atingido. Nenhuma API respondeu em 1 segundo.\n")
	}
}

func requestAPI(ctx context.Context, url, apiName string, c chan<- Message) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	c <- Message{
		APIName: apiName,
		Body:    string(body),
	}
}
