# Goal

Add a pure function that computes the 0-based Fibonacci number at a given index, and print the
26th one (121393) on a line after the "Hello, world!" greeting.

## Steps

### 1. Add the Fibonacci function and print its result after the greeting

**File:** [main.go](main.go)

- Add `const fibonacciIndex = 26`, with a comment that the index is 0-based, so F(0) = 0 and F(1) = 1
- Add `func Fibonacci(index int) int`, a pure function that computes the number iteratively
  with two running values and returns 0 for index 0 (P5: pure core, no I/O; P13: simple loop,
  no recursion, which would be exponential)
- In `func main()`, keep the call `fmt.Println(Greeting())` first, then add
  `fmt.Println(Fibonacci(fibonacciIndex))` so the number is printed on the following line

**Context:** none