Elevator software
===========================

(See attached figures to get a graphical representation of what is presented here)

The elevator software to run on a single computer is comprised of three main modules (outside the main function used to call them and run the program): main_elevator, main_controller and network.

main_elevator and main_controller are part of the main package to allow for swift sharing of constants and flags. Variables are shared using message passing over channels.

Terminology to understand. In our architecture we distinguish between service orders and request orders. Service orders are orders supposed to be executed by an elevator (given by the master/primary). Request orders are requests registered from the button panel of the elevator, and sent up the chain to the controller and onwards to the master/primary. The request order is not added to the elevators queue. Eventually the request will be registered by the master/primary, and then be turned into a service order that one of the elevators will receive through its controller.

The main_elevator and main_controller modules are cooperative. The main_elevator controls the movement of the elevator and processes the I/O information from the button panel. This information is then passed on to the controller. The controller is sort of a local "administrator". It receives status updates from the elevator, and chooses actions to take based on this. The controller is also the interface between the elevator and the network.

Looking more closely at the controller, one will see that it is run as a state machine. Its behaviour will depend on the state.

When the elevator receives a new request order it will notify the controller. The controller will then augment the status of the elevator, by adding the requestorder to its status message (as if it was part of the elevator queue). This is used to reduce the number of message types, but the elevator is not adding this order to its queue yet. The master will then receive status messages from all controllers and create an overview of all active hall calls on the system. This is passed on to an assigner function, which reassigns all orders. It will then output an orderline containing all the orders and broadcast this to the network. Each controller will then extract its orderline (based on controller ID) and send these orders to its elevator to be added to the queue.

Looking at the network module, we see that there are two packages. When the controller is in any other mode than primary, it will be running the network_slave functionality. This allows the controller to send satus messages to the network, and receive orderlines. When the controller is in the primary mode, however, it has the added funcitonality of compiling the status messages and using the hall call assigner to create the orderline message and sending it to the network.