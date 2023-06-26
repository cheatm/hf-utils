package main

import (
	"bytes"
	"context"
	"fmt"
	socks5t "hf-utils/socks5"
	"io"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/armon/go-socks5"
	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
)

func main() {
	// Create a SOCKS5 server

	if len(os.Args) == 2 {
		switch os.Args[1] {
		case "gin":
			RunGin()
		case "proxy":
			RunProxy()
		case "redis":
			RunRedis()
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
	if err := server.ListenAndServe("tcp", "0.0.0.0:1080"); err != nil {
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
		host := ctx.ClientIP()
		data, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			fmt.Printf("err: %s\n", err)
			return
		}
		hostBytes := []byte(host)
		headSize := len(hostBytes)

		response := make([]byte, 1+headSize+len(data))
		// data = append([]byte(host), data...)
		response[0] = uint8(headSize)
		for i := 0; i < headSize; i++ {
			response[1+i] = hostBytes[i]
		}
		for i := 0; i < len(data); i++ {
			response[1+headSize+i] = data[i]
		}
		// response[1:1+headSize] = hostBytes[:]

		ctx.Writer.Write(response)

	})
	r.Run("0.0.0.0:8080")

}

func UnsetProxy() {
	os.Unsetenv("HTTP_PROXY")
	os.Unsetenv("http_proxy")
	os.Unsetenv("HTTPS_PROXY")
	os.Unsetenv("https_proxy")
	os.Unsetenv("ALL_PROXY")
	os.Unsetenv("all_proxy")
}

func RunRedis() {
	UnsetProxy()
	pool, err := socks5t.NewRedisPool("redis://172.16.20.81:6379", 20, 20)
	if err != nil {
		panic(err)
	}
	streamer := socks5t.NewRedisStreamer(pool, socks5t.DefaultSendChan)
	streamer.SetHandler(func(m *redis.Message) error {
		data, err := doPost("http://172.16.20.91:8080/send", 4096)
		if err != nil {
			data = []byte(err.Error())
		}
		// data := m.Data

		num, err := streamer.Pub(socks5t.DefaultRecvChan, data)
		if err != nil {
			return err
		}
		if num == 0 {
			return fmt.Errorf("no reciever")
		}
		return nil
	})
	streamer.Run(context.TODO())

}

func doPost(url string, payloadSize int) (result []byte, err error) {
	data := make([]byte, payloadSize)
	data[0] = 1
	data[payloadSize-1] = 1
	resp, err := http.Post(url, "plain/text", bytes.NewBuffer(data))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	result, err = io.ReadAll(resp.Body)
	return
}
