package socks5

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func init() {
	os.Setenv("GIN_MODE", "release")
	// os.Setenv("HTTP_PROXY", "socks5://127.0.0.1:1080")
	// os.Setenv("http_proxy", "socks5://127.0.0.1:1080")
	// os.Setenv("HTTPS_PROXY", "socks5://127.0.0.1:1080")
	// os.Setenv("https_proxy", "socks5://127.0.0.1:1080")
	// os.Setenv("ALL_PROXY", "socks5://127.0.0.1:1080")
	// os.Setenv("all_proxy", "socks5://127.0.0.1:1080")
}

var running atomic.Bool

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

func TestRunGin(t *testing.T) {
	RunGin()
}

func TestRawGin(t *testing.T) {
	go RunGin()
	time.Sleep(time.Millisecond * 100)

	resp, err := http.Get("http://172.31.57.106:8080/ping")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(string(data))
}

func TestPost(t *testing.T) {
	go RunGin()
	time.Sleep(time.Millisecond * 100)

	data, err := doPost("http://127.0.0.1:8080/send", 8)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%v", data)

}

func BenchmarkProxy(b *testing.B) {

	for i := 0; i < b.N; i++ {
		err := getPing("http://172.31.57.106:8080/ping")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkDierct(b *testing.B) {

	for i := 0; i < b.N; i++ {
		err := getPing("http://127.0.0.1:8080/ping")
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPostProxy(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := doPost("http://172.31.57.106:8080/send", 4096)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkPostDirect(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := doPost("http://127.0.0.1:8080/send", 4096)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func getPing(url string) (err error) {
	resp, err := http.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	_, err = io.ReadAll(resp.Body)
	if err != nil {
		return
	}
	return
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
