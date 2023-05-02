package proxy

import (
	"net"
	"testing"
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
