package main

import (
	"fmt"
	"os"

	zmq "github.com/pebbe/zmq4"
)

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
		RunPubSubEcho(pubAddr, subAddr)
	}
}
