//go:build !desktop

package main

import "fmt"

func main() { fmt.Println("bpm desktop app requires building with -tags desktop (see README)") }
