package arch

import (
	"fmt"
)

type CPU struct {
    Registers [32]int32
    PC        uint32
}

const INSTRUCTION_SIZE = 4

func NewCPU() *CPU {
    return &CPU{
        Registers: [32]int32{},
        PC:        0,
    }
}

func (c *CPU) SetRegister(idx RegIndex, value int32) {
    if idx != 0 {
        c.Registers[idx] = value
    }
}

// execOpcode decodes and executes the instruction's opcode.
func (c *CPU) execOpcode(instr uint32) error {
    // Extract the opcode from the instruction by masking the lowest 7 bits.
	opcode := instr & 0x7F

    switch opcode {
    case 0x00: // NOP
        return nil
    default:
        return fmt.Errorf("unknown opcode: 0x%X", opcode)
    }
}

func (c *CPU) Step(memory WordReader) error {
    instr, err := memory.ReadWord(c.PC)
    if err != nil {
        return err
    }

	err = c.execOpcode(instr)
	if err != nil {
		return err
	}

    c.PC += INSTRUCTION_SIZE

    return nil
}
