# go-internals-lab
**Go version:** 1.27.3 (output of `go version`)

**Note:** results may vary depending on the Go version.

## About this repository
- The experiments are based on a personal study plan and a lab manual that I created with the help of AI.
- The numbers in the folder names are lab numbers, which roughly follow the days of the plan.
- I wrote this README myself. AI was used only to proofread my English.
- Feel free to use this repository as study material.

## Experiments

### 01-escape: stack vs. heap (escape analysis) 

**Goal**: see how the compiler decides between the stack and the heap.

**Execution**: run `go build -gcflags="-m -l" ./01-escape`
This prints which variables escape to the heap. The `-l` flag disables inlining, which could otherwise hide some escapes.


**Results**:

The compiler decides at compile time (escape analysis) whether a variable stays on the stack or escapes to the heap. Freeing stack memory is almost free: when the function returns, the stack pointer simply moves back. Freeing heap memory is more expensive, because it is done later by the garbage collector (see 02-gc).

Keeping this in mind is important for performance: fewer heap allocations mean less work for the garbage collector and faster programs, especially in hot code paths.

This experiment covered five cases:

| case № | What escaped/moved to heap                       | Why                                                                                                                      |
|--------|--------------------------------------------------|--------------------------------------------------------------------------------------------------------------------------|
| 1      | pointer (u) moved                                | Its address is returned, so it must outlive the function                                                                 |
| 2      | value (n) escaped                                | `fmt.Println` takes `...any`, so the value is boxed into an interface, and the compiler cannot prove it will not be kept |
| 3      | variable (c) moved to heap, func literal escaped | The returned closure captures `c`, so `c` must live as long as the closure                                               |
| 4      | channel (ch) didn't escape, variable (u) moved   | A pointer sent into a channel can be used later by another goroutine                                                     |
| 5      | slice (s) escaped to heap                        | The slice is too big for the compiler's stack limit                                                                      |

These are not the only cases, but they are the most common ones.

Results are based on the output:
```terminaloutput
# go-internals-lab/01-escape
01-escape/main.go:9:2: moved to heap: u
01-escape/main.go:15:13: ... argument does not escape
01-escape/main.go:15:14: n escapes to heap
01-escape/main.go:20:2: moved to heap: c
01-escape/main.go:21:9: func literal escapes to heap
01-escape/main.go:25:11: ch does not escape
01-escape/main.go:26:2: moved to heap: u
01-escape/main.go:32:11: make([]int, 100000) escapes to heap
```

### 02-gc: The garbage collector 

**Goal**: see how GOGC and GOMEMLIMIT change how often the GC runs.

**Execution**: run `$env:GOGC="xx"; go run ./02-gc` - where xx is a value from [50, 100, 400, off].

Then also run `$env:GOGC="off"; $env:GOMEMLIMIT="30MiB"; go run ./02-gc` and `$env:GODEBUG="gctrace=1"; go run ./02-gc`

**Results**:

There are two common ways to affect how often GC runs (both involve environment variables):
- GC - you specify how much a live heap can grow in percentage comparing to the previous live heap. Works a bit more strictly than GOMEMLIMIT
- GOMEMLIMIT - you specify "the soft ceiling" for a live heap and when it grows up to around that memory amount - GC runs.

In the experiment, the total amount of allocated memory is 157 MB. 
In the table, you can see how every variable value affects how frequently GC is run:

| env                 | time (ms) | number of GC cycles | heap allocations (MB)   | total time of pauses |
|---------------------|-----------|---------------------|-------------------------|----------------------|
| GOGC=50             | 388.7821  | 135                 | 3                       | 8.243ms              |
| GOGC=100            | 207.0329  | 59                  | 6                       | 5.1259ms             |
| GOGC=400            | 177.7891  | 11                  | 11                      | 989.4µs              |
| GOGC=off            | 194.4756  | 0                   | 157 (as in total alloc) | 0s                   |
| GOMEMLIMIT="30MiB"  | 2.0456946 | 789                 | 3                       | 56.9069ms            |

GC has 2 concurrent and 2 stop-the-world (STW) phases. They go roughly in this order: Sweep Termination -> Mark -> Mark Termination -> Sweep.
Termination phases are STW (but they take less than 1ms in general), whereas the others are concurrent.
In the output of `$env:GODEBUG="gctrace=1"; go run ./02-g` commands we can see how much exactly every phase took:
```
gc 1 @0.066s 0%: 0+2.6+0 ms clock, 0+0/2.7/0+0 ms cpu, 3->4->1 MB, 4 MB goal, 0 MB stacks, 0 MB globals, 16 P
gc 2 @0.075s 0%: 0+3.2+0 ms clock, 0+1.5/7.3/0+0 ms cpu, 3->4->1 MB, 4 MB goal, 0 MB stacks, 0 MB globals, 16 P
gc 3 @0.084s 2%: 0+5.2+0.55 ms clock, 0+1.5/9.5/0.58+8.8 ms cpu, 3->4->1 MB, 4 MB goal, 0 MB stacks, 0 MB globals, 16 P
gc 4 @0.099s 2%: 0.54+2.6+0 ms clock, 8.7+0/5.9/3.7+0 ms cpu, 3->5->2 MB, 4 MB goal, 0 MB stacks, 0 MB globals, 16 P
gc 5 @0.105s 2%: 0+1.7+0.57 ms clock, 0+0.56/2.8/0.56+9.2 ms cpu, 4->4->1 MB, 5 MB goal, 0 MB stacks, 0 MB globals, 16 P
gc 6 @0.139s 1%: 0.50+1.0+0 ms clock, 8.0+0/2.0/0+0 ms cpu, 3->4->2 MB, 4 MB goal, 0 MB stacks, 0 MB globals, 16 P
gc 7 @0.143s 2%: 0+1.8+0 ms clock, 0+1.0/4.6/1.6+0 ms cpu, 4->5->2 MB, 4 MB goal, 0 MB stacks, 0 MB globals, 16 P
...
```
Looking at `0+5.2+0.55 ms clock` part of the output we see that Sweep Termination took 0s, Marking took 5.2ms and Mark Termination took 0.55ms.
You also can tell how well your program is doing by looking at what the memory levels were before and after GC in this part: `3->4->1 MB`.
First number is heap size before GC, second - after GC and third - amount of live data. You also can spot a leak here, if the third number is constantly growing!

The trade-off of GC is how fast your code runs. Keeping memory low costs a lot of CPU power, that could be used for running your code.
So I would mainly tweak GOGC number if GC runs too frequently. But if I get `OOM killed` message - then I would adjust GOMEMLIMIT.
