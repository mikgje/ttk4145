// Compile with `gcc foo.c -Wall -std=gnu99 -lpthread`, or use the makefile
// The executable will be named `foo` if you use the makefile, or `a.out` if you use gcc directly

#include <pthread.h>
#include <stdio.h>
#include <stdlib.h>

int i = 0;
pthread_t ptid1;
pthread_t ptid2;
pthread_mutex_t lock;

// Note the return type: void*
void* incrementingThreadFunction(){
    // TODO: increment i 1_000_000 times
    for (int j = 0; j<1000000; ++j) {
		pthread_mutex_lock(&lock);
		++i;
		pthread_mutex_unlock(&lock);
    }
    return NULL;
}

void* decrementingThreadFunction(){
    // TODO: decrement i 1_000_000 times
    for (int j = 0; j<1000000; ++j) {
		pthread_mutex_lock(&lock);
		--i;
		pthread_mutex_unlock(&lock);
	}
    return NULL;
}


int main(){
    // TODO: 
    // start the two functions as their own threads using `pthread_create`
    // Hint: search the web! Maybe try "pthread_create example"?

    pthread_create(&ptid1, NULL, &incrementingThreadFunction, NULL);
    pthread_create(&ptid2, NULL, &decrementingThreadFunction, NULL);
	pthread_mutex_init(&lock, NULL);

    // TODO:
    // wait for the two threads to be done before printing the final result
    // Hint: Use `pthread_join`   
    pthread_join(ptid1, NULL);
    pthread_join(ptid2, NULL);

    printf("The magic number is: %d\n", i);
    return 0;
}
