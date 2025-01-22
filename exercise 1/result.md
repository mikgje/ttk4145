3:
Both the C and Go versions give "random" answers and therefore not the intended answer. This happens because theswapping is performed while each thread is executing. As such, we are dealing with race conditions. Both threads have stored the same previous value, instead of one thread modifying the value of another.
An example:
Stored value of i is 16
Incrementing thread increments this to 17
Decrementing thread decrements thihs to 15
Incrementing thread saves the value
Decrementing thread saves the value
The stored value is now 15

4:
C: Mutex is the correct choice. We want that the thread that locks the resource (the integer i in this case) is the only thread that is allowed to unlock the resource. Semaphores would allow other threads to unlock the resoure, which is undesired.
Go: There result is off by one sometimes. Not sure why.
