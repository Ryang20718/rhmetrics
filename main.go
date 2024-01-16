package main

import (
	// "bufio"
	"context"
	// "fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	// "os"
	// "strings"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
	"rh_metrics/m/src/rhwrapper"
)

func isAuthenticated(c *gin.Context) {
	session := sessions.Default(c)
	// Check if "authenticated" is set to true in the session
	isAuthenticated := session.Get("authenticated") == true
	if !isAuthenticated {
		// If the user is not authenticated, redirect them to the login page
		c.Redirect(http.StatusSeeOther, "/metrics")
		c.Abort() // Prevent the handler from running
		return
	}
	c.Next() // If the user is authenticated, proceed to the handler
}

func main() {
	rhClient := rhwrapper.Hood{}
	router := gin.Default()
	store := cookie.NewStore([]byte("secret"))
	router.Use(sessions.Sessions("stateStorage", store))

	router.LoadHTMLGlob("templates/*.tmpl")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.tmpl", nil)
	})

	router.GET("/error", func(c *gin.Context) {
		// Get the last error message
		errMsg := c.Errors.ByType(gin.ErrorTypePrivate).Last().Error()
		c.HTML(http.StatusOK, "error.tmpl", gin.H{
			"ErrorMessage": errMsg,
		})
	})

	router.POST("/", func(c *gin.Context) {
		email := c.PostForm("email")
		password := c.PostForm("password")
		mfa := c.PostForm("mfa")
		cli, err := rhClient.Auth(email, password, mfa)
		if err != nil {
			c.Error(err) //nolint:errcheck
			c.Redirect(http.StatusSeeOther, "/error")
			return
		}
		session := sessions.Default(c)
		if err := session.Save(); err != nil {
			c.Error(err) //nolint:errcheck
			c.Redirect(http.StatusSeeOther, "/error")
			return
		}
		session.Set("authenticated", true)
		c.Redirect(http.StatusMovedPermanently, "/metrics")
		rhClient.Cli = cli
	})

	router.GET("/metrics", isAuthenticated, func(c *gin.Context) {
		session := sessions.Default(c)
		session.Set("authenticated", false)
		ctx := context.Background()
		profitDf, err := rhClient.ProcessRealizedEarnings(ctx)
		if err != nil {
			log.Fatalf("failing %v", err)
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
			GroupBy("Year", "Ticker").
			Aggregation([]dataframe.AggregationType{dataframe.Aggregation_SUM}, []string{"Amount"})

		earningsByTickerAmount := earningsDfByTicker.Col("Amount_SUM").Records()
		earningsByTickerLabels := earningsDfByTicker.Col("Ticker").Records()
		earningsByTickerYear := earningsDfByTicker.Col("Year").Records()

		labels := years
		data := ytdRealizedGains

		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"LabelsTimeSeries":    labels,
			"DataTimeSeries":      data,
			"LabelsYearsTag":      yearsTag,
			"LabelsTags":          tag,
			"DataAmount":          amount,
			"DataLabelsByTicker":  earningsByTickerLabels,
			"DataValByTicker":     earningsByTickerAmount,
			"DataValByTickerYear": earningsByTickerYear,
		})
	})

	router.Run(":8080") //nolint:errcheck

}
