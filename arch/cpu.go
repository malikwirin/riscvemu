package arch

import (
	"fmt"
	"github.com/malikwirin/riscvemu/assembler"
)

type CPU struct {
	Reg [32]int32
	PC  uint32
}

const INSTRUCTION_SIZE = 4

func NewCPU() *CPU {
	return &CPU{
		Reg: [32]int32{},
		PC:  0,
	}
}

func (c *CPU) SetReg(idx RegIndex, value int32) {
	if idx != 0 {
		c.Reg[idx] = value
	}
}

// execOpcode decodes and executes the instruction's opcode.
func (c *CPU) exec(instr assembler.Instruction) error {
	opcode := instr.Opcode()

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

	err = c.exec(assembler.Instruction(instr))
	if err != nil {
		return err
	}

	c.PC += INSTRUCTION_SIZE

	return nil
}
