package network_master

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

type Node_msg struct {
	dist_msg utilities.OrderDistributionMessage
}

func network_master(assign_chan chan->OrderDistributionMessage, master_chan chan<-OrderDistributionMessage) {
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
		// TODO: implement channel logic to receive orders to be distributed from cost function
		node_msg := Node_msg{
			utilities.OrderDistributionMessage{
				"D",
				[3][elevator.N_FLOORS][elevator.N_BUTTONS-1]bool{
					//orderline0
					{
						{true, false},
						{false, true},
						{true, true},
						{false, false},
					},
					//orderline1
					{
						{true, false},
						{false, true},
						{true, true},
						{false, false},
					},
					//orderline2
					{
						{true, false},
						{false, true},
						{true, true},
						{false, false},
					},
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
			fmt.Printf("Received: %v\n", a)
//			fmt.Printf("\nIndex: %v\n", a.dist_msg.Orderline1)
			master_chan <- a	
		}
	}
}
