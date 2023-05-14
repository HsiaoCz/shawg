package main

import (
	"net/http"
	"shawg/shawg"
)

func main() {
	r := shawg.New()
	r.GET("/hello", handleHello)
	r.GET("/user", handleUser)
	r.Run(":9091")
}

func handleHello(c *shawg.Context) {
	c.String(http.StatusOK, "Hello %s", "bob")
}

func handleUser(c *shawg.Context) {
	c.JSON(http.StatusOK, shawg.H{
		"User":    "alex",
		"Message": "hello",
	})
}
