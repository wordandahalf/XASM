package xasm

type xasmFile struct {
    path            string
    lines           []xasmLine
    instructions    []xasmInstruction

    symbols         map[string]byte
}

func (file xasmFile) GetLength() int {
    return len(file.lines)
}

type xasmLine struct {
    number      int                 // The line number
    content     string              // The raw content of the line
}

type xasmInstruction struct {
    line        int
    offset      int
    length      int

    name        string
    operands    []xasmOperand
}

func (instruction xasmInstruction) GetOperandTypes() [3]string {
    var operandTypes [3]string

    for i, operand := range instruction.operands {
        operandTypes[i] = operand.operandType
    }

    return operandTypes
}

const (
    // Operand types
    registerOperand         = "REGISTER"
    registerPointerOperand  = "REGISTER_POINTER"
    immediateOperand        = "IMMEDIATE"
    labelOperand            = "STRING"
    flagOperand             = "FLAG"
    aluOpcodeOperand        = "ALU_OPCODE"
    invalidOperand          = ""
)

type xasmOperand struct {
    operandType string
    value       interface{}
}
