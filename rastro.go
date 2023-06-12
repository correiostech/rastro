package rastro

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type rastro struct {
	client *http.Client
	base   string
}

func (c *rastro) doReq(method, url string, token string, body io.Reader) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, fmt.Errorf("rastro doreq: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Connection", "close")
	if method == "POST" {
		req.Header.Set("Content-Type", "application/json")
	}
	res, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("rastro doreq: %v", err)
	}
	return res, nil
}

func New(urlBase string, token string) (rastro, error) {
	return rastro{
		client: &http.Client{
			Transport: &http.Transport{
				TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
				MaxIdleConns:        1000,
				MaxConnsPerHost:     1000,
				MaxIdleConnsPerHost: 1000,
			},
		},
		base: urlBase,
	}, nil
}

type Resultado struct {
	Objetos []struct {
		CodigoObjeto string `json:"codObjeto"`
		Eventos      []struct {
			Codigo    string `json:"codigo"`
			Tipo      string `json:"tipo"`
			Descricao string `json:"descricao"`
			DataHora  string `json:"dtHrCriado"`
			Unidade   struct {
				Nome      string `json:"nome"`
				CodigoSRO string `json:"codSro"`
				MCU       string `json:"codMcu"`
				SE        string `json:"se"`
			} `json:"unidade"`
		} `json:"eventos"`
	} `json:"objetos"`
}

type ResultadoAsync struct {
	User       string `json:"user"`
	Numero     string `json:"numero"`
	DtCriacao  string `json:"dtCriacao"`
	DtValidade string `json:"dtValidade"`
	QtdObjetos int    `json:"qtdObjetos"`
	Resultado  string `json:"resultado"`
	Idioma     string `json:"idioma"`
}

func (c *rastro) Rastreia(objetos string, token string) (Resultado, error) {
	var result Resultado
	params := url.Values{}
	params.Add("resultado", "U")
	codigosObjetos := strings.Split(objetos, ",")
	for _, codigo := range codigosObjetos {
		params.Add("codigosObjetos", codigo)
	}
	res, err := c.doReq("GET", c.base+"?"+params.Encode(), token, nil)
	if err != nil {
		return result, fmt.Errorf("rastro-rs objetos: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return result, fmt.Errorf("erro ao rastrear o objeto: " + res.Status)
	}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return result, fmt.Errorf("rastro-rs rastros: %v", err)
	}
	return result, nil
}

func (c *rastro) RastreiaAsync(objetos []string, token string) (ResultadoAsync, error) {
	var result ResultadoAsync
	objetosJSON, err := json.Marshal(objetos)
	if err != nil {
		return result, fmt.Errorf("rastro-async objetos: %v", err)
	}
	res, err := c.doReq("POST", c.base, token, strings.NewReader(string(objetosJSON)))
	if err != nil {
		return result, fmt.Errorf("rastro-async objetos: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != 202 {
		return result, fmt.Errorf("erro ao registrar objetos para rastro: " + res.Status)
	}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return result, fmt.Errorf("rastro-async decode: %v", err)
	}
	return result, nil
}

func (c *rastro) Recibo(recibo string, token string) (Resultado, error) {
	var result Resultado

	res, err := c.doReq("GET", c.base+recibo, token, nil)
	if err != nil {
		return result, fmt.Errorf("rastro-rs objetos: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return result, fmt.Errorf("erro ao rastrear o objeto: " + res.Status)
	}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return result, fmt.Errorf("rastro-rs rastros: %v", err)
	}
	return result, nil
}
