package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

// Estruturas para armazenar os dados retornados pelas APIs
type BrasilAPIResponse struct {
	Cep          string `json:"cep"`
	State        string `json:"state"`
	City         string `json:"city"`
	Neighborhood string `json:"neighborhood"`
	Street       string `json:"street"`
}

type ViaCEPResponse struct {
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	Uf          string `json:"uf"`
}

func fetchFromBrasilAPI(cep string, wg *sync.WaitGroup, ch chan<- interface{}) {
	defer wg.Done()
	url := fmt.Sprintf("https://brasilapi.com.br/api/cep/v1/%s", cep)
	client := &http.Client{Timeout: 1 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		ch <- err
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		ch <- err
		return
	}

	var result BrasilAPIResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		ch <- err
		return
	}
	ch <- result
}

func fetchFromViaCEP(cep string, wg *sync.WaitGroup, ch chan<- interface{}) {
	defer wg.Done()
	url := fmt.Sprintf("http://viacep.com.br/ws/%s/json/", cep)
	client := &http.Client{Timeout: 1 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		ch <- err
		return
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		ch <- err
		return
	}

	var result ViaCEPResponse
	err = json.Unmarshal(body, &result)
	if err != nil {
		ch <- err
		return
	}
	ch <- result
}

func main() {
	cep := "01153000"
	ch := make(chan interface{})
	var wg sync.WaitGroup
	wg.Add(2)

	go fetchFromBrasilAPI(cep, &wg, ch)
	go fetchFromViaCEP(cep, &wg, ch)

	go func() {
		wg.Wait()
		close(ch)
	}()

	select {
	case result := <-ch:
		switch res := result.(type) {
		case BrasilAPIResponse:
			fmt.Printf("Resposta da BrasilAPI: %+v\n", res)
		case ViaCEPResponse:
			fmt.Printf("Resposta da ViaCEP: %+v\n", res)
		case error:
			fmt.Printf("Erro: %s\n", res.Error())
		}
	case <-time.After(1 * time.Second):
		fmt.Println("Erro: Timeout")
	}
}
