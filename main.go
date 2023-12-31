package main

import (
	"context"
	"fmt"
	"bufio"
	"os"
	"strings"
	"github.com/gin-gonic/gin"
	"net/http"

	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
	"rh_metrics/m/src/rhwrapper"
)

func main() {
	rhClient := rhwrapper.Hood{}
	if len("GG") == 10 {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Please Enter Your MFA: ")
		mfa, _ := reader.ReadString('\n')
		mfa = strings.TrimSuffix(mfa, "\n")
		username := os.Getenv("ROBINHOOD_USERNAME")
		password := os.Getenv("ROBINHOOD_PASSWORD")
		cli, err := rhClient.Auth(username, password, mfa)
		if err != nil {
			return
		}
		rhClient.Cli = cli
	}


	ctx := context.Background()
	profitDf, err := rhClient.ProcessRealizedEarnings(ctx)
	if err != nil {
		return
	}

	// see rhwrapper.go
	aggregatedDf := profitDf.
		GroupBy("Year").
		Aggregation([]dataframe.AggregationType{dataframe.Aggregation_SUM}, []string{"Amount"})

	aggregatedDf = aggregatedDf.Arrange(
		dataframe.Sort("Year"),
	)
	years := aggregatedDf.Col("Year").Records()
	ytdRealizedGains := aggregatedDf.Col("Amount_SUM").Records()

	tagDf := profitDf.
		GroupBy("Year", "Tag").
		Aggregation([]dataframe.AggregationType{dataframe.Aggregation_SUM}, []string{"Amount"})

	tagDf = tagDf.Arrange(
		dataframe.Sort("Year"),
	)
	amount := tagDf.Col("Amount_SUM").Records()
	yearsTag := tagDf.Col("Year").Records()
	tag := tagDf.Col("Tag").Records()

	earningsDfByTicker := profitDf.
		Filter(dataframe.F{
			Colname:    "Ticker",
			Comparator: series.Neq,
			Comparando: "",
		}).
		GroupBy("Ticker").
		Aggregation([]dataframe.AggregationType{dataframe.Aggregation_SUM}, []string{"Amount"})

	earningsByTickerAmount := earningsDfByTicker.Col("Amount_SUM").Records()
	earningsByTickerLabels := earningsDfByTicker.Col("Ticker").Records()

	fmt.Println(aggregatedDf)
	if len("GG") > 10 {
		router := gin.Default()
		router.LoadHTMLGlob("templates/*.tmpl")
		router.GET("/", func(c *gin.Context) {
			// Replace the following with your actual time series data
			labels := years
			data := ytdRealizedGains

			c.HTML(http.StatusOK, "index.tmpl", gin.H{
				"LabelsTimeSeries":   labels,
				"DataTimeSeries":     data,
				"LabelsYearsTag":     yearsTag,
				"LabelsTags":         tag,
				"DataAmount":         amount,
				"DataLabelsByTicker": earningsByTickerLabels,
				"DataValByTicker":    earningsByTickerAmount,
			})
		})
		router.Run(":8080")
	}

}
