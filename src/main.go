package main

import (
    "./xasm"
    "flag"
    "strings"
)

func main() {
    var inputFile string
    var outputFile string

    flag.StringVar(&inputFile, "i", "", "Input file")
    flag.StringVar(&outputFile, "o", "", "Output file")
    flag.Parse()

    if outputFile == "" {
        outputFile = inputFile[0:strings.LastIndex(inputFile, ".")] + ".bin"
    }

    var file = xasm.Load(inputFile)
    file.Parse()
    file.Assemble(outputFile)
}