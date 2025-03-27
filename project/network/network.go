package network

/*-------------------------------------*/
// INPUT:
// net*: Network struct to collect and store all information relevant to the node and its perceived network
// service_orders_chan: Channel for receiving service orders from the master controller
// node_status_chan: Channel for receiving status messages from the local controller
/*-------------------------------------*/
// OUTPUT:
// node_statuses_chan: Channel for sending status messages from all controllers to the local controller
// send_orders_chan: Sending orders received on the network to local base controller
// dropped_peer_chan: Channel for sending dropped peer status messages to the local controller
/*-------------------------------------*/

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
	Master 				bool
	Connection			bool
	Ctrl_id				int	
	// System interface
	node_msg 			Node_msg
	statuses 			[utilities.N_ELEVS]utilities.Status_message
	dropped_peer_status	utilities.Status_message
	N_nodes				int
	// Peer-to-peer interface
	id					string
	nodes 				map[string]int
	other_peers			[]string
	alive_ids			[]string
	dropped_peer_id		string
	dropped_peer_flag	bool
}

type Node_msg struct {
	Label 	string
	ODM 	utilities.Order_distribution_message
	SM		utilities.Status_message
}

func Network_run(
	net* 				Network, 
	service_orders_chan <-chan utilities.Order_distribution_message, 
	send_orders_chan 	chan<- utilities.Order_distribution_message, 
	node_status_chan 	<-chan utilities.Status_message, 
	node_statuses_chan 	chan<- utilities.Status_message,
	dropped_peer_chan	chan<- utilities.Status_message,
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
	net.id = construct_network_id(id)

	peer_update_chan := make(chan peers.PeerUpdate)
	peer_tx_enable_chan := make(chan bool)
	go peers.Transmitter(*utilities.Peers, id, peer_tx_enable_chan)
	go peers.Receiver(*utilities.Peers, peer_update_chan)

	node_tx_chan := make(chan Node_msg)
	node_rx_chan := make(chan Node_msg)
	go bcast.Transmitter(*utilities.Bcast, node_tx_chan)
	go bcast.Receiver(*utilities.Bcast, node_rx_chan)

	initialize_statuses(net)

	go system_interface(net, node_tx_chan, node_rx_chan, service_orders_chan, send_orders_chan, node_status_chan, node_statuses_chan, dropped_peer_chan)
	go p2p_interface(net, id, peer_update_chan)

	for {}
}

func system_interface(
	net* Network, 
	node_tx_chan 		chan<- Node_msg, 
	node_rx_chan 		<-chan Node_msg, 
	service_orders_chan <-chan utilities.Order_distribution_message, 
	send_orders_chan 	chan<- utilities.Order_distribution_message, 
	node_status_chan 	<-chan utilities.Status_message, 
	node_statuses_chan 	chan<- utilities.Status_message,
	dropped_peer_chan	chan<- utilities.Status_message,
) {
	for {
		if net.Master {
			select {
			case assign := <-service_orders_chan:
				net.node_msg.ODM = assign
				net.node_msg.Label = "O"
				node_tx_chan <- net.node_msg
				net.node_msg.Label = ""
			default:
			}
			write_statuses(net.nodes, net.alive_ids, net.statuses, node_statuses_chan)
		}
		
		if net.dropped_peer_flag {
			all_nodes := sort_peers(append(net.other_peers, net.id, net.dropped_peer_id))
			net.dropped_peer_status = net.statuses[all_nodes[construct_network_id(net.dropped_peer_id)]]
			net.statuses[all_nodes[construct_network_id(net.dropped_peer_id)]] = utilities.Status_message{Controller_id: utilities.Default_id, Behaviour: utilities.Default_behaviour, Direction: utilities.Default_direction}
			dropped_peer_chan <- net.dropped_peer_status
			net.dropped_peer_flag = false
		}

		select {
		case new_status := <- node_status_chan:
			net.node_msg.SM = new_status
			net.node_msg.Label = net.id
			node_tx_chan <- net.node_msg
			net.node_msg.Label = ""
		case received := <- node_rx_chan:
			if received.Label == "O" {
				if !net.Master {
					net.node_msg.ODM = received.ODM
				}
				select {
				case send_orders_chan <- net.node_msg.ODM:
				default:
				}
			} else if ctrl_id, contains_label := net.nodes[received.Label]; contains_label {
				net.statuses[ctrl_id] = received.SM
			}
		default:
		}
		net.Ctrl_id = net.nodes[net.id]
	}
}

func initialize_statuses(net* Network) {
	for i := 0; i < utilities.N_ELEVS; i++ {
		net.statuses[i].Controller_id = utilities.Default_id
		net.statuses[i].Behaviour = utilities.Default_behaviour
		net.statuses[i].Direction = utilities.Default_direction
	}
}

func write_statuses(nodes map[string]int, alive_ids []string, statuses [utilities.N_ELEVS]utilities.Status_message, elevator_statuses chan<- utilities.Status_message) {
	if len(elevator_statuses) == 0 {
		for _, ids := range alive_ids {
			id := construct_network_id(ids)
			elevator_statuses <- statuses[nodes[id]]
		}
	}
}

func find_last_octet(ip string) string {
	return string(ip[strings.LastIndex(ip,".")+1:])
}

func construct_network_id(id string) string {
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
	net* 		Network, 
	id 				string, 
	peer_update 	chan peers.PeerUpdate,
) {
	for {
		select {
		case peers := <-peer_update:
			net.other_peers = find_other_peers(peers, id)
			net.nodes = sort_peers(append(net.other_peers, id))
			net.Master = decide_if_master(net.id, net.other_peers)
			net.Connection = check_connection(peers, id)
			net.N_nodes = len(net.nodes)
			if len(peers.New) > 0 {
				net.alive_ids = append(net.alive_ids, peers.New)
			}
			if len(peers.Lost) > 0 {
				if net.dropped_peer_id != peers.Lost[0] {
					net.dropped_peer_flag = true
				}
				net.dropped_peer_id = peers.Lost[0]
				net.alive_ids = utilities.Remove_from_slice(net.alive_ids, peers.Lost[0])
			} else {
				net.dropped_peer_id = ""
			}
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", peers.Peers)
			fmt.Printf("  New:      %q\n", peers.New)
			fmt.Printf("  Lost:     %q\n", peers.Lost)	
		}
	}
}

func find_other_peers(p peers.PeerUpdate, id string) []string {
	var other_peers []string
	for _, peer_id := range p.Peers {
		if peer_id != id && strings.Contains(peer_id, utilities.Network_prefix) {
			other_peers = append(other_peers, peer_id)
		}
	}
	return other_peers
}

func check_connection(p peers.PeerUpdate, id string) bool {
	for _, peer_id := range p.Lost {
		if peer_id == id {
			return false
		}
	}
	return true
}

func decide_if_master(id string, other_peers []string) bool {
	var lowest_id = id
	var peer_id string
	for i := range(len(other_peers)) {
		peer_id = construct_network_id(other_peers[i])
		if peer_id < lowest_id {
			lowest_id = peer_id
		}
	}
	if lowest_id < id {
		return false
	} else {
		return true
	}
}
