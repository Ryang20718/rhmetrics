package rhwrapper

import (
	"encoding/json"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"net/http"
	"strings"
	"time"
)

func BeforeDate(originalDate string, dateToCompare string) bool {
	date1, _ := time.Parse("2006-01-02", originalDate)
	date2, _ := time.Parse("2006-01-02", dateToCompare)

	// Calculate the duration between the two dates
	duration := date2.Sub(date1)

	// Check if the dateToCompare is before the originalDate
	if duration.Hours() <= 0 {
		return true
	}
	return false
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

func FetchStockSplits(symbol string) ([]Split, error) {
	url := fmt.Sprintf("%s/v8/finance/chart/%s?period1=0&period2=9999999999&interval=3mo&events=split", "https://query2.finance.yahoo.com", symbol)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
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

				fmt.Println("TIME", date)
				fmt.Println(val)
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