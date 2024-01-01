package rhwrapper

import (
	"time"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/PuerkitoBio/goquery"
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

var proxies = make(map[int]string)

func FetchProxies() (string, error) {
	if len(proxies) > 0 {
		return proxies[rand.Intn(len(proxies))], nil
	}
	// input: symbol potentially delisted and return current symbol
	req, err := http.NewRequest("GET", fmt.Sprintf("https://www.socks-proxy.net/"), nil)
	if err != nil {
		return "", err
	}
	// Add headers
	req.Header.Set("User-Agent", "Mozilla/80.0")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	document, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal("Error loading HTTP response body. ", err)
	}
	proxyIndex := 0
	document.Find("tr").Each(func(i int, s *goquery.Selection) {
		index := 0
		ip := ""
		port := ""
		isUSA := false
		s.Find("td").Each(func(j int, t *goquery.Selection) {
			index += 1
			if index == 1 {
				if strings.Count(t.Text(), ".") == 3 {
					ip = t.Text()
				}
			}
			if index == 2 {
				port = t.Text()
			}
			if index == 3 {
				isUSA = (t.Text() == "US")
			}
		})
		if isUSA && ip != "" {
			proxyUrl := fmt.Sprintf("http://%s:%s", ip, port)
			proxies[proxyIndex] = proxyUrl
			proxyIndex += 1
		}
	})
	return proxies[rand.Intn(len(proxies))], nil
}

func FetchStockSymbolChange(symbol string) (string, error) {
	// input: symbol potentially delisted and return current symbol
	getProxy, err := FetchProxies()
	if err != nil {
		return "", fmt.Errorf("failing to fetch proxies %v", err)
	}
	proxyURL, err := url.Parse(getProxy) // TODO make this more robust
	if err != nil {
		return "", err
	}

	transport := &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
	}

	client := &http.Client{
		Transport: transport,
	}

	req, err := client.Get(fmt.Sprintf("https://www.bloomberg.com/quote/%s:US", symbol))
	if err != nil {
		return "", err
	}
	// Add headers
	req.Header.Set("User-Agent", "Mozilla/80.0")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	document, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Fatal("Error loading HTTP response body. ", err)
	}

	// TODO convert this to some sort of regex
	currentSymbol := ""
	document.Find(".TickerStatusMessage_tickerMessage__A8272").Each(func(index int, element *goquery.Selection) {
		currentSymbol = strings.Split(element.Text(), ":")[2]
	})
	return currentSymbol, nil
}