package main

import (
		"github.com/gin-gonic/gin"
		"net/http"
)

func main() {
	router := gin.Default()
	router.Static("styles", "./styles")
	router.LoadHTMLGlob("templates/*")

	//	v1 route
	v1 := router.Group("/v1")
	{
		v1.GET("/top", topPageEndPoint)
		v1.GET("/create_account_page", createAccountPageEndPoint)
	}
	router.Run(":8080")
}

func topPageEndPoint(c *gin.Context) {
	c.HTML(http.StatusOK, "top.tmpl", gin.H{
		"title": "Top Page",
	})
}

func createAccountPageEndPoint(c *gin.Context) {
	c.HTML(http.StatusOK, "createAccount.tmpl", gin.H{})
}