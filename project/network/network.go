//package network_master
package main

import (
	"Network-go/network/bcast"
	"Network-go/network/localip"
	"Network-go/network/peers"
	"main/utilities"
	"flag"
	"fmt"
	"os"
	"time"
	"strings"
	"sort"
	"strconv"
)

// TODO: implement extract_orderline functionality
// TODO: implement network struct to shorten function calls
// TODO: must use localIP at some point, since same PID can be assign on different nodes

type Node_msg struct {
	Label 		string
	Target 		string
	Dist_msg 	utilities.OrderDistributionMessage
	Status_msg	utilities.StatusMessage
}

// For testing purposes
func main() {
	assign_chan := make(chan utilities.OrderDistributionMessage)
	bcast_sorders_chan := make(chan utilities.OrderDistributionMessage)
	elevator_chan := make(chan utilities.StatusMessage)

	go Network(assign_chan, bcast_sorders_chan, elevator_chan)
	for {	
		assign_chan <- utilities.OrderDistributionMessage{Label : "Ø", Orderlines : [3][utilities.N_FLOORS][utilities.N_BUTTONS-1]bool{
			{	{true,false},
				{false,true},
				{true,true},
				{false,false},
			},
			{	{true,false},
				{false,true},
				{true,true},
				{false,false},
			},
			{	{true,false},
				{false,true},
				{true,true},
				{false,false},
			},
		}	}
	
	}
}

func Network(assign_chan <-chan utilities.OrderDistributionMessage, bcast_sorders_chan chan<- utilities.OrderDistributionMessage, elevator_chan <-chan utilities.StatusMessage) {
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

	// Must be used as a pointer to have make updates inside of function affect outside scope
	var master		bool
	var master_msg  Node_msg = Node_msg{"M", "", utilities.OrderDistributionMessage{}, utilities.StatusMessage{}}
	var node_msg 	Node_msg = Node_msg{id, "", utilities.OrderDistributionMessage{}, utilities.StatusMessage{}}
	var inc_msg		Node_msg

	// Network and rest of system interface
	go network_interface(&master, id, node_tx, node_rx, &node_msg, &master_msg, &inc_msg, assign_chan, bcast_sorders_chan, elevator_chan)

	var p 			peers.PeerUpdate
	var others 		[]string
	var other_id	[]string

	// P2P and master-slave interface
	go p2p_interface(&master, id, peerUpdateCh, &p, others, other_id)

	for {
		fmt.Println("Sorted", sort_peers(p))
		time.Sleep(time.Second)
	}
}

// DONE: implement logic such that master loads the message with Dist_msg, all nodes loads Dist_msg into channel to be sent to controller.
// DONE: implement logic such that nodes only reads from master
// TODO: implement logic to specify targets for order_assigner
// TODO: implement logic to transmit status over network and locally through channels
func network_interface(master* bool, id string, node_tx chan Node_msg, node_rx chan Node_msg, node_msg* Node_msg, master_msg* Node_msg, inc_msg* Node_msg, assign_chan <-chan utilities.OrderDistributionMessage, bcast_sorders_chan chan<- utilities.OrderDistributionMessage, elevator_chan <-chan utilities.StatusMessage) {
	for {
		// Only master
		if *master {
//			fmt.Println("Master", *master_msg)
			select {
			case assign := <-assign_chan:
				// TODO: specify target
				master_msg.Target = id
				master_msg.Dist_msg = assign
				node_tx <- *master_msg
			default:
			}
//			node_tx <- *node_msg
		}
		// Everyone
		select {
		// Update the status if the elevator has sent a new status
		case new_status := <- elevator_chan:
			node_msg.Status_msg = new_status
		case *inc_msg = <- node_rx:
			// TODO: fix target specification
			if (inc_msg.Label == "M") && (inc_msg.Target == id) {
				node_msg.Dist_msg = inc_msg.Dist_msg
//				fmt.Println("Master", inc_msg.Dist_msg)
//				fmt.Println("Local ", node_msg.Dist_msg)
			}
//			fmt.Println(*inc_msg)
		case bcast_sorders_chan <- inc_msg.Dist_msg:
		default:
		}
		node_tx <- *node_msg
		time.Sleep(1 * time.Second)
	}
}

// DONE: move node_rx to network_interface
func p2p_interface(master* bool, id string, peerUpdateCh chan peers.PeerUpdate, p* peers.PeerUpdate, others []string, other_id []string) {
	for {
		other_id = p.Peers
//		fmt.Println(others)
		select {
		case *p = <-peerUpdateCh:
			other_id = p.Peers
			others = find_others(*p, id)
			*master = decide_master(find_pid(id),others)
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)	
		}
//		fmt.Println("Master:", *master)
	}
}

func find_pid(id string) string {
	other_pid := string(id[strings.LastIndex(id, "-")+1:])
	return other_pid
}

func find_others(p peers.PeerUpdate, id string) []string {
	var others []string
	for _, element := range p.Peers {
		if element != id {
			others = append(others, element)
		}
	}
	return others
}

func decide_master(pid string, others []string) bool {
	var lowest_pid = pid
	var element_pid string
	for index := range(len(others)) {
		element_pid = find_pid(others[index])
		if element_pid < lowest_pid {
			lowest_pid = element_pid
		}
	}
	if lowest_pid < pid {
		return false
	} else {
		return true
	}
}

// Function to be used by assigner to get have ids sorted by lowest to highest
func sort_peers(p peers.PeerUpdate) []int {
	ids := p.Peers
	nums := make([]int, len(ids))
	for index, element := range ids {
		num, err := strconv.Atoi(find_pid(element))
		if err != nil {
			panic("Failed to convert str to int in sort_peers")
		}
		nums[index] = num
	}
	sort.Ints(nums)
	return nums
}
