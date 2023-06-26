package socks5t

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sync/atomic"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gomodule/redigo/redis"
)

const (
	// ProxyHost = "socks5://172.16.20.91:1080"
	ProxyHost = "socks5://127.0.0.1:1080"
	GinSever  = "http://172.16.20.91:8080"
)

func init() {
	os.Setenv("GIN_MODE", "release")
}

func SetProxy() {
	proxy := ProxyHost
	os.Setenv("HTTP_PROXY", proxy)
	os.Setenv("http_proxy", proxy)
	os.Setenv("HTTPS_PROXY", proxy)
	os.Setenv("https_proxy", proxy)
	os.Setenv("ALL_PROXY", proxy)
	os.Setenv("all_proxy", proxy)
}

func UnsetProxy() {
	os.Unsetenv("HTTP_PROXY")
	os.Unsetenv("http_proxy")
	os.Unsetenv("HTTPS_PROXY")
	os.Unsetenv("https_proxy")
	os.Unsetenv("ALL_PROXY")
	os.Unsetenv("all_proxy")
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

	resp, err := http.Get(GinSever + "/ping")
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
	// go RunGin()
	UnsetProxy()
	// SetProxy()

	// transport := new(http.Transport)
	// *transport = *http.DefaultTransport.(*http.Transport)
	// _u, _ := url.Parse(ProxyHost)
	// transport.Proxy = http.ProxyURL(_u)
	// client := &http.Client{Transport: transport}
	client := &http.Client{}

	data, err := clientPost(client, "http://172.16.20.91:8080/send", 8)
	if err != nil {
		t.Fatal(err)
	}
	// t.Logf("%v", data)
	source, reply := DecodeSendResponse(data)
	t.Logf("%s | %v", source, reply)
}

func TestPostRedis(t *testing.T) {
	UnsetProxy()
	pool, err := NewRedisPool("redis://172.16.20.81:6379", 20, 20)
	if err != nil {
		t.Fatal(err)
	}

	streamer := NewRedisStreamer(pool, DefaultRecvChan)
	ch := make(chan []byte, 10)
	streamer.SetHandler(func(m *redis.Message) error {
		// t.Logf("Recieve: %s", string(m.Data))
		ch <- m.Data
		return nil
	})
	go streamer.Run(context.TODO())
	for !streamer.Listening() {
	}
	streamer.Pub(DefaultSendChan, "send")

	resp := <-ch
	t.Logf("resp: %s", string(resp))
}

var Streaming = atomic.Bool{}

func BenchmarkPostRedis(b *testing.B) {
	pool, err := NewRedisPool("redis://172.16.20.81:6379", 20, 20)
	if err != nil {
		b.Fatal(err)
	}

	streamer := NewRedisStreamer(pool, DefaultRecvChan)
	ch := make(chan []byte, 1000)
	streamer.SetHandler(func(m *redis.Message) error {
		ch <- m.Data
		return nil
	})
	ctx, cancel := context.WithCancel(context.TODO())
	streamer.cancel = cancel
	go streamer.Run(ctx)

	for !streamer.Listening() {
	}
	b.Logf("Start")
	count := atomic.Int64{}
	data := make([]byte, 128)
	data[0] = 1
	for i := 0; i < b.N; i++ {
		streamer.Pub(DefaultSendChan, data)
		<-ch
		count.Add(1)
	}
	b.Logf("Count = %d", count.Load())
	streamer.Stop()
}

func DecodeSendResponse(data []byte) (source string, reply []byte) {
	headSize := data[0]
	source = string(data[1 : 1+headSize])
	reply = data[1+headSize:]
	return
}

func BenchmarkProxy(b *testing.B) {

	for i := 0; i < b.N; i++ {
		err := getPing("http://127.0.0.1:8080/ping")
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
	transport := new(http.Transport)
	*transport = *http.DefaultTransport.(*http.Transport)
	_u, _ := url.Parse(ProxyHost)
	transport.Proxy = http.ProxyURL(_u)
	client := &http.Client{Transport: transport}

	host := GinSever + "/send"
	addr := ""
	for i := 0; i < b.N; i++ {
		data, err := clientPost(client, host, 4096)
		if err != nil {
			b.Fatal(err)
		}
		source, _ := DecodeSendResponse(data)
		if len(addr) == 0 {
			addr = source
		} else if addr != source {
			b.Fatalf("source unmatch: %s <-> %s", addr, source)
		}

	}
	b.Logf("Source: %s", addr)
}

func BenchmarkPostDirect(b *testing.B) {
	UnsetProxy()
	client := &http.Client{}
	host := GinSever + "/send"
	addr := ""
	for i := 0; i < b.N; i++ {
		data, err := clientPost(client, host, 4096)
		if err != nil {
			b.Fatal(err)
		}
		source, _ := DecodeSendResponse(data)
		if len(addr) == 0 {
			addr = source
		} else if addr != source {
			b.Fatalf("source unmatch: %s <-> %s", addr, source)
		}

	}
	b.Logf("Source: %s", addr)
}

func BenchmarkBoth(b *testing.B) {
	b.Run("RPOXY ", BenchmarkPostProxy)
	b.Run("DIRECT", BenchmarkPostDirect)
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

func clientPost(client *http.Client, url string, payloadSize int) (result []byte, err error) {
	data := make([]byte, payloadSize)
	data[0] = 1
	data[payloadSize-1] = 1
	resp, err := client.Post(url, "plain/text", bytes.NewBuffer(data))
	if err != nil {
		return
	}
	defer resp.Body.Close()

	result, err = io.ReadAll(resp.Body)
	return
}
