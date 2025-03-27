package network

//package main
// FAILURE: REMEMBER TO CHANGE PORTS!
// FAILURE: MAKE SURE THAT THERE AREN'T PROBLEMATIC SYNCHRONIZATION PROBLEMS BETWEEN READING MASTER IN NETWORK AND WRITING MASTER IN P2P
import (
	"Network-go/network/bcast"
	"Network-go/network/localip"
	"Network-go/network/peers"
	"fmt"
	"main/utilities"
	"sort"
	"strconv"
	"strings"
)

type Network struct {
	Master 		bool
	Connection	bool
	Ctrl_id		int	
	// System interface
	node_msg 	Node_msg
	statuses 	[utilities.N_ELEVS]utilities.StatusMessage
	lost_status	utilities.StatusMessage
	N_nodes		int
	// Peer-to-peer interface
	id			string
	nodes 		map[string]int
	others 		[]string
	alive_ids	[]string
	lost_id		string
	lost_flag	bool
}

type Node_msg struct {
	Label 	string
	ODM 	utilities.OrderDistributionMessage
	SM		utilities.StatusMessage
}

func Network_master(
	network* 			Network, 
	orders_to_assign 	<-chan utilities.OrderDistributionMessage, 
	assign_orders 		chan<- utilities.OrderDistributionMessage, 
	elevator_status 	<-chan utilities.StatusMessage, 
	elevator_statuses 	chan<- utilities.StatusMessage,
	lost_status			chan<- utilities.StatusMessage,
) {
	var id string
	id = *utilities.Id

	if id == "" {
		localIP, err := localip.LocalIP()
		if err != nil {
			fmt.Println(err)
			panic("Please provide an ID")
		}
		id = fmt.Sprintf("%s-%s-%s", utilities.Network_prefix, "AUTO", localIP)
	} else {
		id = fmt.Sprintf("%s-%s-%s", utilities.Network_prefix, "MANUAL", id)
	}
	network.id = construct_network_id(id)

	peer_update := make(chan peers.PeerUpdate)
	peer_tx_enable := make(chan bool)
	go peers.Transmitter(*utilities.Peers, id, peer_tx_enable)
	go peers.Receiver(*utilities.Peers, peer_update)

	node_tx := make(chan Node_msg)
	node_rx := make(chan Node_msg)
	go bcast.Transmitter(*utilities.Bcast, node_tx)
	go bcast.Receiver(*utilities.Bcast, node_rx)

	initialize_statuses(network)

	// Network and rest of system interface
	go network_interface(network, node_tx, node_rx, orders_to_assign, assign_orders, elevator_status, elevator_statuses, lost_status)

	// P2P and master-slave interface
	go p2p_interface(network, id, peer_update)

	for {
	}
}

func network_interface(
	network* Network, 
	node_tx 			chan<- Node_msg, 
	node_rx 			<-chan Node_msg, 
	orders_to_assign 	<-chan utilities.OrderDistributionMessage, 
	assign_orders 		chan<- utilities.OrderDistributionMessage, 
	elevator_status 	<-chan utilities.StatusMessage, 
	elevator_statuses 	chan<- utilities.StatusMessage,
	send_lost_status	chan<- utilities.StatusMessage,
) {
	for {
		if network.Master {
			select {
			case assign := <-orders_to_assign:
				network.node_msg.ODM = assign
				network.node_msg.Label = "O"
				node_tx <- network.node_msg
				network.node_msg.Label = ""
			default:
			}
			write_statuses(network.nodes, network.alive_ids, network.statuses, elevator_statuses)
		}

		if network.lost_flag {
			all_nodes := sort_peers(append(network.others, network.id, network.lost_id))
			network.lost_status = network.statuses[all_nodes[construct_network_id(network.lost_id)]]
			network.statuses[all_nodes[construct_network_id(network.lost_id)]] = utilities.StatusMessage{Controller_id: utilities.Default_id, Behaviour: utilities.Default_behaviour, Direction: utilities.Default_direction}
			send_lost_status <- network.lost_status
			network.lost_flag = false
		}
		select {
		case new_status := <- elevator_status:
			network.node_msg.SM = new_status
			network.node_msg.Label = network.id
			node_tx <- network.node_msg
			network.node_msg.Label = ""
		case received := <- node_rx:
			if received.Label == "O" {
				if !network.Master {
					network.node_msg.ODM = received.ODM
				}
				// TODO: check if bcast should halt (not use a default)
				select {
				case assign_orders <- network.node_msg.ODM:
				default:
				}
			} else if ctrl_id, contains_label := network.nodes[received.Label]; contains_label {
				// TODO: Index error when elevators > 2
				network.statuses[ctrl_id] = received.SM
			}
		default:
		}
		network.Ctrl_id = network.nodes[network.id]
	}
}

func initialize_statuses(network* Network) {
	for i := 0; i < utilities.N_ELEVS; i++ {
		network.statuses[i].Controller_id = utilities.Default_id
		network.statuses[i].Behaviour = utilities.Default_behaviour
		network.statuses[i].Direction = utilities.Default_direction
	}
}

func write_statuses(nodes map[string]int, alive_ids []string, statuses [utilities.N_ELEVS]utilities.StatusMessage, elevator_statuses chan<- utilities.StatusMessage) {
	if len(elevator_statuses) == 0 {
		// TODO: presumed that own ID is i = 0, since this will be used only(?) when master. Problem?
		for _, ids := range alive_ids {
			id := construct_network_id(ids)
			elevator_statuses <- statuses[nodes[id]]
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
	if strings.Contains(id, "MANUAL") {
		return string(id[strings.LastIndex(id,"-")+1:])
	} else {
		return strings.ReplaceAll(find_last_octet(id), "-", "")
	}
}

func sort_peers(ids []string) map[string]int {
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

func p2p_interface(
	network* 		Network, 
	id 				string, 
	peer_update 	chan peers.PeerUpdate,
) {
	for {
		select {
		case peers := <-peer_update:
			network.others = find_others(peers, id)
			network.nodes = sort_peers(append(network.others, id))
			network.Master = decide_master(network.id, network.others)
			network.Connection = check_connection(peers, id)
			network.N_nodes = len(network.nodes)
			if len(peers.New) > 0 {
				network.alive_ids = append(network.alive_ids, peers.New)
				// fmt.Println("Alive peers: ", network.alive_ids)
				// for _, ids := range network.alive_ids {
				// 	id := construct_network_id(ids)
				// 	fmt.Println("IDs: ", network.nodes[id])
				// }
			}
			if len(peers.Lost) > 0 {
				if network.lost_id != peers.Lost[0] {
					network.lost_flag = true
				}
				network.lost_id = peers.Lost[0]
				network.alive_ids = remove(network.alive_ids, peers.Lost[0])
				// fmt.Println("Alive peers after lost: ", network.alive_ids)
				// for _, ids := range network.alive_ids {
				// 	id := construct_network_id(ids)
				// 	fmt.Println("IDs after lost: ", network.nodes[id])
				// }
			// TODO: check if below is necessary
			} else {
				network.lost_id = ""
			}
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", peers.Peers)
			fmt.Printf("  New:      %q\n", peers.New)
			fmt.Printf("  Lost:     %q\n", peers.Lost)	
		}
	}
}

func find_others(p peers.PeerUpdate, id string) []string {
	var others []string
	for _, element := range p.Peers {
		if element != id && strings.Contains(element, utilities.Network_prefix) {
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

func decide_master(id string, others []string) bool {
	var lowest_id = id
	var element_id string
	for i := range(len(others)) {
		element_id = construct_network_id(others[i])
		if element_id < lowest_id {
			lowest_id = element_id
		}
	}
	if lowest_id < id {
		return false
	} else {
		return true
	}
}

func contains(slice []string, item string) bool {
    for _, element := range slice {
        if element == item {
            return true
        }
    }
    return false
}

func remove(slice []string, item string) []string {
    for i, element := range slice {
        if element == item {
            return append(slice[:i], slice[i+1:]...)
        }
    }
    return slice
}