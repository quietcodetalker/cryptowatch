package token

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"nhooyr.io/websocket"
	"nhooyr.io/websocket/wsjson"
	"strings"
)

const (
	apiURL = "https://min-api.cryptocompare.com/data/pricemulti"
)

type cryptoCompareProvider struct {
	apiKey     string
	httpClient *http.Client
	conn       *websocket.Conn
}

func NewCryptoCompareProvider(apiKey string, httpClient *http.Client) *cryptoCompareProvider {
	return &cryptoCompareProvider{
		apiKey:     apiKey,
		httpClient: httpClient,
	}
}

func (c *cryptoCompareProvider) Start(ctx context.Context, ch chan<- map[string]float64) error {
	var err error
	c.conn, _, err = websocket.Dial(ctx, "wss://streamer.cryptocompare.com/v2?api_key="+c.apiKey, nil)
	if err != nil {
		return ErrInternalError
	}

	go func() {
		for {
			var v answer
			err = wsjson.Read(ctx, c.conn, &v)
			if err != nil {
				log.Printf("error: %v", err)
				continue
			}

			if v.Type == "5" && v.Price != nil {
				ch <- map[string]float64{
					v.FromSymbol: *v.Price,
				}
			}
		}
	}()

	return nil
}

func (c *cryptoCompareProvider) GetPrices(ctx context.Context, tokens []string) (map[string]float64, error) {

	req, err := http.NewRequest(http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, ErrInternalError
	}

	values := req.URL.Query()
	values.Add("fsyms", strings.Join(tokens, ","))
	values.Add("tsyms", "USD,BTC")
	req.URL.RawQuery = values.Encode()

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, ErrInternalError
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, ErrInternalError
	}
	defer res.Body.Close()
	var rawRes map[string]map[string]float64
	json.Unmarshal(b, &rawRes)

	result := make(map[string]float64, len(rawRes))
	for k, v := range rawRes {
		result[k] = v["USD"]
	}

	return result, nil
}

type message struct {
	Action string   `json:"action"`
	Subs   []string `json:"subs"`
}

type answer struct {
	Type       string   `json:"TYPE"`
	FromSymbol string   `json:"FROMSYMBOL"`
	Price      *float64 `json:"PRICE"`
}

func (p *cryptoCompareProvider) Subscribe(ctx context.Context, tickers []string) error {
	msg := message{
		Action: "SubAdd",
		Subs:   nil,
	}
	for _, t := range tickers {
		msg.Subs = append(msg.Subs, fmt.Sprintf("5~CCCAGG~%s~USD", t))
	}
	msg2 := map[string]interface{}{
		"action": "SubAdd",
		"subs":   msg.Subs,
	}

	err := wsjson.Write(ctx, p.conn, msg2)
	if err != nil {
		return ErrInternalError
	}

	return nil
}
