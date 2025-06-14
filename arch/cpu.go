package arch

import (
	"fmt"
	"reflect"
	"github.com/malikwirin/riscvemu/assembler"
)

type CPU struct {
	Reg [32]uint32
	PC  uint32
}

const INSTRUCTION_SIZE = assembler.INSTRUCTION_SIZE

func NewCPU() *CPU {
	return &CPU{
		Reg: [32]uint32{},
		PC:  0,
	}
}

func (c *CPU) SetReg(idx RegIndex, value uint32) {
	if idx != 0 {
		c.Reg[idx] = value
	}
}

func (c *CPU) exec(instr assembler.Instruction) error {
	opcode := instr.Opcode()
	if opcode == assembler.OPCODE_INVALID {
		return fmt.Errorf("invalid opcode: 0x%X (from instruction 0x%X)", opcode, uint32(instr))
	}
	switch opcode {
	case assembler.OPCODE_R_TYPE:
		switch instr.Funct3() {
		case assembler.FUNCT3_ADD_SUB:
			rd := instr.Rd()
			rs1 := instr.Rs1()
			rs2 := instr.Rs2()
			if instr.Funct7() == assembler.FUNCT7_ADD {
				if rd != 0 {
					c.Reg[rd] = c.Reg[rs1] + c.Reg[rs2]
				}
			} else if instr.Funct7() == assembler.FUNCT7_SUB {
				fmt.Printf("[DEBUG] SUB: Funct7=%x, rs1=%d, rs2=%d, rd=%d\n", instr.Funct7(), rs1, rs2, rd)
				if rd != 0 {
					c.Reg[rd] = c.Reg[rs1] - c.Reg[rs2]
				}
			} else {
				return fmt.Errorf("unknown R-type funct7: 0x%X", instr.Funct7())
			}
		case assembler.FUNCT3_SLT:
			rd := instr.Rd()
			rs1 := instr.Rs1()
			rs2 := instr.Rs2()
			if rd != 0 {
				// signed comparison!
				if int32(c.Reg[rs1]) < int32(c.Reg[rs2]) {
					c.Reg[rd] = 1
				} else {
					c.Reg[rd] = 0
				}
			}
		default:
			fmt.Printf("[DEBUG] instr=%#v, type=%T, reflect.Kind=%v, uint32(instr)=%#x\n", instr, instr, reflect.TypeOf(instr).Kind(), uint32(instr))
			return fmt.Errorf("unknown R-type funct3: 0x%X", instr.Funct3())
		}
	case assembler.OPCODE_I_TYPE:
		switch instr.Funct3() {
		case assembler.FUNCT3_ADDI:
			rd := instr.Rd()
			rs1 := instr.Rs1()
			imm := instr.ImmI()
			if rd != 0 {
				c.Reg[rd] = c.Reg[rs1] + uint32(imm)
			}
		case assembler.FUNCT3_SLLI:
			rd := instr.Rd()
			rs1 := instr.Rs1()
			shamt := instr.ImmI() & 0x1F
			if rd != 0 {
				c.Reg[rd] = c.Reg[rs1] << shamt
			}
		default:
			return fmt.Errorf("unknown I-type funct3: 0x%X", instr.Funct3())
		}
	case assembler.OPCODE_BRANCH:
		rs1 := instr.Rs1()
		rs2 := instr.Rs2()
		imm := instr.ImmB()
		switch instr.Funct3() {
		case assembler.FUNCT3_BEQ:
			if c.Reg[rs1] == c.Reg[rs2] {
				c.PC = uint32(int32(c.PC) + imm)
				return nil
			} else {
				c.PC += INSTRUCTION_SIZE
				return nil
			}
		case assembler.FUNCT3_BNE:
			if c.Reg[rs1] != c.Reg[rs2] {
				c.PC = uint32(int32(c.PC) + imm)
				return nil
			} else {
				c.PC += INSTRUCTION_SIZE
				return nil
			}
		default:
			return fmt.Errorf("unknown branch funct3: 0x%X", instr.Funct3())
		}
	case assembler.OPCODE_JAL:
		rd := instr.Rd()
		imm := instr.ImmJ()
		if rd != 0 {
			c.Reg[rd] = c.PC + INSTRUCTION_SIZE
		}
		c.PC = uint32(int32(c.PC) + imm)
		return nil
	case assembler.OPCODE_JALR:
		rd := instr.Rd()
		rs1 := instr.Rs1()
		imm := instr.ImmI()
		next := c.PC + INSTRUCTION_SIZE
		target := (c.Reg[rs1] + uint32(imm)) &^ 1
		if rd != 0 {
			c.Reg[rd] = next
		}
		c.PC = target
		return nil
	default:
		return fmt.Errorf("unknown or unimplemented opcode: 0x%X (Opcode: 0x%X, type: %T)", instr.Opcode(), instr.Opcode(), instr)
	}
	return nil
}

func (c *CPU) Step(memory WordReader) error {
	word, err := memory.ReadWord(c.PC)
	fmt.Printf("[CPU] Step: PC = %#x, Instruction = %#x\n", c.PC, word)
	if err != nil {
		return err
	}
	instr := assembler.Instruction(word)
	fmt.Printf("[DEBUG-TYPE] instr=%#v, type=%T, reflect=%v\n", instr, instr, reflect.TypeOf(instr))
	err = c.exec(instr)
	if err != nil {
		return err
	}
	// Only increment PC if it wasn't already set (by branch/jump)
	switch instr.Opcode() {
	case assembler.OPCODE_BRANCH, assembler.OPCODE_JAL, assembler.OPCODE_JALR:
		// PC already set
	default:
		c.PC += INSTRUCTION_SIZE
	}
	return nil
}
