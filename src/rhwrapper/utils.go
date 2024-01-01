package rhwrapper

import (
	"time"
	"fmt"
	"log"
	"net/http"
	"io"
	"net/url"
	"os"
	"bufio"
	"strings"
	"math/rand"
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
var agent = make(map[int]string)

func FetchProxies() (string, error) {
	// This is needed for crawling :)
	// Fetches random proxy from proxy
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

func FetchAgent() (string, error) {
	// This is needed for crawling :)
	// fetches random user agent to configure bot
	if len(agent) > 0 {
		return agent[rand.Intn(len(agent))], nil
	}
	file, err := os.Open("src/rhwrapper/agents.txt")
	if err != nil {
		return "", err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	idx := 0
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err != io.EOF {
				return "", err
			}
			break
		}
		agent[idx] = line
		idx += 1
	}
	return agent[rand.Intn(len(agent))], nil
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

	req, err := http.NewRequest("GET", fmt.Sprintf("https://www.bloomberg.com/quote/%s:US", symbol), nil)
	if err != nil {
		return "", err
	}
	// Add headers
	agent, err := FetchAgent()
	req.Header.Set("User-Agent", agent)
	fmt.Println(agent)

	// TODO REMOVE THIS dumb wait
	time.Sleep(time.Duration(rand.Intn(3000)) * time.Millisecond)
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