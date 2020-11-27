package main

import (
    "./xasm"
    "fmt"
    "os"
)

func main() {
    var args = os.Args[1:]

    fmt.Printf("Args: %v\n", args)

    var file = xasm.Load(args[0])
    file.Parse()

    for i, instruction := range file.GetParsedInstructions() {
        fmt.Printf("Line %d:\n", i)
        fmt.Printf("Instruction:\t%v\n", instruction)
        fmt.Println()
    }
}