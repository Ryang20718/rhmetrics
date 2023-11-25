package main

import (
	"bufio"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"

	"rh_metrics/m/src/rhwrapper"
)

func main() {

	if len("GG") > 10 {
		router := gin.Default()
		router.LoadHTMLGlob("templates/*.tmpl")
		router.GET("/index", func(c *gin.Context) {
			// Replace the following with your actual time series data
			labels := []string{"2023-01-01", "2023-01-02", "2023-01-03"}
			data := []float64{10, 20, 15}

			c.HTML(http.StatusOK, "index.tmpl", gin.H{
				"Labels": labels,
				"Data":   data,
			})
		})
		router.Run(":8080")
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Please Enter Your MFA: ")
	mfa, _ := reader.ReadString('\n')

	rhClient := rhwrapper.Hood{}

	username := os.Getenv("ROBINHOOD_USERNAME")
	password := os.Getenv("ROBINHOOD_PASSWORD")
	rhClient.Auth(username, password, mfa)
	fmt.Println(rhClient)

}
