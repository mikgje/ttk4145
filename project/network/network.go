package network
//package main
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
	//"time"
	"strings"
	"sort"
	"strconv"
)

// NOT MY PROBLEM?: implement extract_orderline functionality
// DONE: implement network struct to shorten function calls, with global accesses
// DONE: must use localIP at some point, since same PID can be assign on different nodes
// DONE?: make other status chan
// DONE: make map
// TODO: find out if its an issue with status message using ctrl_id of type int

type Network struct {
	Master 		bool
	Connection	bool
	Ctrl_id		int	
	// System interface
	node_msg 	Node_msg
	statuses 	[utilities.N_ELEVS]utilities.StatusMessage
	N_nodes		int
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

/*
// For testing purposes
func main() {
	// Channel to receive service orders TODO: decide if there either should be a goroutine permanently writing/reading, or if the channel should be used with a buffer size 1
	assign_chan := make(chan utilities.OrderDistributionMessage, 1)
	// Channel to transmit service orders
	bcast_sorders_chan := make(chan utilities.OrderDistributionMessage)
	// Channel to receive local elevator status
	controller_chan := make(chan utilities.StatusMessage)
	// Channel to transmit all statuses except local
	status_chan := make(chan utilities.StatusMessage, utilities.N_ELEVS-1)

	go Network(assign_chan, bcast_sorders_chan, controller_chan, status_chan)
	for {	
		select {
		case assign_chan <- utilities.OrderDistributionMessage{Orderlines : [3][utilities.N_FLOORS][utilities.N_BUTTONS-1]bool{
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
		}	}:
		default:
		}
		controller_chan <- utilities.StatusMessage{Controller_id : 5, Behaviour : "Dunno", Floor : 2, Direction : "UP", Node_orders : [utilities.N_FLOORS][utilities.N_BUTTONS]bool{}}
		time.Sleep(3*time.Second)
	}
}
*/

func Network_master(network* Network, assign_chan <-chan utilities.OrderDistributionMessage, bcast_sorders_chan chan<- utilities.OrderDistributionMessage, controller_chan <-chan utilities.StatusMessage, status_chan chan<- utilities.StatusMessage) {
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
	go peers.Transmitter(*utilities.Peers, id, peerTxEnable)
	go peers.Receiver(*utilities.Peers, peerUpdateCh)

	node_tx := make(chan Node_msg)
	node_rx := make(chan Node_msg)
	go bcast.Transmitter(*utilities.Bcast, node_tx)
	go bcast.Receiver(*utilities.Bcast, node_rx)

	initialize_statuses(network)

	// Network and rest of system interface
	go network_interface(network, node_tx, node_rx, assign_chan, bcast_sorders_chan, controller_chan, status_chan)

	// P2P and master-slave interface
	go p2p_interface(network, id, peerUpdateCh)

	for {
	}
}

// DONE: implement logic such that master loads the message with ODM, all nodes loads ODM into channel to be sent to controller.
// DONE: implement logic such that nodes only reads from master
// NOT RELEVANT: implement logic to specify targets for order_assigner
// DONE: implement logic to transmit status over network and locally through channels
func network_interface(network* Network, node_tx chan Node_msg, node_rx chan Node_msg, assign_chan <-chan utilities.OrderDistributionMessage, bcast_sorders_chan chan<- utilities.OrderDistributionMessage, controller_chan <-chan utilities.StatusMessage, status_chan chan<- utilities.StatusMessage) {
	for {
		// Only master
		if network.Master {
			select {
			case assign := <-assign_chan:
				network.node_msg.ODM = assign
				network.node_msg.Label = "O"
				node_tx <- network.node_msg
				network.node_msg.Label = ""
			default:
			}
			// TODO: A hasnt implemented status_chan yet, this will halt
			write_statuses(network.statuses, status_chan)
		}
		// Everyone
		select {
		// Update the status if the elevator has sent a new status
		case new_status := <- controller_chan:
			network.node_msg.SM = new_status
			network.node_msg.Label = network.id
			node_tx <- network.node_msg
			network.node_msg.Label = ""
		case received := <- node_rx:
			//fmt.Println(received)
			if received.Label == "O" {
				if !network.Master {
					network.node_msg.ODM = received.ODM
				}
				// TODO: check if bcast should halt (not use a default)
				select {
				case bcast_sorders_chan <- network.node_msg.ODM:
				default:
				}
			} else if ctrl_id, contains_label := sort_peers(network.nodes)[received.Label]; contains_label {
				// TODO: Index error when elevators > 2
				network.statuses[ctrl_id] = received.SM
			}
		default:
		}
		network.Ctrl_id = sort_peers(network.nodes)[network.id]
//		node_tx <- network.node_msg
		//time.Sleep(100 * time.Millisecond)
	}
}

// DONE: move node_rx to network_interface
func p2p_interface(network* Network, id string, peerUpdateCh chan peers.PeerUpdate) {
	for {
		select {
		case network.nodes = <-peerUpdateCh:
			network.others = find_others(network.nodes, id)
			network.Master = decide_master(network.id, network.others)
			network.Connection = check_connection(network.nodes, id)
			network.N_nodes = len(network.nodes.Peers)
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", network.nodes.Peers)
			fmt.Printf("  New:      %q\n", network.nodes.New)
			fmt.Printf("  Lost:     %q\n", network.nodes.Lost)	
		}
	}
}

func initialize_statuses(network* Network) {
	for i := 0; i < utilities.N_ELEVS; i++ {
		network.statuses[i].Controller_id = utilities.Default_id
		network.statuses[i].Behaviour = utilities.Default_behaviour
		network.statuses[i].Direction = utilities.Default_direction
	}
}

func write_statuses(statuses [utilities.N_ELEVS]utilities.StatusMessage, status_chan chan<- utilities.StatusMessage) {
	if len(status_chan) == 0 {
		// TODO: presumed that own ID is i = 0, since this will be used only(?) when master. Problem?
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

func check_connection(p peers.PeerUpdate, id string) bool {
	for _, element := range p.Lost {
		if element == id {
			return false
		}
	}
	return true
}

func decide_master(pid string, others []string) bool {
	var lowest_pid = pid
	var element_pid string
	for index := range(len(others)) {
		element_pid = construct_network_id(others[index])
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

func sort_peers(p peers.PeerUpdate) map[string]int {
	ids := p.Peers
	network_ids := make([]int, len(ids))
	idm := make(map[string]int)
	for i, id := range ids {
		network_id, err := strconv.Atoi(construct_network_id(id))
		if err != nil {
			panic("Failed to convert str to int in sort_peers")
		}
		network_ids[i] = network_id
	}
	sort.Ints(network_ids)
	for i, network_id := range network_ids {
		idm[strconv.Itoa(network_id)] = i
	}
	return idm
}
