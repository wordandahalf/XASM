package xasm

import (
    "bufio"
    "log"
    "os"
    "regexp"
    "strconv"
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
    operands    []xasmOperand
}

type xasmOperand struct {
    operandType string
    value       interface{}
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
    commentPattern, _           = regexp.Compile(commentDelimiter + "\\s*(.*)$")
    spacePattern, _             = regexp.Compile("[[:blank:]]{2,}")

    labelPattern, _             = regexp.Compile("^([a-zA-Z_]+)" + labelSuffix + "$")
    instructionPattern, _       = regexp.Compile("^([a-zA-Z]+)(?: |$)")
    operandPattern, _           = regexp.Compile("^\\s*\\w+\\s*$")

    registerPattern, _          = regexp.Compile("^[r][0-7]$")
    registerPointerPattern, _   = regexp.Compile("^\\[[r][0-7]]$")
    numberPattern, _            = regexp.Compile("[0-9]+")// regexp.Compile("^(0x[[:xdigit:]]{1,2}|0[0-7]{0,3})$")
    labelOperandPattern, _      = regexp.Compile("^([a-zA-Z_]+)$")

    // Special parsers for instructions that require further
    // processing of the mnemonic and operands.
    instructionParsers      = map[string]func(string, []xasmOperand) xasmInstruction {
        "JP": parseJump,
        "JZ": parseJump,
        "JC": parseJump,

        "NOT":  parseAlu,
        "AND":  parseAlu,
        "OR":   parseAlu,
        "XOR":  parseAlu,
        "SHL":  parseAlu,
        "SHR":  parseAlu,
        "ADD":  parseAlu,
        "SUB":  parseAlu,
    }

    // Maps mnemonics to ALU opcodes.
    aluOpcodes              = map[string]byte {
        "NOT":  0b000,
        "AND":  0b001,
        "OR":   0b010,
        "XOR":  0b011,
        "SHL":  0b100,
        "SHR":  0b101,
        "ADD":  0b110,
        "SUB":  0b111,
    }
)

func parseLine(line xasmLine) xasmInstruction {
    text := line.content
    instruction := xasmInstruction{"", []xasmOperand{}}

    if text == "" {
        return instruction
    }

    /*
        Perform pre-processing actions...
     */

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
       Parse line...
    */

    if strings.Contains(text, labelSuffix) {
        labelMatches := labelPattern.FindStringSubmatch(text)

        if len(labelMatches) == 2 {
            instruction = xasmInstruction{"LABEL", []xasmOperand { {"STRING", labelMatches[1] } }}
        } else {
            index := strings.Index(line.content, strings.Split(strings.TrimSpace(text), " ")[0])

            log.Fatalf( "Syntax error on line %d: malformed label\n%s\n%s^\n", line.number, line.content, strings.Repeat(" ", index))
        }
    } else
    if instructionPattern.MatchString(text) {
        instructionMatches := instructionPattern.FindStringSubmatch(text)

        text := strings.Replace(text, instructionMatches[0], "", 1)

        mnemonic := instructionMatches[1]
        operands := make([]xasmOperand, 0)

        if text != "" {
            rawOperands := strings.Split(text, ",")

            for _, operand := range rawOperands {
                parsedOperand := parseOperand(strings.TrimSpace(operand))

                if parsedOperand.operandType != "" {
                    operands = append(operands, parsedOperand)
                } else {
                    // This accounts for malformed operands with multiple spaces in between.
                    // Since the parser replaces multiple spaces with single spaces,
                    // a simple strings.Index call with the provided operand could result in
                    // -1 being returned.
                    index := strings.Index(line.content, strings.Split(strings.TrimSpace(operand), " ")[0])

                    log.Fatalf( "Syntax error on line %d: malformed operand\n%s\n%s^\n", line.number, line.content, strings.Repeat(" ", index))
                }
            }
        }

        parser, e := instructionParsers[mnemonic]

        if e {
            instruction = parser(mnemonic, operands)
        } else {
            instruction = xasmInstruction{mnemonic, operands}
        }
    }

    return instruction
}

func parseOperand(operand string) xasmOperand {
    if registerPattern.MatchString(operand) {
        return xasmOperand{"REGISTER", operand}
    } else
    if registerPointerPattern.MatchString(operand) {
        return xasmOperand{"REGISTER_POINTER", operand}
    } else
    if numberPattern.MatchString(operand) {
        val, e := strconv.ParseInt(operand, 0, 8)

        if e == nil {
            return xasmOperand{"IMMEDIATE", val}
        }
    } else
    if labelOperandPattern.MatchString(operand) {
        return xasmOperand{"STRING", operand}
    }

    return xasmOperand{"", operand}
}

// Parses a jump instruction by prepending the second character of the mnemonic as a flag operand.
func parseJump(mnemonic string, operands []xasmOperand) xasmInstruction {
    return xasmInstruction{ "J", append([]xasmOperand { { "FLAG", strings.TrimPrefix(mnemonic, "J") } }, operands...)}
}

// Parses an ALU instruction by prepending the ALU opcode of the instruction.
func parseAlu(mnemonic string, operands []xasmOperand) xasmInstruction {
    return xasmInstruction{"ALU", append([]xasmOperand { { "ALU_OPCODE", aluOpcodes[mnemonic] } }, operands...)}
}