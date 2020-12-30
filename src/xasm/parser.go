package xasm

import (
    "bufio"
    "log"
    "os"
    "regexp"
    "strconv"
    "strings"
)

func check(e error) {
    if e != nil {
        panic(e)
    }
}

func Load(path string) *xasmFile {
    file := xasmFile{path, nil, []xasmInstruction {}, make(map[string]byte)}

    f, err := os.Open(file.path)
    check(err)
    defer f.Close()

    var lines []xasmLine
    var index = 0

    scanner := bufio.NewScanner(f)

    for scanner.Scan() {
        lines = append(lines, xasmLine{index, scanner.Text()})
        index++
    }

    file.lines = lines

    return &file
}

func (file *xasmFile) Parse() {
    previousInstruction := xasmInstruction{0, 0, 0, "", []xasmOperand {}}
    for _, line := range file.lines {
        parseLine(file, previousInstruction, line)

        length := len(file.instructions)
        if length > 0 {
            previousInstruction = file.instructions[length- 1]
        }
    }
}

const (
    commentDelimiter        = ";"
    labelSuffix             = ":"
)

var (
    // Pre-processing regexes
    commentPattern, _           = regexp.Compile(commentDelimiter + "\\s*(.*)$")
    spacePattern, _             = regexp.Compile("[[:blank:]]{2,}")

    // Regexes for line parsing
    labelPattern, _             = regexp.Compile("^([a-zA-Z_]+)" + labelSuffix + "$")
    instructionPattern, _       = regexp.Compile("^([a-zA-Z]+)(?: |$)")

    // Operand regexes
    registerPattern, _          = regexp.Compile("^r[0-7]$")
    registerPointerPattern, _   = regexp.Compile("^\\[r[0-7]]$")
    numberPattern, _            = regexp.Compile("[0-9]+")// regexp.Compile("^(0x[[:xdigit:]]{1,2}|0[0-7]{0,3})$")
    labelOperandPattern, _      = regexp.Compile("^([a-zA-Z_]+)$")

    // Special parsers for instructions that require further
    // processing of the mnemonic and operands.
    instructionParsers      = map[string]func(int, int, string, []xasmOperand) xasmInstruction {
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

    flagValues              = map[string]byte {
        "P":    0b00,
        "Z":    0b01,
        "C":    0b10,
    }
)

func sanitizeLine(text string) string {
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

    return text
}

// Parses the value of an xasmLine into an xasmInstruction
func parseLine(file *xasmFile, previous xasmInstruction, line xasmLine) {
    text := line.content

    if text == "" {
        return
    }

    /*
        Perform pre-processing actions...
     */
    text = sanitizeLine(text)

    // If the line is empty, return the line
    if text == "" {
        return
    }

    /*
       Parse line...
    */

    offset := previous.offset + previous.length

    if strings.Contains(text, labelSuffix) {
        labelMatches := labelPattern.FindStringSubmatch(text)

        if len(labelMatches) == 2 {
            file.symbols[labelMatches[1]] = (byte) (offset & 0xFF)
            return
        } else {
            index := strings.Index(line.content, strings.Split(strings.TrimSpace(text), " ")[0])

            log.Fatalf( "Syntax error on line %d: malformed label\n%s\n%s^\n", line.number, line.content, strings.Repeat(" ", index))
        }
    } else
    if instructionPattern.MatchString(text) {
        instructionMatches := instructionPattern.FindStringSubmatch(text)

        text := strings.Replace(text, instructionMatches[0], "", 1)

        mnemonic := instructionMatches[1]
        var operands []xasmOperand

        // If there is any text left it must be an operand
        if text != "" {
            // Split them by commas
            rawOperands := strings.Split(text, ",")

            for _, operand := range rawOperands {
                parsedOperand := parseOperand(strings.TrimSpace(operand))

                if parsedOperand.operandType != invalidOperand {
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

        parser, found := instructionParsers[mnemonic]

        if found {
            file.instructions = append(file.instructions, parser(line.number, offset, mnemonic, operands))
        } else {
            file.instructions = append(file.instructions, xasmInstruction{line.number, offset, 1 + getOperandsLength(operands), mnemonic, operands})
        }
    }
}

// Parses an operand into an xasmOperand struct
func parseOperand(operand string) xasmOperand {
    if registerPattern.MatchString(operand) {
        return xasmOperand{registerOperand, operand}
    } else
    if registerPointerPattern.MatchString(operand) {
        return xasmOperand{registerPointerOperand, operand}
    } else
    if numberPattern.MatchString(operand) {
        val, e := strconv.ParseInt(operand, 0, 8)

        if e == nil {
            return xasmOperand{immediateOperand, (byte) (val & 0xFF)}
        }
    } else
    if labelOperandPattern.MatchString(operand) {
        return xasmOperand{labelOperand, operand}
    }

    // TODO: error message
    return xasmOperand{invalidOperand, operand}
}

func getOperandsLength(operands []xasmOperand) int {
    length := 0
    for _, operand := range operands {
        if operand.operandType == immediateOperand {
            length++
        }
    }

    return length
}

// Parses a jump instruction by prepending the second character of the mnemonic as a flag operand.
func parseJump(offset int, line int, mnemonic string, operands []xasmOperand) xasmInstruction {
    flagValue, found := flagValues[strings.TrimPrefix(mnemonic, "J")]

    if found {
        return xasmInstruction{ line, offset, 2, "J", append([]xasmOperand { {flagOperand, flagValue } }, operands...)}
    } else {
        // TODO: error message
        return xasmInstruction{line, offset, 0, invalidOperand, operands}
    }
}

// Parses an ALU instruction by prepending the ALU opcode of the instruction.
func parseAlu(offset int, line int, mnemonic string, operands []xasmOperand) xasmInstruction {
    return xasmInstruction{line, offset, 1, "ALU", append([]xasmOperand { {aluOpcodeOperand, aluOpcodes[mnemonic] } }, operands...)}
}