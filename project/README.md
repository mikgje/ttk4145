
# Elevator Project

Our philosophy for the project has been to restrict the use of different network packet types to its absolute minimum. This means we only have two message types on the network. One which sends out the distributed orders and one which communicates node statuses.

The project is divided into three main modules. Network, controller and single elevator. Each elevator has an associated controller which consitutes a network *node*. Each node communicates over UDP with all other nodes in a hybrid peer-to-peer architecture with master-slave decision making. One controller is master, while the others are slaves.

All nodes listen to the network and store all other nodes statuses locally. Only the master is allowed to make desicions on the information from the network, and uses this information to distribute incomming hall calls to suitable elevators. I.e. the master is the only node allowed to act on the information, and therefore, if the master does not know something, then it is not taken into account until the updated information reaches the master.

When a button is pressed on the elevator panel, a message is sent to the network. No lights (service guarnatees) are set until the node recieves orders back from network. It will then set relevant lights for itself, and the other nodes' orders. When a button is pressed, the controller adds the new request order to the node's service orders which are then sent to the master. The new order is, however, not immediately a service order for the local elevator to handle. This step is only done to inform the master of the new order without introducing new messages on the network.

Orders are confirmed on the network with their absence. The master redistributes all orders for every cycle, if the distribution differs from what it sent out previously, so as long as some node has an active order this order will either be returned to the same node or another node if deemed more efficient. When a node completes the order, it is removed from its queue and will not be sent to the master by any node and is therefore not redistributed and marked completed by all nodes.

Practical information:
We have tried to separate between code we have written/translated and provided (unchanged) code. This has been done by using snake-case for our code, while using pascal/camel case for the unchanged code we have been given.