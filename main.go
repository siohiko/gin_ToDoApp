package main

import (
		"github.com/gin-gonic/gin"
		"net/http"
)

func main() {
	router := gin.Default()

	router.Static("styles", "./styles")
	router.LoadHTMLGlob("templates/*")

	router.GET("/index", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{
			"title": "Top Page",
		})
	})
	router.Run(":8080")
}
