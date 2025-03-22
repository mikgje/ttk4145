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

// Node message used by the slaves containing a status message
type Node_msg struct {
	Status_msg utilities.StatusMessage
}

// For testing purposes, not finished
func main() {
	var id string
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			localIP = "DISCONNECTED"
		}
		id = fmt.Sprintf("peer-%s-%d", localIP, os.Getpid())
	}

	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)
	go peers.Transmitter(15647, id, peerTxEnable)
	go peers.Receiver(15647, peerUpdateCh)

	node_tx := make(chan Node_msg)
	node_rx := make(chan Node_msg)
	go bcast.Transmitter(16569, node_tx)
	go bcast.Receiver(16569, node_rx)

	go func() {
		// Test value
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
