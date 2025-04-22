import std.algorithm, std.concurrency, std.format, std.range, std.stdio, std.traits;
import core.thread, core.sync.semaphore, core.sync.mutex, core.sync.condition;

immutable Duration tick = 33.msecs;

// --- RESOURCE CLASS --- //
/* 
Implementasjon med semaforer:
- mtx brukes som en lås (initialverdi 1).
- sems[0] og sems[1] er køene for lav og høy prioritet.
- numWaiting[0] og numWaiting[1] teller antall ventende tråder for henholdsvis lav og høy prioritet.
- busy angir om ressursen er i bruk.
*/
class Resource(T) {
    private {
        T               value;
        Semaphore       mtx;
        Semaphore[2]    sems;
        int[2]          numWaiting;
        bool            busy;
    }
    
    
    this(){
        mtx = new Semaphore(1);
        foreach(ref sem; sems){
            sem = new Semaphore(0);
        }
        busy = false;
        numWaiting[0] = 0;
        numWaiting[1] = 0;
        value = [];
    }
    
    T allocate(int priority){
        mtx.wait();
        if (busy) {
            numWaiting[priority]++;
            mtx.notify();
            sems[priority].wait();
            mtx.wait();
            busy = true;
            mtx.notify();
            return value;
        } else {
            busy = true;
            mtx.notify();
            return value;
        }
    }
    
    void deallocate(T v){
        mtx.wait();
        busy = false;
        if (numWaiting[1] > 0) {
            numWaiting[1]--;
            busy = true;  // Gi ressursen til den ventende tråden
            sems[1].notify();
            mtx.notify();
        }
        else if (numWaiting[0] > 0) {
            numWaiting[0]--;
            busy = true;
            sems[0].notify();
            mtx.notify();
        }
        else {
            mtx.notify();
        }
        value = v;
    }
}





void main(){

    // Resource type er `int[]`. Hver bruker (resourceUser) legger til sin egen id til ressursen.
    auto resource = new Resource!(int[])();

    executionStates = new ExecutionState[](10);
    
    auto cfgs = [
        ResourceUserConfig(0, 0, 1, 1),
        ResourceUserConfig(1, 0, 3, 1),
        ResourceUserConfig(2, 1, 5, 1),
        
        ResourceUserConfig(0, 1, 10, 2),
        ResourceUserConfig(1, 0, 11, 1),
        ResourceUserConfig(2, 1, 11, 1),
        ResourceUserConfig(3, 0, 11, 1),
        ResourceUserConfig(4, 1, 11, 1),
        ResourceUserConfig(5, 0, 11, 1),
        ResourceUserConfig(6, 1, 11, 1),
        ResourceUserConfig(7, 0, 11, 1),
        ResourceUserConfig(8, 1, 11, 1),
        
        ResourceUserConfig(0, 1, 25, 3),
        ResourceUserConfig(6, 0, 26, 2),
        ResourceUserConfig(7, 0, 26, 2),
        ResourceUserConfig(1, 1, 26, 2),
        ResourceUserConfig(2, 1, 27, 2),
        ResourceUserConfig(3, 1, 28, 2),
        ResourceUserConfig(4, 1, 29, 2),
        ResourceUserConfig(5, 1, 30, 2),
    ];
    
    spawn(&executionLogger);
    foreach(cfg; cfgs){
        spawnLinked(&resourceUser, cfg, cast(shared)resource);
    }
    foreach(_; 0..cfgs.length){
        receive(
            (LinkTerminated lt){
            }
        );
    }
    Thread.sleep(tick*2);
    
    auto val = resource.allocate(0);
    
    assert(val.length == cfgs.length,
        "Test failed: Did not run all users once");
    assert(val[0..3] == [0, 1, 2],
        format("Test 1 failed: Did not run users in ascending order, instead ran %s", val[0..3]));
    
    assert(val[3] == 0,
        format("Test 2 failed: Did not run initial (high priority) user, instead ran %s", val[3]));
    assert(val[4..8].all!("(a & 1) == 0"),
        format("Test 2 failed: Did not run high priority (even id) users first, instead ran %s", val[4..8]));
    assert(val[8..12].all!("a & 1"),
        format("Test 2 failed: Did not run low priority (odd id) users last, instead ran %s", val[8..12]));
    
    assert(val[12] == 0,
        format("Test 3 failed: Did not run initial (high priority) user, instead ran %s", val[12]));
    assert(val[13..18].all!("a >= 1") && val[13..18].all!("a <= 5"),
        format("Test 3 failed: Did not run high priority users first, instead ran %s", val[13..18]));
    assert(val[18..20].all!("a >= 6") && val[18..20].all!("a <= 7"),
        format("Test 3 failed: Did not run low priority users last, instead ran %s", val[18..20]));
    writeln("All tests pass");
}



// --- RESOURCE USERS -- //

struct ResourceUserConfig {
    int     id;
    int     priority;
    int     release;
    int     execute;
}

void resourceUser(ResourceUserConfig cfg, shared Resource!(int[]) r){
    Thread.getThis.isDaemon = true;    
    auto resource = cast(Resource!(int[]))r;
    
    Thread.sleep(cfg.release * tick);
    
    executionStates[cfg.id] = ExecutionState.waiting;
    auto val = resource.allocate(cfg.priority);
    
    executionStates[cfg.id] = ExecutionState.executing;
    
    Thread.sleep(cfg.execute * tick);
    val ~= cfg.id;
    resource.deallocate(val);
    
    executionStates[cfg.id] = ExecutionState.done;
}



// --- EXECUTION LOGGING --- //

version(Windows){
    enum ExecutionState : char {
        none        = ' ',
        waiting     = cast(char)177,
        executing   = cast(char)178,
        done        = cast(char)223,
    }
    enum Grid : char {
        none        = ' ',
        horizontal  = cast(char)196,
    }
} else {
    enum ExecutionState : wchar {
        none        = ' ',
        waiting     = '\u2592',
        executing   = '\u2593',
        done        = '\u2580',
    }
    enum Grid : wchar {
        none        = ' ',
        horizontal  = '\u2500',
    }
}

__gshared ExecutionState[] executionStates;

void executionLogger(){
    Thread.getThis.isDaemon = true;
    Thread.sleep(tick/2);
    
    writefln("  id:%(%3d%)", iota(0, executionStates.length));
    
    auto t = 0;
    while(true){
        writef("%04d : " , t);
        foreach(id, ref state; executionStates){
            auto grid = (t % 5 == 0) ? Grid.horizontal : Grid.none;
            writef("%c%c%c", cast(OriginalType!ExecutionState)state, grid, grid);
            if(state == ExecutionState.done){
                state = ExecutionState.none;
            }
        }
        writeln;
        t++;
        Thread.sleep(tick);
    }
}