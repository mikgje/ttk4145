package network_slave

import (
	"Network-go/network/bcast"
	"Network-go/network/localip"
	"Network-go/network/peers"
	"main/utilities"
	"main/elev_algo_go/elevator"
	"flag"
	"fmt"
	"os"
	"time"
)

// We define some custom struct to send over the network.
// Note that all members we want to transmit must be public. Any private members
//  will be received as zero-values.
type Node_msg struct {
	Status_msg utilities.StatusMessage
}

func main() {
	// Our id can be anything. Here we pass it on the command line, using
	//  `go run main.go -id=our_id`
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	// ... or alternatively, we can use the local IP address.
	// (But since we can run multiple programs on the same PC, we also append the
	//  process ID)
	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}

	// We make a channel for receiving updates on the id's of the peers that are
	//  alive on the network
	peerUpdateCh := make(chan peers.PeerUpdate)
	// We can disable/enable the transmitter after it has been started.
	// This could be used to signal that we are somehow "unavailable".
	peerTxEnable := make(chan bool)
	go peers.Transmitter(15647, id, peerTxEnable)
	go peers.Receiver(15647, peerUpdateCh)

	// We make channels for sending and receiving our custom data types
	node_tx := make(chan Node_msg)
	node_rx := make(chan Node_msg)
	// ... and start the transmitter/receiver pair on some port
	// These functions can take any number of channels! It is also possible to
	//  start multiple transmitters/receivers on the same port.
	go bcast.Transmitter(16569, node_tx)
	go bcast.Receiver(16569, node_rx)

	// The example message. We just send one of these every second.
	go func() {
		// TODO: implement channel logic to receive status from elevator
		node_msg := Node_msg{
			utilities.StatusMessage{
				"S",1,"test",1,"up",[elevator.N_FLOORS][elevator.N_BUTTONS]bool{
					{true, false},
					{false, true},
					{true, true},
					{false, false},
				},
			},
		}
		for {
			node_tx <- node_msg
			time.Sleep(1 * time.Second)
		}
	}()

	fmt.Println("Started")
	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

		case a := <- node_rx:
			fmt.Printf("Received: %#v\n", a)
		}
	}
}
