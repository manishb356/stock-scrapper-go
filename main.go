package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"

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

	c := colly.NewCollector()

	c.OnRequest(func(r *colly.Request) {
		r.Headers.Set("User-Agent", "Mozilla 5.0")
		fmt.Println("Visiting:", r.URL)
	})

	c.OnError(func(_ *colly.Response, err error) {
		log.Println("Something went wrong ", err)
	})

	c.OnHTML("aside.stocks-sidebar-container", func(e *colly.HTMLElement) {
		stock := stock{}
		stock.name = e.ChildText("h1.security-name")
		stock.ticker = e.ChildText("div.sidebar-security-ticker span.ticker")
		stock.price = e.ChildText("span.current-price")
		stock.change = strings.Trim(e.ChildText("span.change.percentage-value"), "()")

		stocks = append(stocks, stock)
	})

	c.Wait()

	for _, t := range tickers {
		c.Visit("https://www.tickertape.in/stocks/" + t + "/")
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
