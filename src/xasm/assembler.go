package xasm

import (
    "log"
    "os"
    "strconv"
    "strings"
)

var (
    instructionEncoders = map[string]map[[3]string]func([]xasmOperand) []byte {
        "NOP": {
            {}: encodeConstant(0x00),
        },
        "LD": {
            { registerOperand, registerOperand }: encodeLoadRegisterWithRegister,
            { registerPointerOperand, registerOperand }: encodeLoadRegisterPointerWithRegister,
            { registerOperand, immediateOperand }: encodeLoadRegisterWithImmediate,
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

    for _, line := range file.lines {
        instruction := line.instruction

        if instruction.name == "" || instruction.name == "LABEL" {
            continue
        }

        encoder := instructionEncoders[instruction.name][instruction.GetOperandTypes()]

        if encoder != nil {
            _, _ = f.Write(
                encoder(instruction.operands),
            )
        } else {
            log.Fatalf( "Assembly error on line %d: unknown instruction '%s' with operands %v.\n", line.number, instruction.name, instruction.operands)
        }
    }
}

func encodeRegister(register string) byte {
    val, e := strconv.ParseInt(strings.TrimPrefix(register, "r"), 10, 3)

    if e == nil {
        return (byte) (val & 0xFF)
    } else {
        log.Fatalf("Assembly error: invalid register '%s'.\n", register)
    }

    return 0
}

func encodeConstant(value byte) func([]xasmOperand) []byte {
    return func([]xasmOperand) []byte {
        return []byte { value }
    }
}

func encodeLoadRegisterWithRegister(operands []xasmOperand) []byte {
    dst := encodeRegister(operands[0].value.(string))
    src := encodeRegister(operands[1].value.(string))

    return []byte { dst << 3 | src }
}

func encodeLoadRegisterPointerWithRegister(operands []xasmOperand) []byte {
    dst := encodeRegister(operands[0].value.(string)[1:3])
    src := encodeRegister(operands[1].value.(string))

    if dst != 0 {
        log.Fatalf("Assembly error: invalid destination register pointer '%s'. Must be [r0].\n", operands[0].value.(string))
    }

    return []byte { 0x40 | src }
}

func encodeLoadRegisterWithImmediate(operands []xasmOperand) []byte {
    src := encodeRegister(operands[0].value.(string))
    immediate := operands[1].value.(byte)

    if src != 0 {
        log.Fatalf("Assembly error: invalid immediate register destination '%s'. Must be r0.\n", operands[0].value.(string))
    }

    return []byte { 0x40, immediate }
}

func encodeDisplayRegister(operands []xasmOperand) []byte {
    return []byte { 0x78 | encodeRegister(operands[0].value.(string)) }
}

func encodeAluOperation(operands []xasmOperand) []byte {
    var src byte

    if len(operands) == 3 {
        src = encodeRegister(operands[2].value.(string))

        if encodeRegister(operands[1].value.(string)) != 0 {
            log.Fatalf("Assembly error: invalid ALU destination register '%s'. Must be r0.\n", operands[1].value.(string))
        }
    } else {
        src = encodeRegister(operands[1].value.(string))
    }

    return []byte { 0x80 | (operands[0].value.(byte) << 3) | src }
}

func encodeJump(operands []xasmOperand) []byte {
    return []byte { 0xC0 | operands[0].value.(byte) }
}