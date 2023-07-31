package main

import (
	"fmt"
	"os"
	"time"

	zmq "github.com/pebbe/zmq4"
)

func RunPublishOne(pubAddr, op, topic, msg string) (err error) {
	context, err := zmq.NewContext()
	if err != nil {
		return fmt.Errorf("[new context] %w", err)
	}
	defer context.Term()
	pubSocket, err := context.NewSocket(zmq.PUB)
	if err != nil {
		return fmt.Errorf("[new pubsocket] %w", err)
	}
	defer pubSocket.Close()

	switch op {
	case "connect":
		err = pubSocket.Connect(pubAddr)
		if err != nil {
			err = fmt.Errorf("connect failed: %w", err)
			return
		}
	default:
		err = pubSocket.Bind(pubAddr)
		if err != nil {
			err = fmt.Errorf("bind failed: %s", err)
			return
		}
		fmt.Printf("bind %s ok\n", pubAddr)

	}

	for i := 0; i < 10; i++ {

		total, err := pubSocket.SendMessage(topic, msg)
		if err != nil {
			return fmt.Errorf("send failed: %s", err)
		}
		fmt.Printf("send %d\n", total)
		time.Sleep(time.Second)
	}

	return
}

func RunSubscribe(subAddr, op, topic string) (err error) {
	context, err := zmq.NewContext()
	if err != nil {
		return fmt.Errorf("[new context] %w", err)
	}
	defer context.Term()
	subSocket, err := context.NewSocket(zmq.SUB)
	if err != nil {
		return fmt.Errorf("[new pubsocket] %w", err)
	}
	defer subSocket.Close()
	subSocket.SetSubscribe(topic)
	switch op {
	case "bind":
		err = subSocket.Bind(subAddr)
		if err != nil {
			err = fmt.Errorf("bind failed: %s", err)
			return
		}
	default:
		err = subSocket.Connect(subAddr)
		if err != nil {
			err = fmt.Errorf("connect failed: %w", err)
			return
		}
	}

	for {
		msgs, err := subSocket.RecvMessage(0)
		if err != nil {
			fmt.Printf("recv failed: %s\n", err)
			continue
		}
		fmt.Printf("[msgs] %v\n", msgs)

	}

}

func RunPubSubEcho(pubAddr, subAddr string) (err error) {
	context, err := zmq.NewContext()
	if err != nil {
		return fmt.Errorf("[new context] %w", err)
	}
	defer context.Term()
	subSocket, err := context.NewSocket(zmq.SUB)
	if err != nil {
		return fmt.Errorf("[new subsocket] %w", err)
	}
	defer subSocket.Close()
	pubSocket, err := context.NewSocket(zmq.PUB)
	if err != nil {
		return fmt.Errorf("[new pubsocket] %w", err)
	}
	defer pubSocket.Close()

	subName := "echo"
	ch := make(chan []byte, 100)

	go func() {
		subSocket.SetSubscribe(subName)
		subSocket.Connect(subAddr)
		for {
			data, err := subSocket.RecvMessageBytes(0)
			if err != nil {
				fmt.Printf("sub err: %s\n", err)
				continue
			}
			// recvTime := time.Now().UnixNano()
			// sendTime, _ := strconv.ParseInt(string(data[1]), 10, 64)
			// fmt.Printf("%s, latency = %d\n", data[1], recvTime-sendTime)
			ch <- data[1]
		}
	}()

	pubSocket.Bind(pubAddr)
	for data := range ch {
		// fmt.Printf("recv %s\n", data)
		_, err := pubSocket.SendMessage(subName, data)
		if err != nil {
			fmt.Printf("[pub err]: %s", err)
		}
	}

	return

}

func main() {
	var err error
	args := os.Args
	if len(args) < 2 {
		args = []string{"", "echo", "tcp://*:5556", "tcp://127.0.0.1:5555"}
	}

	switch args[1] {
	case "echo":
		pubAddr := "tcp://*:5556"
		subAddr := "tcp://127.0.0.1:5555"
		if len(args) > 2 {
			pubAddr = args[2]
		}
		if len(args) > 3 {
			subAddr = args[3]
		}
		err = RunPubSubEcho(pubAddr, subAddr)
	case "pub":
		if len(args) < 6 {
			err = fmt.Errorf("not enough arguments, require 6")
		} else {
			err = RunPublishOne(args[2], args[3], args[4], args[5])
		}
	case "sub":
		if len(args) < 5 {
			err = fmt.Errorf("not enough arguments, require 5")
		} else {
			err = RunSubscribe(args[2], args[3], args[4])
		}
	}
	if err != nil {
		panic(err)
	}
}
