package main

import (
	"net/http"
	"shawg/shawg"
)

func main() {
	r := shawg.New()
	// 使用一下全局中间件
	r.Use(shawg.Logger())
	r.GET("/hello", handleHello)
	r.GET("/user", handleUser)
	v1 := r.Group("/v1")
	{
		v1.GET("/", func(ctx *shawg.Context) {
			ctx.HTML(http.StatusOK, "<h1>Hello Everyone</h1>")
		})
		v1.GET("/hello", func(ctx *shawg.Context) {
			ctx.JSON(http.StatusOK, shawg.H{
				"message": "hello",
			})
		})
	}
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
