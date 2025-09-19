package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type QuotationResponse struct {
	USDBRL struct {
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
	} `json:"USDBRL"`
}

func main() {
	db, err := sql.Open("sqlite3", "./quotation.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	createQuotationTableSQL := `CREATE TABLE IF NOT EXISTS quotations (
		ID INTEGER PRIMARY KEY AUTOINCREMENT,
		bid TEXT,
		timestamp TEXT
	);`
	_, err = db.Exec(createQuotationTableSQL)
	if err != nil {
		panic(err)
	}

	http.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		quotationHandler(w, r, db)
	})

	log.Println("Servidor iniciado na porta 8080")
	http.ListenAndServe(":8080", nil)
}

func quotationHandler(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	log.Println("Requisição recebida em /cotacao")

	quotationData, err := GetQuotationUsdBrl()
	if err != nil {
		log.Printf("Erro ao obter cotação: %v\n", err)
		http.Error(w, "Erro ao obter cotação", http.StatusInternalServerError)
		return
	}

	err = SaveQuotation(db, quotationData.USDBRL.Bid)
	if err != nil {
		log.Printf("Erro ao salvar cotação: %v\n", err)
		http.Error(w, "Erro ao salvar cotação", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	json.NewEncoder(w).Encode(quotationData)
	log.Println("Resposta enviada com sucesso.")
}

func GetQuotationUsdBrl() (*QuotationResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return nil, err
	}

	log.Println("Buscando cotação da API externa...")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Println("Erro: Timeout de 200ms excedido ao buscar a cotação.")
		}
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var quotation QuotationResponse
	err = json.Unmarshal(body, &quotation)
	if err != nil {
		return nil, err
	}

	log.Printf("Cotação recebida da API: %s\n", quotation.USDBRL.Bid)
	return &quotation, nil
}

func SaveQuotation(db *sql.DB, bid string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	stmt, err := db.Prepare("INSERT INTO quotations(bid, timestamp) VALUES(?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	log.Println("Salvando cotação no banco de dados...")
	_, err = stmt.ExecContext(ctx, bid, time.Now().Format(time.RFC3339))
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			log.Println("Erro: Timeout de 10ms excedido ao salvar no banco de dados.")
		}
		return err
	}

	log.Println("Cotação salva com sucesso.")
	return nil
}
