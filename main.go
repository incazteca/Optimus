package main

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/shopspring/decimal"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

//Holding describes a holding that a user has
type Holding struct {
	Symbol           string
	Quantity         decimal.Decimal
	TargetAllocation decimal.Decimal
	Price            decimal.Decimal
	PriceFetchedAt   time.Time
}

// QuoteResponse from Global Quote endpoint from alphavantage
type QuoteResponse struct {
	GlobalQuote QuoteData `json:"Global Quote"`
}

// QuoteData is data from QuoteReesponse
type QuoteData struct {
	Symbol           string `json:"01. symbol"`
	Open             string `json:"02. open"`
	High             string `json:"03. high"`
	Low              string `json:"04. low"`
	Price            string `json:"05. price"`
	Volume           string `json:"06. volume"`
	LatestTradingDay string `json:"07. latest trading day"`
	PreviousClose    string `json:"08. previous close"`
	Change           string `json:"09. change"`
	ChangePercent    string `json:"10. change percent"`
}

const dataFile = "data.csv"
const avAPIURL = "https://www.alphavantage.co/query?function=%s&symbol=%s&apikey=%s"

func main() {
	holdings, err := fetchHoldings()

	if err != nil {
		log.Fatal(err)
	}

	err = fetchSymbolData(holdings)

	if err != nil {
		log.Fatal(err)
	}

	for _, holding := range holdings {
		fmt.Println(holding)
	}
}

// fetchHoldings gets the current holdings in data.csv, In the future we'll just
// fetch from Vanguard
func fetchHoldings() ([]*Holding, error) {
	csvFile, err := os.Open(dataFile)

	if err != nil {
		return nil, err
	}

	reader := csv.NewReader(bufio.NewReader(csvFile))
	var holdings []*Holding

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		// Skip header
		if record[0] == "symbol" {
			continue
		}

		holding := Holding{
			Symbol:           record[0],
			Quantity:         decimal.RequireFromString(record[1]),
			TargetAllocation: decimal.RequireFromString(record[2]),
		}

		holdings = append(holdings, &holding)
	}

	return holdings, nil
}

//fetchSymbolData gets extra data from AlphaVantage
func fetchSymbolData(holdings []*Holding) error {
	avKey := os.Getenv("AV_API_KEY")
	avFunction := "GLOBAL_QUOTE"

	for _, holding := range holdings {
		url := fmt.Sprintf(avAPIURL, avFunction, holding.Symbol, avKey)
		response, err := http.Get(url)

		// If we get an error just return
		if err != nil {
			return err
		}

		defer response.Body.Close()
		body, err := ioutil.ReadAll(response.Body)

		if err != nil {
			return err
		}

		var quote QuoteResponse
		json.Unmarshal(body, &quote)

		holding.Price = decimal.RequireFromString(quote.GlobalQuote.Price)
		holding.PriceFetchedAt = time.Now()
	}

	return nil
}
