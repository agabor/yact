# Goal

Create a minimal Go program that prints "Hello, world!" to the console.

## Steps

### 1. Create the hello world program

**File:** [main.go](main.go) (new)

- Add `package main` and an import of `fmt`
- Add `func Greeting() string` that returns the constant text `Hello, world!`, so the message is a pure function that can be tested without console output (P5: functional core)
- Add `func main()` that calls `fmt.Println(Greeting())` and nothing else (P5: thin outer shell)

**Context:** none

## Assumptions

- [ ] **A1. Unclear:** the task does not say where the program lives or whether it needs a `go.mod`. Taken: a single `main.go` in the project root, runnable with `go run main.go`.