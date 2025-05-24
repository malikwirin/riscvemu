package assembler

import (
	"fmt"
)

// Opcode represents the main opcode field of an instruction.
type Opcode uint32

const (
    // R-Type (e.g., add, sub, and, or, xor)
    OPCODE_R_TYPE Opcode = 0x33

    // I-Type (addi, andi, ori, jalr, lw)
    OPCODE_I_TYPE   Opcode = 0x13
    OPCODE_JALR     Opcode = 0x67
    OPCODE_LOAD     Opcode = 0x03

    // S-Type (sw)
    OPCODE_STORE    Opcode = 0x23

    // B-Type (beq, bne)
    OPCODE_BRANCH   Opcode = 0x63

    // J-Type (jal)
    OPCODE_JAL      Opcode = 0x6F
)

// Funct3 field values
const (
    FUNCT3_ADD_SUB uint32 = 0x0
    FUNCT3_AND     uint32 = 0x7
    FUNCT3_OR      uint32 = 0x6
    FUNCT3_XOR     uint32 = 0x4

    FUNCT3_BEQ     uint32 = 0x0
    FUNCT3_BNE     uint32 = 0x1

    FUNCT3_LW      uint32 = 0x2
    FUNCT3_SW      uint32 = 0x2

    FUNCT3_ADDI    uint32 = 0x0
    FUNCT3_ANDI    uint32 = 0x7
    FUNCT3_ORI     uint32 = 0x6

    FUNCT3_JALR    uint32 = 0x0
)

// Funct7 field values (only relevant for add/sub)
const (
    FUNCT7_ADD uint32 = 0x00
    FUNCT7_SUB uint32 = 0x20
)

func (op Opcode) String() string {
    switch op {
    case OPCODE_R_TYPE:
        return "R-Type"
    case OPCODE_I_TYPE:
        return "I-Type"
    case OPCODE_JALR:
        return "JALR"
    case OPCODE_LOAD:
        return "LOAD"
    case OPCODE_STORE:
        return "STORE"
    case OPCODE_BRANCH:
        return "BRANCH"
    case OPCODE_JAL:
        return "JAL"
    default:
        return fmt.Sprintf("Unknown(0x%X)", uint32(op))
    }
}
