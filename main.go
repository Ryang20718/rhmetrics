package main

import (
	// "bufio"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	// "os"
	// "strings"

	"github.com/go-gota/gota/dataframe"
	"github.com/go-gota/gota/series"
	"rh_metrics/m/src/rhwrapper"
)

func redirectError(c *gin.Context, err error) {
	c.Error(err)
	c.Redirect(http.StatusSeeOther, "/error")
}

func isAuthenticated(c *gin.Context) {
	session := sessions.Default(c)
	// Check if "authenticated" is set to true in the session
	isAuthenticated := session.Get("authenticated") == true

	if !isAuthenticated {
		// If the user is not authenticated, redirect them to the login page
		c.Redirect(http.StatusSeeOther, "/login")
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
	
	router.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.tmpl", nil)
	})
	
	router.GET("/error", func(c *gin.Context) {
		// Get the last error message
		errMsg := c.Errors.ByType(gin.ErrorTypePrivate).Last().Error()
		c.HTML(http.StatusOK, "error.tmpl", gin.H{
		  "ErrorMessage": errMsg,
		})
	})

	router.POST("/login", func(c *gin.Context) {
		email := c.PostForm("email")
		password := c.PostForm("password")
		mfa := c.PostForm("mfa")
		cli, err := rhClient.Auth(email, password, mfa)
		if err != nil {
			c.Error(err)
			c.Redirect(http.StatusSeeOther, "/error")
			return
		}
		session := sessions.Default(c)
		session.Set("authenticated", true)
		if err := session.Save(); err != nil {
			c.Error(err)
			c.Redirect(http.StatusSeeOther, "/error")
			return
		}
	
		c.Redirect(http.StatusMovedPermanently, "/")
		rhClient.Cli = cli
	})

	// if len("GG") == 3 {
	// 	reader := bufio.NewReader(os.Stdin)
	// 	fmt.Print("Please Enter Your MFA: ")
	// 	mfa, _ := reader.ReadString('\n')
	// 	mfa = strings.TrimSuffix(mfa, "\n")
	// 	username := os.Getenv("ROBINHOOD_USERNAME")
	// 	password := os.Getenv("ROBINHOOD_PASSWORD")
	// 	cli, err := rhClient.Auth(username, password, mfa)
	// 	if err != nil {
	// 		log.Fatalf("failing to authenticate rhood %v", err)
	// 	}
	// 	rhClient.Cli = cli
	// }
	
	router.GET("/", isAuthenticated, func(c *gin.Context) {
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
		fmt.Println(tagDf)
		
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
			"LabelsTimeSeries":   labels,
			"DataTimeSeries":     data,
			"LabelsYearsTag":     yearsTag,
			"LabelsTags":         tag,
			"DataAmount":         amount,
			"DataLabelsByTicker": earningsByTickerLabels,
			"DataValByTicker":    earningsByTickerAmount,
			"DataValByTickerYear": earningsByTickerYear,
		})
	})
	
	router.Run(":8080")

}
