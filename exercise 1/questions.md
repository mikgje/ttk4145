Exercise 1 - Theory questions
-----------------------------

### Concepts

What is the difference between *concurrency* and *parallelism*?
> *Your answer here*
Concurrency is about executing multiple tasks simultaneously that share resources. Parellelism typically refers to independent tasks, where the goal is to run the tasks on independent hardware, such as on multiple cores.

What is the difference between a *race condition* and a *data race*? 
> *Your answer here*
A race condition is an error that occurs when two tasks update the same value.
A data race is the term used when data is accessed by at least two tasks, where at least one of them writes to the value. The tasks cannot have an ordering, where one task is done before the other. An example is one task writing 0b11 to a 2-bit register containing the value 0b00. A possible data race would be if the other task reads 0b01 (only one out of two write operations has had time to execute). 
 
*Very* roughly - what does a *scheduler* do, and how does it do it?
> *Your answer here*
A scheduler decides which thread to run. Scheduling can be divided into cooperative and preemptive scheduling. Preemptive scheduling is the common implementation of a scheduler. The scheduler calls the running thread to yield, i.e. making it stop and saving the context (register values and stack pointer). It then loads some other context belonging to another thread.


### Engineering

Why would we use multiple threads? What kinds of problems do threads solve?
> *Your answer here*
There are often situations where we want to run tasks that depend on each other. Running one task after the other will not give the desired result. For example in the elevator project, we would like to "continously" communicate with other elevators while servicing the users. Implementing this in a pure sequential fashion would yield either bad performance or messy code. If the elevator polls the status of the other elevators while moving, there would have to be a lot of polling inserted in the code to achieve good performance.
Using multiple threads solves real-time problems in an elegant fashion.

Some languages support "fibers" (sometimes called "green threads") or "coroutines"? What are they, and why would we rather use them over threads?
> *Your answer here*
Fibers are further divison of the program flow. One thread will consist of multiple fivbers running independant tasks. It achieves blocking-like performance while being non-blocking. Fibers are cooperative, while threads are generally preemptive.
Fibers are useful when running systems that have higher memory and CPU constraints.

Does creating concurrent programs make the programmer's life easier? Harder? Maybe both?
> *Your answer here*
It depends on the application. Concurrent programs are typically more difficult to introduce, but easier to maintain when more features are added.

What do you think is best - *shared variables* or *message passing*?
> *Your answer here*
It depends on the problem, such as message passing being an easier solution to bounded buffers, while shared variables being the easier solution for the inc/decrement problem.
