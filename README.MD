# go-internals-lab
**Go version:** 1.27.3 (output of `go version`)

**Note:** results may vary depending on the Go version.

## About this repository
- The experiments are based on a personal study plan and a lab manual that I created with the help of AI.
- The numbers in the folder names are lab numbers, which roughly follow the days of the plan.
- I wrote this README myself. AI was used only to proofread my English.
- Feel free to use this repository as study material.
- 
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
