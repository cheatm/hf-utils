package zmq_test

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	zmq "github.com/pebbe/zmq4"
)

func TestBasicPubSub(t *testing.T) {
	go TestBasicPubServer(t)
	TestBasicSubClient(t)
}

func TestBasicPubServer(t *testing.T) {
	context, _ := zmq.NewContext()
	socket, _ := context.NewSocket(zmq.PUB)
	defer context.Term()
	defer socket.Close()
	socket.Bind("tcp://*:5556")

	zipcode := "59937"
	// loop for a while aparently
	for {

		//  make values that will fool the boss
		temperature := rand.Intn(215) - 80
		relhumidity := rand.Intn(50) + 10

		msg := fmt.Sprintf("%s %d %d", zipcode, temperature, relhumidity)
		t.Logf("[pub] %s", msg)
		// msgs := [][]byte{
		// 	[]byte(zipcode),
		// 	[]byte(fmt.Sprintf("%d %d", temperature, relhumidity)),
		// }
		//  Send message to all subscribers
		// _, err := socket.SendBytes([]byte(msg), 0)
		_, err := socket.SendMessage(
			[]byte(zipcode),
			[]byte(fmt.Sprintf("%d %d", temperature, relhumidity)),
		)
		// socket.SendMessage()
		if err != nil {
			t.Logf("[pub] send failed: %s", err)
			break
		}
		time.Sleep(time.Millisecond * 100)

	}

}

func TestBasicSubClient(t *testing.T) {
	context, _ := zmq.NewContext()
	socket, _ := context.NewSocket(zmq.SUB)
	defer context.Term()
	defer socket.Close()

	var err error
	var temps []string
	// var temp int64
	// total_temp := 0
	filter := "59937"

	//  Subscribe to just one zipcode (whitefish MT 59937)
	t.Logf("Collecting updates from weather server for %s…", filter)
	socket.SetSubscribe(filter)
	socket.Connect("tcp://localhost:5556")

	for i := 0; i < 101; i++ {
		// found temperature point
		temps, err = socket.RecvMessage(0)
		// datapt, err := socket.Recv(0)
		if err != nil {
			t.Fatalf("[recv] failed: %s", err)
		}
		t.Logf("[recv] %v", temps)
		// temps = strings.Split(string(datapt), " ")
		// temp, err = strconv.ParseInt(temps[1], 10, 64)
		// if err == nil {
		// 	// Invalid string
		// 	total_temp += int(temp)
		// }
	}

	// t.Logf("Average temperature for zipcode %s was %dF ", filter, total_temp/100)
}

func TestEchoSender(t *testing.T) {
	zctx, _ := zmq.NewContext()
	pubSocket, _ := zctx.NewSocket(zmq.PUB)
	defer zctx.Term()
	defer pubSocket.Close()
	pubSocket.Bind("tcp://*:5555")

	subSocket, err := zctx.NewSocket(zmq.SUB)
	if err != nil {
		t.Fatalf("[new subsocket] %s", err)
	}
	defer subSocket.Close()

	subSocket.SetSubscribe("echo")
	// subSocket.Connect("tcp://126.0.0.1:5556")
	subSocket.Connect("tcp://172.16.20.81:5556")
	time.Sleep(time.Millisecond * 200)
	reply := make(chan string)
	var count int = 10
	ctx, cancel := context.WithCancel(context.TODO())

	go func(count int) {
		for i := 0; i < count; i++ {
			ts := time.Now().UnixNano()
			pubSocket.SendMessage("echo", ts)
			t.Logf("send %d", ts)
			<-reply
		}
		cancel()
	}(count)
	var ttls int64 = 0
	go func() {
		for {

			msg, err := subSocket.RecvMessage(0)
			if err != nil {
				t.Logf("recv err: %s", err)
				break
			}
			now := time.Now().UnixNano()

			t.Logf("[reply] %v", msg)
			sendTime, err := strconv.ParseInt(msg[1], 10, 64)
			if err != nil {
				t.Logf("parse \"%s\" err: %s", msg[1], err)
			}

			latency := now - sendTime

			t.Logf("TTL = %d", latency)
			ttls += latency
			reply <- msg[1]
		}
	}()

	select {
	case <-ctx.Done():
		t.Logf("total: %d", ttls)
		t.Logf("avg: %d", ttls/int64(count))
	case <-time.After(time.Second * 5):
		t.Fatalf("timeout")
	}
}

func subscriber(zctx *zmq.Context, ttlCh chan int64, host string, subName string, clients int64, finished *atomic.Int64) (err error) {
	subSocket, err := zctx.NewSocket(zmq.SUB)
	if err != nil {
		fmt.Print(err)
		return
	}
	defer subSocket.Close()

	subSocket.SetSubscribe(subName)
	subSocket.Connect(host)

	var msg []string

	for {
		msg, err = subSocket.RecvMessage(0)
		if err != nil {
			fmt.Print(err)
			return
		}
		now := time.Now().UnixNano()
		if msg[1] == "stop" {
			break
		}

		sendTime, err := strconv.ParseInt(msg[1], 10, 64)
		if err != nil {
			fmt.Printf("parse \"%s\" err: %s\n", msg[1], err)
		}

		ttlCh <- (now - sendTime)
	}

	if finished.Add(1) == clients {
		close(ttlCh)
	}

	return
}

func TestEchoMultiClient(t *testing.T) {
	chName := "echo"
	zctx, _ := zmq.NewContext()
	pubSocket, _ := zctx.NewSocket(zmq.PUB)
	defer zctx.Term()
	defer pubSocket.Close()
	pubSocket.Bind("tcp://*:5555")
	subAddrs := []string{
		"tcp://172.16.20.81:5556",
		"tcp://172.16.20.81:5557",
		"tcp://172.16.20.81:5558",
		"tcp://172.16.20.81:5559",
	}
	finished := atomic.Int64{}

	msgCount := 100

	ttlCh := make(chan int64, 100)
	totalTTL := int64(0)
	totalCount := int64(0)

	for _, addr := range subAddrs {
		go subscriber(zctx, ttlCh, addr, chName, int64(len(subAddrs)), &finished)
	}

	// Publish
	go func() {
		time.Sleep(time.Second)
		for i := 0; i < msgCount; i++ {
			sendTs := time.Now().UnixNano()
			pubSocket.SendMessage(chName, sendTs)
			time.Sleep(time.Millisecond * 2)
		}
		pubSocket.SendMessage(chName, "stop")
	}()

	expireAt := time.Now().Add(time.Second * 10)
	timer := time.NewTimer(time.Until(expireAt))
	do := true
	for do {

		select {
		case ttl, ok := <-ttlCh:
			if ok {
				totalTTL += ttl
				totalCount++
			} else {
				do = false
			}
		case <-timer.C:
			do = false
		}
	}

	avgTTL := totalTTL / totalCount

	t.Logf("msg: %d", totalCount)
	t.Logf("avg: %d", avgTTL)

}

func TestChanClose(t *testing.T) {
	ch := make(chan int, 10)

	ch <- 1
	t.Logf("%d", <-ch)
	ch <- 2
	close(ch)
	var ok bool = true
	var d int
	for ok {
		select {
		case d, ok = <-ch:
			if ok {
				t.Logf("ok -> %d", d)
			} else {
				t.Logf("not ok")
			}
		case <-time.After(time.Second):
			t.Logf("timeout")
			ok = false
		}
	}

}
