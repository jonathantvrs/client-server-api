package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type ServerResponse struct {
	USDBRL struct {
		Bid string `json:"bid"`
	} `json:"USDBRL"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "http://localhost:8080/cotacao", nil)
	if err != nil {
		log.Fatalf("Erro ao criar requisição: %v\n", err)
	}

	log.Println("Enviando requisição para o servidor...")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Println("Erro: Timeout de 300ms excedido ao aguardar a resposta do servidor.")
		}
		log.Fatalf("Erro ao fazer requisição: %v\n", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("Erro ao ler resposta: %v\n", err)
	}

	var serverResponse ServerResponse
	err = json.Unmarshal(body, &serverResponse)
	if err != nil {
		log.Fatalf("Erro ao decodificar resposta JSON: %v\n", err)
	}

	bid := serverResponse.USDBRL.Bid
	log.Printf("Cotação USD-BRL recebida: %s\n", bid)

	err = saveQuotationInFile(bid)
	if err != nil {
		log.Fatalf("Erro ao salvar cotação no arquivo: %v\n", err)
	}

	log.Println("Cotação salva com sucesso no arquivo cotacao.txt")
}

func saveQuotationInFile(bid string) error {
	file, err := os.Create("cotacao.txt")
	if err != nil {
		return err
	}
	defer file.Close()

	content := fmt.Sprintf("Dólar: %s", bid)
	_, err = file.WriteString(content)
	if err != nil {
		return err
	}

	return nil
}
