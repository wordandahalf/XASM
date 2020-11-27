package xasm

import (
    "bufio"
    "log"
    "os"
    "regexp"
    "strings"
)

type xasmFile struct {
    path        string
    lines       []xasmLine
}

type xasmLine struct {
    number      int                 // The line number
    content     string              // The raw content of the line

    instruction xasmInstruction     // The instruction parsed from this line
}

type xasmInstruction struct {
    name        string
    operands    []string
}

func check(e error) {
    if e != nil {
        panic(e)
    }
}

func Load(path string) *xasmFile {
    file := xasmFile{path, nil}

    f, err := os.Open(file.path)
    check(err)
    defer f.Close()

    var lines []xasmLine
    var index = 0

    scanner := bufio.NewScanner(f)

    for scanner.Scan() {
        lines = append(lines, xasmLine{index, scanner.Text(), xasmInstruction {"", nil}})
        index++
    }

    file.lines = lines

    return &file
}

func (file xasmFile) GetLength() int {
    return len(file.lines)
}

func (file xasmFile) GetParsedInstructions() []xasmInstruction {
    instructions := make([]xasmInstruction, file.GetLength())

    for i, line := range file.lines {
        instructions[i] = line.instruction
    }

    return instructions
}

func (file *xasmFile) Parse() {
    for i, line := range file.lines {
        file.lines[i].instruction = parseLine(line)
    }
}

const (
    commentDelimiter        = ";"
    labelSuffix             = ":"
)

var (
    commentPattern, _       = regexp.Compile(commentDelimiter + "\\s*(.*)$")
    spacePattern, _         = regexp.Compile("[[:blank:]]{2,}")

    labelPattern, _         = regexp.Compile("^([a-zA-Z_]+)" + labelSuffix + "$")
    instructionPattern, _   = regexp.Compile("^([a-zA-Z]+)(?: |$)")
    opcodePattern, _        = regexp.Compile("^\\s*\\w+\\s*$")
)

func parseLine(line xasmLine) xasmInstruction {
    text := line.content
    instruction := xasmInstruction{"", []string{}}

    if text == "" {
        return instruction
    }

    // Extract any comments
    if strings.Contains(text, commentDelimiter) {
        text = commentPattern.ReplaceAllString(text, "")
    }

    // Replace every tab with a space
    text = strings.ReplaceAll(text, "\t", " ")

    // Replace multiple spaces with single space
    text = spacePattern.ReplaceAllString(text, " ")

    // Remove leading and trailing whitespace
    text = strings.TrimSpace(text)

    // If the line is empty, return the line
    if text == "" {
        return instruction
    }

    /*
       Parse line
    */

    if strings.Contains(text, labelSuffix) {
        labelMatches := labelPattern.FindStringSubmatch(text)

        if len(labelMatches) == 2 {
            instruction = xasmInstruction{"LABEL", []string {labelMatches[1]}}
        } else {
            index := strings.Index(line.content, strings.Split(strings.TrimSpace(text), " ")[0])

            log.Fatalf( "Syntax error on line %d: malformed label\n%s\n%s^\n", line.number, line.content, strings.Repeat(" ", index))
        }
    } else
    if instructionPattern.MatchString(text) {
        instructionMatches := instructionPattern.FindStringSubmatch(text)

        text := strings.Replace(text, instructionMatches[0], "", 1)

        instruction = xasmInstruction{instructionMatches[1], []string {}}

        if text != "" {
            opcodes := strings.Split(text, ",")

            for _, opcode := range opcodes {
                if opcodePattern.MatchString(opcode) {
                    instruction.operands = append(instruction.operands, strings.TrimSpace(opcode))
                } else {
                    // This accounts for malformed operands with multiple spaces in between.
                    // Since the parser replaces multiple spaces with single spaces,
                    // a simple strings.Index call with the provided operand could result in
                    // -1 being returned.
                    index := strings.Index(line.content, strings.Split(strings.TrimSpace(opcode), " ")[0])

                    log.Fatalf( "Syntax error on line %d: malformed operand\n%s\n%s^\n", line.number, line.content, strings.Repeat(" ", index))
                }
            }
        }
    }

    return instruction
}