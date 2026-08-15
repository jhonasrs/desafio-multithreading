package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

type Address struct {
	CEP          string `json:"cep"`
	Street       string `json:"street"`
	Neighborhood string `json:"neighborhood"`
	City         string `json:"city"`
	State        string `json:"state"`
	API          string `json:"api"`
}

type Result struct {
	Address Address
	Error   error
	API     string
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Uso: go run main.go <CEP>")
		fmt.Println("Exemplo: go run main.go 01153000")
		os.Exit(1)
	}

	cep := os.Args[1]

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	ch := make(chan Result, 2)

	go fetchBrasilAPI(ctx, cep, ch)
	go fetchViaCEP(ctx, cep, ch)

	// Condição de corrida: aguarda pela resposta bem-sucedida mais rápida
	for i := 0; i < 2; i++ {
		select {
		case res := <-ch:
			if res.Error == nil {
				printAddress(res.Address, res.API)
				return
			}
			fmt.Printf("Aviso: %s falhou: %v\n", res.API, res.Error)
		case <-ctx.Done():
			fmt.Println("Erro: Tempo limite de 1 segundo excedido (nenhuma API respondeu em até 1s).")
			return
		}
	}

	fmt.Println("Erro: Ambas as APIs falharam ao retornar um endereço válido.")
}

func fetchBrasilAPI(ctx context.Context, cep string, ch chan<- Result) {
	url := fmt.Sprintf("https://brasilapi.com.br/api/cep/v1/%s", cep)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		ch <- Result{Error: err, API: "BrasilAPI"}
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		ch <- Result{Error: err, API: "BrasilAPI"}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		ch <- Result{Error: fmt.Errorf("código de status %d", resp.StatusCode), API: "BrasilAPI"}
		return
	}

	var raw struct {
		CEP          string `json:"cep"`
		State        string `json:"state"`
		City         string `json:"city"`
		Neighborhood string `json:"neighborhood"`
		Street       string `json:"street"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		ch <- Result{Error: err, API: "BrasilAPI"}
		return
	}

	ch <- Result{
		Address: Address{
			CEP:          raw.CEP,
			Street:       raw.Street,
			Neighborhood: raw.Neighborhood,
			City:         raw.City,
			State:        raw.State,
			API:          "BrasilAPI",
		},
		API: "BrasilAPI",
	}
}

func fetchViaCEP(ctx context.Context, cep string, ch chan<- Result) {
	url := fmt.Sprintf("http://viacep.com.br/ws/%s/json/", cep)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		ch <- Result{Error: err, API: "ViaCEP"}
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		ch <- Result{Error: err, API: "ViaCEP"}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		ch <- Result{Error: fmt.Errorf("código de status %d", resp.StatusCode), API: "ViaCEP"}
		return
	}

	var raw struct {
		CEP         string      `json:"cep"`
		Logradouro  string      `json:"logradouro"`
		Complemento string      `json:"complemento"`
		Bairro      string      `json:"bairro"`
		Localidade  string      `json:"localidade"`
		UF          string      `json:"uf"`
		Erro        interface{} `json:"erro"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		ch <- Result{Error: err, API: "ViaCEP"}
		return
	}

	// ViaCEP retorna "erro": true quando o CEP não é encontrado
	if errBool, ok := raw.Erro.(bool); ok && errBool {
		ch <- Result{Error: fmt.Errorf("cep não encontrado"), API: "ViaCEP"}
		return
	}

	ch <- Result{
		Address: Address{
			CEP:          raw.CEP,
			Street:       raw.Logradouro,
			Neighborhood: raw.Bairro,
			City:         raw.Localidade,
			State:        raw.UF,
			API:          "ViaCEP",
		},
		API: "ViaCEP",
	}
}

func printAddress(addr Address, apiName string) {
	fmt.Printf("\n=== Resposta mais rápida recebida de: %s ===\n", apiName)
	fmt.Printf("CEP:          %s\n", addr.CEP)
	fmt.Printf("Logradouro:   %s\n", addr.Street)
	fmt.Printf("Bairro:       %s\n", addr.Neighborhood)
	fmt.Printf("Cidade:       %s\n", addr.City)
	fmt.Printf("Estado:       %s\n", addr.State)
	fmt.Println("==========================================")
}
