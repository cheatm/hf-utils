package proxy

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"testing"
	"time"
)

func TestLocalIP(t *testing.T) {
	interfaces, err := net.Interfaces()
	if err != nil {
		t.Fatal(err)
	}
	for _, it := range interfaces {
		// t.Logf("interface: %#v", it)
		addrs, err := it.Addrs()
		if err != nil {
			t.Fatal(err)
		}
		t.Logf("%s: %s", it.Name, addrs)
		for _, addr := range addrs {
			t.Logf("%s: %s", addr.Network(), addr.String())
		}
	}
}

func getAddr(name string, index int) (net.Addr, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, it := range interfaces {
		if it.Name != name {
			continue
		}
		addrs, err := it.Addrs()
		if err != nil {
			return nil, err
		}
		if index < len(addrs) {
			return addrs[index], nil
		}
		return nil, fmt.Errorf("index out of bounds: %d", len(addrs))
	}
	return nil, fmt.Errorf("name not found: %s", name)

}

func TestResolve(t *testing.T) {
	addr, err := net.ResolveUnixAddr("unix", "tcp")
	if err != nil {
		panic(err)
	}
	t.Logf("%s: %s", addr.Network(), addr.String())
	hint, err := getAddr("eth0", 0)
	if err != nil {
		panic(err)
	}
	t.Logf("%s: %s", hint.Network(), hint.String())
	ipa, err := net.ResolveIPAddr("ip", "172.20.104.196")
	if err != nil {
		panic(err)
	}
	t.Logf("%s: %s", ipa.Network(), ipa.IP)

}

func TestType(t *testing.T) {
	addr, err := getAddr("eth0", 1)
	if err != nil {
		panic(err)
	}
	ipa := addr.(*net.IPNet)
	t.Logf("IP: %s", ipa.IP)
}

const HOST string = "testnet.binancefuture.com"

func TestClient(t *testing.T) {
	name := os.Getenv("NET_NAME")
	if len(name) == 0 {
		name = "eth0"
	}
	addr, err := getAddr(name, 0)
	// addr, err := net.ResolveIPAddr("ip", "172.20.104.196")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%s: %s", addr.Network(), addr.String())

	dialer := net.Dialer{
		LocalAddr: &net.TCPAddr{IP: addr.(*net.IPNet).IP},
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		Dial:                  dialer.Dial,
		DialContext:           dialer.DialContext,
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	client := http.Client{Transport: transport}

	resp, err := client.Get("https://" + HOST + "/fapi/v1/ping")
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("ping: %s", string(data))
}
