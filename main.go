package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"

	"github.com/gocolly/colly"
)

type stock struct {
	name, ticker, price, change string
}

func main() {

	tickers := []string{
		"hdfc-bank-HDBK",
		"icici-bank-ICBK",
		"reliance-industries-RELI",
		"infosys-INFY",
		"itc-ITC",
		"tata-consultancy-services-TCS",
		"larsen-and-toubro-LART",
	}

	stocks := []stock{}
	stocksChan := make(chan stock, len(tickers))

	var wg sync.WaitGroup

	c := colly.NewCollector(
		colly.Async(true),
	)

	c.Limit(&colly.LimitRule{
		DomainGlob:  "*",
		Parallelism: len(tickers),
		Delay:       1,
	})

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("User-Agent", "Mozilla 5.0")
		fmt.Println("Visiting:", r.URL)
	})

	c.OnError(func(_ *colly.Response, err error) {
		log.Println("Something went wrong ", err)
		wg.Done()
	})

	c.OnHTML("aside.stocks-sidebar-container", func(e *colly.HTMLElement) {
		stock := stock{}
		stock.name = e.ChildText("h1.security-name")
		stock.ticker = e.ChildText("div.sidebar-security-ticker span.ticker")
		stock.price = e.ChildText("span.current-price")
		stock.change = strings.Trim(e.ChildText("span.change.percentage-value"), "()")

		stocksChan <- stock
		wg.Done()
	})

	c.Wait()

	for _, t := range tickers {
		wg.Add(1)

		go func(ticker string) {
			err := c.Visit("https://www.tickertape.in/stocks/" + ticker + "/")
			if err != nil {
				log.Printf("Error visiting %s: %v\n", ticker, err)
				wg.Done()
			}
		}(t)
	}

	go func() {
		wg.Wait()
		close(stocksChan)
	}()

	for stock := range stocksChan {
		stocks = append(stocks, stock)
	}

	file, err := os.Create("stocks.csv")
	if err != nil {
		log.Fatal("Failed to create csv file")
	}
	defer file.Close()
	writer := csv.NewWriter(file)

	headers := []string{
		"company name",
		"company ticker",
		"price",
		"price change",
	}

	writer.Write(headers)

	for _, stock := range stocks {
		record := []string{
			stock.name,
			stock.ticker,
			stock.price,
			stock.change,
		}

		writer.Write(record)
	}
	defer writer.Flush()
}
