package zmq

import (
	"context"
	"os"
	"testing"
	"time"

	zmq "github.com/pebbe/zmq4"
)

func TestHasPgm(t *testing.T) {
	t.Logf("%t", zmq.HasPgm())
	t.Logf("$PGM_MIN_LOG_LEVEL=%s", os.Getenv("PGM_MIN_LOG_LEVEL"))
}

func TestEMPG(t *testing.T) {
	var err error
	ctx, cancel := context.WithCancel(context.TODO())

	// multicastAddr := "epgm://eno1;172.16.20.91:6555"
	// multicastAddr := "epgm://eno1;172.16.20.91:6555"
	multicastAddr := "epgm://172.16.20.91:6555"
	listenAddr := "epgm://172.16.20.91:6555"
	// multicastAddr := "tcp://172.16.20.91:5555"
	chName := "echo"
	zctx, _ := zmq.NewContext()
	defer zctx.Term()
	pubSocket, err := zctx.NewSocket(zmq.PUB)
	if err != nil {
		t.Fatal(err)
	}
	defer pubSocket.Close()
	err = pubSocket.Bind(multicastAddr)
	if err != nil {
		t.Fatalf("bind err: %s", err)
	}

	subsocket, err := zctx.NewSocket(zmq.SUB)
	if err != nil {
		t.Fatal(err)
	}
	defer subsocket.Close()

	subsocket.SetSubscribe(chName)
	err = subsocket.Connect(listenAddr)
	if err != nil {
		t.Fatalf("connect err: %s", err)
	}

	go func() {
		time.Sleep(time.Second)
		total, err := pubSocket.SendMessage(chName, "hello")
		if err != nil {
			t.Logf("pub err: %s", err)
		} else {
			t.Logf("pub rst: %d", total)
		}

	}()

	go func() {
		msg, err := subsocket.RecvMessage(0)
		if err != nil {
			t.Logf("recv failed: %s", err)
		} else {
			t.Logf("Msg: %v", msg)
		}
		cancel()
	}()

	select {
	case <-ctx.Done():
		t.Logf("Done")
	case <-time.After(5 * time.Second):
		t.Fatal("timeout")
	}
}
