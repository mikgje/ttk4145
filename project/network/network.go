//package network_master
package main
// FAILURE: REMEMBER TO CHANGE PORTS!
// FAILURE: MAKE SURE THAT THERE AREN'T PROBLEMATIC SYNCHRONIZATION PROBLEMS BETWEEN READING MASTER IN NETWORK AND WRITING MASTER IN P2P
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

// NOT MY PROBLEM?: implement extract_orderline functionality
// DONE: implement network struct to shorten function calls, with global accesses
// TODO: must use localIP at some point, since same PID can be assign on different nodes
// DONE?: make other status chan
// DONE: make map

type Network struct {
	master		bool
	// System interface
	node_msg 	Node_msg
	incm_msg	Node_msg
	statuses 	[utilities.N_ELEVS]utilities.StatusMessage
	// Peer-to-peer interface
	id			string
	nodes 		peers.PeerUpdate
	others 		[]string
}

type Node_msg struct {
	Label 	string
	ODM 	utilities.OrderDistributionMessage
	SM		utilities.StatusMessage
}

// For testing purposes
func main() {
	// Channel to receive service orders
	assign_chan := make(chan utilities.OrderDistributionMessage)
	// Channel to transmit service orders
	bcast_sorders_chan := make(chan utilities.OrderDistributionMessage)
	// Channel to receive local elevator status
	controller_chan := make(chan utilities.StatusMessage)
	// Channel to transmit all statuses except local
	status_chan := make(chan utilities.StatusMessage, utilities.N_ELEVS-1)

	go network(assign_chan, bcast_sorders_chan, controller_chan, status_chan)
	for {	
		assign_chan <- utilities.OrderDistributionMessage{Orderlines : [3][utilities.N_FLOORS][utilities.N_BUTTONS-1]bool{
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

func network(assign_chan <-chan utilities.OrderDistributionMessage, bcast_sorders_chan chan<- utilities.OrderDistributionMessage, controller_chan <-chan utilities.StatusMessage, status_chan chan<- utilities.StatusMessage) {
	var network Network
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
		network.id = construct_network_id(id)
	}

	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)
	go peers.Transmitter(15647, id, peerTxEnable)
	go peers.Receiver(15647, peerUpdateCh)

	node_tx := make(chan Node_msg)
	node_rx := make(chan Node_msg)
	go bcast.Transmitter(16569, node_tx)
	go bcast.Receiver(16569, node_rx)


	// Network and rest of system interface
	go network_interface(&network, node_tx, node_rx, assign_chan, bcast_sorders_chan, controller_chan, status_chan)

	// P2P and master-slave interface
	go p2p_interface(&network, id, peerUpdateCh)

	for {
//		fmt.Println("Sorted", sort_peers(network.nodes))
		fmt.Println("Sorted map", sort_peers2(network.nodes))
//		fmt.Println(network)
//		fmt.Println(find_last_octet(Ip)+find_pid(id))
//		fmt.Println(network.id)
//		fmt.Println("Network id", construct_network_id(network.id))
		time.Sleep(time.Second)
	}
}

// DONE: implement logic such that master loads the message with ODM, all nodes loads ODM into channel to be sent to controller.
// DONE: implement logic such that nodes only reads from master
// NOT RELEVANT: implement logic to specify targets for order_assigner
// TODO: implement logic to transmit status over network and locally through channels
func network_interface(network* Network, node_tx chan Node_msg, node_rx chan Node_msg, assign_chan <-chan utilities.OrderDistributionMessage, bcast_sorders_chan chan<- utilities.OrderDistributionMessage, controller_chan <-chan utilities.StatusMessage, status_chan chan<- utilities.StatusMessage) {
	for {
		// Only master
		if network.master {
			select {
			case assign := <-assign_chan:
				network.node_msg.ODM = assign
				network.node_msg.Label = "O"
				node_tx <- network.node_msg
				network.node_msg.Label = ""
			default:
			}
			write_statuses(network.statuses, status_chan)
		} else {
			// Only overwrite ODM if not master	
		}
		// Everyone
		select {
		// Update the status if the elevator has sent a new status
		case new_status := <- controller_chan:
			network.node_msg.SM = new_status
			network.node_msg.Label = network.id
			node_tx <- network.node_msg
			network.node_msg.Label = ""
//		case bcast_sorders_chan <- network.incm_msg.ODM:
		case received := <- node_rx:
			if received.Label == "O" {
				if !(network.master) {
					network.node_msg.ODM = received.ODM
				}
				bcast_sorders_chan <- network.node_msg.ODM
			} else if ctrl_id, contains_label := sort_peers2(network.nodes)[received.Label]; contains_label {
				network.statuses[ctrl_id] = received.SM
			}
		default:
		}
//		node_tx <- network.node_msg
		time.Sleep(1 * time.Second)
	}
}

// DONE: move node_rx to network_interface
func p2p_interface(network* Network, id string, peerUpdateCh chan peers.PeerUpdate) {
	for {
		select {
		case network.nodes = <-peerUpdateCh:
			network.others = find_others(network.nodes, id)
			network.master = decide_master(network.id, network.others)
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", network.nodes.Peers)
			fmt.Printf("  New:      %q\n", network.nodes.New)
			fmt.Printf("  Lost:     %q\n", network.nodes.Lost)	
		}
	}
}

func write_statuses(statuses [utilities.N_ELEVS]utilities.StatusMessage, status_chan chan<- utilities.StatusMessage) {
	if len(status_chan) == 0 {
		for i := 1; i < utilities.N_ELEVS; i++ {
			status_chan <- statuses[i]
		}
	}
}

func find_pid(id string) string {
	return string(id[strings.LastIndex(id, "-")+1:])
}

func find_last_octet(ip string) string {
	return string(ip[strings.LastIndex(ip,".")+1:])
}

func construct_network_id (id string) string {
	if strings.Contains(id, "DISCONNECTED") {
		return find_pid(id)
	} else {
		return strings.ReplaceAll(find_last_octet(id), "-", "")
	}
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

// Should be map use keys as type strings instead, i.e. drop conversion?
/*
func sort_peers2(p peers.PeerUpdate) map[int]int {
	ids := p.Peers
	nums := make([]int, len(ids))
	idm := make(map[int]int)
	for index, element := range ids {
		num, err := strconv.Atoi(find_pid(element))
		if err != nil {
			panic("Failed to convert str to int in sort_peers")
		}
		nums[index] = num
	}
	sort.Ints(nums)
	for index, element := range nums {
		idm[element] = index
	}
	return idm
}
*/

// TODO: rename from PIDs, were not using pids anymore
func sort_peers2(p peers.PeerUpdate) map[string]int {
	ids := p.Peers
	network_ids := make([]int, len(ids))
	idm := make(map[string]int)
	for index, element := range ids {
		network_id, err := strconv.Atoi(construct_network_id(element))
		if err != nil {
			panic("Failed to convert str to int in sort_peers")
		}
		network_ids[index] = network_id
	}
	sort.Ints(network_ids)
	for index, element := range network_ids {
		str_element := strconv.Itoa(element)
		idm[str_element] = index
	}
	return idm
}
