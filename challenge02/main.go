package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"time"
)

type Clasure func()

type addressViaCep struct {
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	Uf          string `json:"uf"`
	Ibge        string `json:"ibge"`
	Gia         string `json:"gia"`
	Ddd         string `json:"ddd"`
	Siafi       string `json:"siafi"`
}

type addressApiCep struct {
	Code     string `json:"code"`
	State    string `json:"state"`
	City     string `json:"city"`
	District string `json:"district"`
	Address  string `json:"address"`
}

type Address struct {
	Address string
	State   string
	City    string
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Por favor, informe um CEP")
		return
	}

	zipCode := ZipCodeFormat(os.Args[1])
	if len(zipCode) < 9 {
		fmt.Println("CEP invalido")
		return
	}
	addrFromApiCep := make(chan Address)
	cancelApiCepCtx, cancelApiCepHandlerFubc := context.WithCancel(context.Background())
	addrFromViaCep := make(chan Address)
	cancelViaCepCtx, cancelViaCepHandlerFubc := context.WithCancel(context.Background())

	go func() {
		addr, err := GetAddressFromAPICep(zipCode, cancelApiCepCtx)
		if err != nil {
			fmt.Println(err)
		}
		addrFromApiCep <- addr
	}()

	go func() {
		addr, err := GetAddressFromViaCep(zipCode, cancelViaCepCtx)
		if err != nil {
			fmt.Println(err)
		}
		addrFromViaCep <- addr
	}()

	select {
	case addr := <-addrFromApiCep:
		cancelViaCepHandlerFubc()
		PrintAddr(addr, "ApiCep")
	case addr := <-addrFromViaCep:
		cancelApiCepHandlerFubc()
		PrintAddr(addr, "ViaCep")
	case <-time.After(1 * time.Second):
		fmt.Println(errors.New("Timeout"))
	}
}

// GetAddressFromAPICep - Get address data from https://apicep.com/
func GetAddressFromAPICep(zipCode string, cancelCtx context.Context) (Address, error) {
	url := fmt.Sprintf("https://cdn.apicep.com/file/apicep/%s.json", zipCode)
	req, err := http.NewRequestWithContext(cancelCtx, "GET", url, nil)
	if err != nil {
		return Address{}, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Address{}, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Address{}, err
	}
	defer resp.Body.Close()
	apiCepAddr := addressApiCep{}
	err = json.Unmarshal(body, &apiCepAddr)
	if err != nil {
		return Address{}, err
	}
	addr := Address{}
	addr.Address = apiCepAddr.Address
	addr.State = apiCepAddr.State
	addr.City = apiCepAddr.City
	return addr, nil
}

func GetAddressFromViaCep(zipCode string, cancelCtx context.Context) (Address, error) {
	url := fmt.Sprintf("https://viacep.com.br/ws/%s/json/", zipCode)
	req, err := http.NewRequestWithContext(cancelCtx, "GET", url, nil)
	if err != nil {
		return Address{}, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return Address{}, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return Address{}, err
	}
	defer resp.Body.Close()
	viaCepAddr := addressViaCep{}
	err = json.Unmarshal(body, &viaCepAddr)
	if err != nil {
		return Address{}, err
	}
	addr := Address{}
	addr.Address = viaCepAddr.Logradouro
	addr.State = viaCepAddr.Uf
	addr.City = viaCepAddr.Localidade
	return addr, nil
}

func PrintAddr(addr Address, api string) {
	fmt.Printf("Resultado da API: %s\n--------------------------\n", api)
	fmt.Printf("Endereço: %s\nCidade: %s\nEstado: %s\n", addr.Address, addr.City, addr.State)
}

func ZipCodeFormat(zipCode string) string {
	rgxClear, _ := regexp.Compile("\\D")
	zipCode = rgxClear.ReplaceAllString(zipCode, "")
	rgxMask, _ := regexp.Compile("([0-9]{5})([0-9]{3})")
	zipCode = rgxMask.ReplaceAllString(zipCode, "$1-$2")
	return zipCode
}
