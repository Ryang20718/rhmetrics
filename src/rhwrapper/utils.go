package rhwrapper

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"strings"
	"time"
)

func OneYearApart(dateStr1 string, dateStr2 string) bool {
	date1, _ := time.Parse("2006-01-02", dateStr1)
	date2, _ := time.Parse("2006-01-02", dateStr2)

	// Calculate the duration between the two dates
	duration := date2.Sub(date1)

	// Check if the duration is exactly 365 days
	if duration.Hours() == 365*24 {
		return true
	}
	return false
}

type Capture struct {
	UrlKey     string
	Timestamp  string
	Original   string
	MimeType   string
	StatusCode string
	Digest     string
	Length     string
}

func FetchStockSymbolChange(symbol string) (string, error) {
	// input: symbol potentially delisted and return current symbol
	url := fmt.Sprintf("http://web.archive.org/cdx/search/cdx?url=https://www.bloomberg.com/quote/%s:US&output=json&limit=10", symbol)
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("failing to get stock symbol change. url: %s ping @ryang", url)
	}
	defer resp.Body.Close()

	var result [][]string
	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return "", fmt.Errorf("failing to get stock symbol change. url: %s ping @ryang", url)
	}
	var capture Capture
	if len(result) == 0 {
		return "", fmt.Errorf("empty result")
	}
	for _, item := range result[len(result)-1:] { // Skip the first item because it's the field names
		capture = Capture{
			UrlKey:     item[0],
			Timestamp:  item[1],
			Original:   item[2],
			MimeType:   item[3],
			StatusCode: item[4],
			Digest:     item[5],
			Length:     item[6],
		}
	}
	webArchiveURL := fmt.Sprintf("http://web.archive.org/web/%s/%s", capture.Timestamp, capture.Original)
	resp, err = http.Get(webArchiveURL)
	if err != nil {
		return "", fmt.Errorf("failing to get stock symbol change. url: %s ping @ryang", webArchiveURL)
	}
	defer resp.Body.Close()
	document, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failing to get stock symbol change. url: %s ping @ryang", webArchiveURL)
	}
	htmlBody, err := document.Html()
	if err != nil {
		return "", fmt.Errorf("failing to get stock symbol change. url: %s ping @ryang", webArchiveURL)
	}

	currentSymbol := ""
	document.Find(".detailMessage__f82c6a6079").Each(func(index int, element *goquery.Selection) {
		if len(strings.Split(element.Text(), ":")) >= 3 {
			currentSymbol = strings.TrimSpace(strings.Split(element.Text(), ":")[2])
		}
	})
	document.Find(".TickerStatusMessage_tickerMessage__A8272").Each(func(index int, element *goquery.Selection) {
		if len(strings.Split(element.Text(), ":")) >= 3 {
			currentSymbol = strings.TrimSpace(strings.Split(element.Text(), ":")[2])
		}
	})
	if strings.Contains(htmlBody, "Ticker Change") && currentSymbol == "" {
		return "", fmt.Errorf("failing to get stock symbol change. url: %s ping @ryang", webArchiveURL)
	}
	return currentSymbol, nil
}
