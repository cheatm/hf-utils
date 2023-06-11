package main

import (
	"fmt"
	"io"
	"os"
	"sync/atomic"

	"github.com/armon/go-socks5"
	"github.com/gin-gonic/gin"
)

func main() {
	// Create a SOCKS5 server

	if len(os.Args) == 2 {
		switch os.Args[1] {
		case "gin":
			RunGin()
		case "proxy":
			RunProxy()
		}

		return
	}
	go RunGin()
	RunProxy()
}

var running atomic.Bool

func RunProxy() {
	conf := &socks5.Config{}
	server, err := socks5.New(conf)
	if err != nil {
		panic(err)
	}

	// Create SOCKS5 proxy on localhost port 8000
	if err := server.ListenAndServe("tcp", "127.0.0.1:1080"); err != nil {
		panic(err)
	}
}

func RunGin() {
	if !running.CompareAndSwap(false, true) {
		return
	}
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})
	r.POST("/send", func(ctx *gin.Context) {
		data, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			fmt.Printf("err: %s\n", err)
			return
		}

		ctx.Writer.Write(data)

	})
	r.Run("0.0.0.0:8080")

}
