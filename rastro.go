package rastro

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type rastro struct {
	client *http.Client
	base   string
}

func (c *rastro) doReq(method, url string, token string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, fmt.Errorf("cep-rs doreq: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Connection", "close")
	res, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cep-rs doreq: %v", err)
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

func (c *rastro) Rastreia(objetos string, token string) (Resultado, error) {
	var result Resultado
	params := url.Values{}
	params.Add("resultado", "U")
	codigosObjetos := strings.Split(objetos, ",")
	for _, codigo := range codigosObjetos {
		params.Add("codigosObjetos", codigo)
	}
	res, err := c.doReq("GET", c.base+"?"+params.Encode(), token)
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
