package rhwrapper

import (
	"encoding/gob"
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

func BeforeDate(originalDate string, dateToCompare string) bool {
	date1, _ := time.Parse("2006-01-02", originalDate)
	date2, _ := time.Parse("2006-01-02", dateToCompare)

	// Calculate the duration between the two dates
	duration := date2.Sub(date1)

	// Check if the dateToCompare is before the originalDate
	return duration.Hours() < 0
}

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
	// TODO make this less hacky
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
	if currentSymbol == "" {
		return symbol, nil
	}
	return currentSymbol, nil
}

type Event struct {
	Date        float64 `json:"date"`
	Denominator float64 `json:"denominator"`
	Numerator   float64 `json:"numerator"`
	SplitRatio  string  `json:"splitRatio"`
}

type Result struct {
	Events     map[string]map[string]Event       `json:"events"`
	Indicators map[string][]map[string][]float64 `json:"indicators"`
}

type Chart struct {
	Error  interface{} `json:"error"`
	Result []Result    `json:"result"`
}

type Data struct {
	Chart Chart `json:"chart"`
}

type Split struct {
	Date        string
	Numerator   int
	Denominator int
}

var CacheStockSplits = make(map[string][]Split)

func FetchStockSplits(symbol string) ([]Split, error) {
	url := fmt.Sprintf("%s/v8/finance/chart/%s?period1=0&period2=9999999999&interval=3mo&events=split", "https://query2.finance.yahoo.com", symbol)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if string(body) == "Will be right back" {
		return nil, fmt.Errorf("*** YAHOO! FINANCE IS CURRENTLY DOWN! ***\nOur engineers are working quickly to resolve the issue. Thank you for your patience.")
	}

	var data Data
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, err
	}

	splits := []Split{}
	for _, result := range data.Chart.Result {
		for _, v := range result.Events {
			for _, val := range v {
				timestamp := int64(val.Date)

				// Convert the Unix timestamp to a time.Time value
				t := time.Unix(timestamp, 0)
				date := t.Format("2006-01-02")
				split := Split{
					Date:        date,
					Numerator:   int(val.Numerator),
					Denominator: int(val.Denominator),
				}
				splits = append(splits, split)
			}
		}
	}
	return splits, nil
}

func GetStockSplitCorrection(symbol string, date string, qty float64, price float64) (float64, float64, error) {
	// returns stock split corrected amount based on present day
	if _, keyFound := CacheStockSplits[symbol]; !keyFound {
		stockSplits, err := FetchStockSplits(symbol)
		if err != nil {
			return 0.0, 0.0, err
		}
		CacheStockSplits[symbol] = stockSplits
	}
	qtyMultiplier := 1.0
	priceMultiplier := 1.0
	for _, split := range CacheStockSplits[symbol] {
		if BeforeDate(split.Date, date) {
			if split.Numerator != 1 {
				// Numerator, so we need to divide cost and multiply count
				qtyMultiplier *= float64(split.Numerator)
				priceMultiplier /= float64(split.Numerator)
			} else {
				// Denominator, so we need to multiply cost and divide count
				// reverse split
				qtyMultiplier /= float64(split.Denominator)
				priceMultiplier *= float64(split.Denominator)
			}
		}
	}
	return (qty * qtyMultiplier), (price * priceMultiplier), nil
}

func CacheAPICall(cacheFilePath string, dataToEncode interface{}) error {
	encodeFile, err := os.Create(cacheFilePath)
	if err != nil {
		return fmt.Errorf("failing to create encoding. ERR: %v", err)
	}

	// Since this is a binary format large parts of it will be unreadable
	encoder := gob.NewEncoder(encodeFile)

	if err := encoder.Encode(dataToEncode); err != nil {
		return fmt.Errorf("failing to encode. ERR: %v", err)
	}
	encodeFile.Close()
	return nil
}
