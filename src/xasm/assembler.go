package xasm

import (
    "log"
    "os"
    "strconv"
    "strings"
)

var (
    instructionEncoders = map[string]map[[3]string]func(*xasmFile, xasmInstruction) []byte {
        "NOP": {
            {}: encodeConstant(0x00),
        },
        "LD": {
            { registerOperand, registerOperand }: encodeLoadRegisterWithRegister,
            { registerOperand, immediateOperand }: encodeLoadRegisterWithImmediate,
            { registerPointerOperand, registerOperand }: encodeLoadRegisterPointerWithRegister,
            { registerOperand, registerPointerOperand }: encodeLoadRegisterWithRegisterPointer,
        },
        "DSPLY": {
            { registerOperand }: encodeDisplayRegister,
        },
        "ALU": {
            { aluOpcodeOperand, registerOperand, registerOperand }: encodeAluOperation,
            { aluOpcodeOperand, registerOperand }: encodeAluOperation,
        },
        "J": {
            { flagOperand, labelOperand }: encodeJump,
        },
        "HLT": {
            {}: encodeConstant(0xFF),
        },
    }
)

func (file *xasmFile) Assemble(path string) {
    f, err := os.Create(path)
    check(err)
    defer f.Close()

    for _, instruction := range file.instructions {
        encoder := instructionEncoders[instruction.name][instruction.GetOperandTypes()]

        if encoder != nil {
            _, _ = f.Write(
                encoder(file, instruction),
            )
        } else {
            log.Fatalf( "Assembly error on line %d: unknown instruction '%s' with operands %v.\n", instruction.line, instruction.name, instruction.operands)
        }
    }
}

func encodeRegister(instruction xasmInstruction, register string) byte {
    val, e := strconv.ParseInt(strings.TrimPrefix(register, "r"), 10, 3)

    if e == nil {
        return (byte) (val & 0xFF)
    } else {
        log.Fatalf("Assembly error on line %d: invalid register '%s'.\n", instruction.line, register)
    }

    return 0
}

func encodeConstant(value byte) func(*xasmFile, xasmInstruction) []byte {
    return func(*xasmFile, xasmInstruction) []byte {
        return []byte { value }
    }
}

func encodeLoadRegisterWithRegister(_ *xasmFile, instruction xasmInstruction) []byte {
    operands := instruction.operands

    dst := encodeRegister(instruction, operands[0].value.(string))
    src := encodeRegister(instruction, operands[1].value.(string))

    return []byte { dst << 3 | src }
}

func encodeLoadRegisterPointerWithRegister(_ *xasmFile, instruction xasmInstruction) []byte {
    operands := instruction.operands

    dst := encodeRegister(instruction, operands[0].value.(string)[1:3])
    src := encodeRegister(instruction, operands[1].value.(string))

    if dst != 0 {
        log.Fatalf("Assembly error on line %d: invalid destination register pointer '%s'. Must be [r0].\n", instruction.line, operands[0].value.(string))
    }

    return []byte { 0x40 | src }
}

func encodeLoadRegisterWithRegisterPointer(_ *xasmFile, instruction xasmInstruction) [] byte {
    operands := instruction.operands

    dst := encodeRegister(instruction, operands[0].value.(string))
    src := encodeRegister(instruction, operands[1].value.(string)[1:3])

    if dst != 0 {
        log.Fatalf("Assembly error on line %d: invalid destination register '%s'. Must be r0.\n", instruction.line, operands[0].value.(string))
    }

    return []byte { 0x48 | src }
}

func encodeLoadRegisterWithImmediate(_ *xasmFile, instruction xasmInstruction) []byte {
    operands := instruction.operands

    src := encodeRegister(instruction, operands[0].value.(string))
    immediate := operands[1].value.(byte)

    if src != 0 {
        log.Fatalf("Assembly error on line %d: invalid immediate register destination '%s'. Must be r0.\n", instruction.line, operands[0].value.(string))
    }

    return []byte { 0x40, immediate }
}

func encodeDisplayRegister(_ *xasmFile, instruction xasmInstruction) []byte {
    operands := instruction.operands

    return []byte { 0x78 | encodeRegister(instruction, operands[0].value.(string)) }
}

func encodeAluOperation(_ *xasmFile, instruction xasmInstruction) []byte {
    operands := instruction.operands

    var src byte

    if len(operands) == 3 {
        src = encodeRegister(instruction, operands[2].value.(string))

        if encodeRegister(instruction, operands[1].value.(string)) != 0 {
            log.Fatalf("Assembly error on line %d: invalid ALU destination register '%s'. Must be r0.\n", instruction.line, operands[1].value.(string))
        }
    } else {
        src = encodeRegister(instruction, operands[1].value.(string))
    }

    return []byte { 0x80 | (operands[0].value.(byte) << 3) | src }
}

func encodeJump(file *xasmFile, instruction xasmInstruction) []byte {
    operands := instruction.operands
    label := operands[1].value.(string)
    labelOffset, found := file.symbols[label]

    if !found {
        log.Fatalf("Assembly error on line %d: undefined symbol '%s'.\n", instruction.line, label)
    }

    return []byte { 0xC0 | operands[0].value.(byte), labelOffset }
}