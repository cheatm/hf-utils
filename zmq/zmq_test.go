package zmq

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/zeromq/goczmq"
)

func TestHello(t *testing.T) {
	// Create a router socket and bind it to port 5555.
	router, err := goczmq.NewRouter("tcp://*:5555")
	log := t
	if err != nil {
		log.Fatal(err)
	}
	defer router.Destroy()

	log.Log("router created and bound")

	// Create a dealer socket and connect it to the router.
	dealer, err := goczmq.NewDealer("tcp://127.0.0.1:5555")
	if err != nil {
		log.Fatal(err)
	}
	defer dealer.Destroy()

	log.Log("dealer created and connected")

	// Send a 'Hello' message from the dealer to the router.
	// Here we send it as a frame ([]byte), with a FlagNone
	// flag to indicate there are no more frames following.
	err = dealer.SendFrame([]byte("Hello"), goczmq.FlagNone)
	if err != nil {
		log.Fatal(err)
	}

	log.Log("dealer sent 'Hello'")

	// Receive the message. Here we call RecvMessage, which
	// will return the message as a slice of frames ([][]byte).
	// Since this is a router socket that support async
	// request / reply, the first frame of the message will
	// be the routing frame.
	request, err := router.RecvMessage()
	if err != nil {
		log.Fatal(err)
	}

	log.Logf("router received '%s' from '%v'", request[1], request[0])

	// Send a reply. First we send the routing frame, which
	// lets the dealer know which client to send the message.
	// The FlagMore flag tells the router there will be more
	// frames in this message.
	err = router.SendFrame(request[0], goczmq.FlagMore)
	if err != nil {
		log.Fatal(err)
	}

	log.Logf("router sent 'World'")

	// Next send the reply. The FlagNone flag tells the router
	// that this is the last frame of the message.
	err = router.SendFrame([]byte("World"), goczmq.FlagNone)
	if err != nil {
		log.Fatal(err)
	}

	// Receive the reply.
	reply, err := dealer.RecvMessage()
	if err != nil {
		log.Fatal(err)
	}

	log.Logf("dealer received '%s'", string(reply[0]))
}

func TestCZMQPubSub(t *testing.T) {
	count := 10
	subscribe := "echo"
	pubsock, err := goczmq.NewPub("tcp://*:5556")
	if err != nil {
		t.Fatal(err)
	}
	defer pubsock.Destroy()

	subsock, err := goczmq.NewSub("tcp://localhost:5556", subscribe)
	if err != nil {
		t.Fatal(err)
	}

	defer subsock.Destroy()

	go func() {
		time.Sleep(time.Second)

		for i := 0; i < count; i++ {
			now := time.Now().UnixNano()
			err := pubsock.SendMessage([][]byte{
				[]byte(subscribe),
				[]byte(fmt.Sprintf("%d", now)),
			})
			if err != nil {
				t.Logf("send err: %s", err)
			} else {
				t.Logf("msg sent")
			}
		}

	}()

	for i := 0; i < count; i++ {
		msgs, err := subsock.RecvMessage()
		if err != nil {
			t.Fatal(err)
		}
		recvTime := time.Now().UnixNano()

		// for i, v := range msgs {
		// 	t.Logf("recv [%d] %s", i, v)
		// }

		msg := string(msgs[1])
		sendTime, _ := strconv.ParseInt(msg, 10, 64)

		d := recvTime - sendTime
		t.Logf("latency: %d", d)

	}

}
