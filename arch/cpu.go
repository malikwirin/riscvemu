package arch

import (
	"fmt"
	"github.com/malikwirin/riscvemu/assembler"
	"reflect"
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

// Handles R-type arithmetic (add, sub, slt, ...)
func (c *CPU) aluRType(rd, rs1, rs2 uint32, op func(a, b uint32) uint32) {
	if rd == 0 {
		return
	}
	c.Reg[rd] = op(c.Reg[rs1], c.Reg[rs2])
}

// Handles I-type arithmetic (addi, ...)
func (c *CPU) aluIType(rd, rs1 uint32, imm int32, op func(a uint32, b int32) uint32) {
	if rd == 0 {
		return
	}
	c.Reg[rd] = op(c.Reg[rs1], imm)
}

// Handles Branches (beq, bne, blt, ...)
func (c *CPU) branch(rs1, rs2 uint32, imm int32, cond func(a, b uint32) bool) {
	if cond(c.Reg[rs1], c.Reg[rs2]) {
		c.PC = uint32(int32(c.PC) + imm)
	} else {
		c.PC += INSTRUCTION_SIZE
	}
}

func (c *CPU) exec(instr assembler.Instruction, memory WordHandler) error {
	opcode := instr.Opcode()
	if opcode == assembler.OPCODE_INVALID {
		return fmt.Errorf("invalid opcode: 0x%X (from instruction 0x%X)", opcode, uint32(instr))
	}
	switch opcode {
	case assembler.OPCODE_R_TYPE:
		rd, rs1, rs2 := instr.Rd(), instr.Rs1(), instr.Rs2()
		switch instr.Funct3() {
		case assembler.FUNCT3_ADD_SUB:
			switch instr.Funct7() {
			case assembler.FUNCT7_ADD:
				c.aluRType(rd, rs1, rs2, func(a, b uint32) uint32 { return a + b })
			case assembler.FUNCT7_SUB:
				c.aluRType(rd, rs1, rs2, func(a, b uint32) uint32 { return a - b })
			default:
				return fmt.Errorf("unknown R-type funct7: 0x%X", instr.Funct7())
			}
		case assembler.FUNCT3_SLT:
			c.aluRType(rd, rs1, rs2, func(a, b uint32) uint32 {
				if int32(a) < int32(b) {
					return 1
				}
				return 0
			})
		default:
			fmt.Printf("[DEBUG] instr=%#v, type=%T, reflect.Kind=%v, uint32(instr)=%#x\n", instr, instr, reflect.TypeOf(instr).Kind(), uint32(instr))
			return fmt.Errorf("unknown R-type funct3: 0x%X", instr.Funct3())
		}
	case assembler.OPCODE_I_TYPE:
		rd, rs1, imm := instr.Rd(), instr.Rs1(), instr.ImmI()
		switch instr.Funct3() {
		case assembler.FUNCT3_ADDI:
			c.aluIType(rd, rs1, imm, func(a uint32, b int32) uint32 { return a + uint32(b) })
		case assembler.FUNCT3_SLLI:
			shamt := imm & 0x1F
			c.aluIType(rd, rs1, shamt, func(a uint32, b int32) uint32 { return a << b })
		default:
			return fmt.Errorf("unknown I-type funct3: 0x%X", instr.Funct3())
		}
	case assembler.OPCODE_LOAD:
		rd, rs1, imm := instr.Rd(), instr.Rs1(), instr.ImmI()
		addr := c.Reg[rs1] + uint32(imm)
		switch instr.Funct3() {
		case assembler.FUNCT3_LW: // Load Word
			value, err := memory.ReadWord(addr)
			if err != nil {
				return fmt.Errorf("LOAD failed: %w", err)
			}
			if rd != 0 {
				c.Reg[rd] = value
			}
		default:
			return fmt.Errorf("unsupported LOAD funct3: 0x%X", instr.Funct3())
		}
	case assembler.OPCODE_STORE:
		rs1, rs2, imm := instr.Rs1(), instr.Rs2(), instr.ImmS()
		addr := c.Reg[rs1] + uint32(imm)
		value := c.Reg[rs2]
		switch instr.Funct3() {
		case assembler.FUNCT3_SW: // Store Word
			return memory.WriteWord(addr, value)
		default:
			return fmt.Errorf("unsupported STORE funct3: 0x%X", instr.Funct3())
		}
	case assembler.OPCODE_BRANCH:
		rs1, rs2, imm := instr.Rs1(), instr.Rs2(), instr.ImmB()
		switch instr.Funct3() {
		case assembler.FUNCT3_BEQ:
			c.branch(rs1, rs2, imm, func(a, b uint32) bool { return a == b })
		case assembler.FUNCT3_BNE:
			c.branch(rs1, rs2, imm, func(a, b uint32) bool { return a != b })
		case assembler.FUNCT3_SLT: // BLT
			c.branch(rs1, rs2, imm, func(a, b uint32) bool { return int32(a) < int32(b) })
		default:
			return fmt.Errorf("unknown branch funct3: 0x%X", instr.Funct3())
		}
	case assembler.OPCODE_JAL:
		rd, imm := instr.Rd(), instr.ImmJ()
		if rd != 0 {
			c.Reg[rd] = c.PC + INSTRUCTION_SIZE
		}
		c.PC = uint32(int32(c.PC) + imm)
	case assembler.OPCODE_JALR:
		rd, rs1, imm := instr.Rd(), instr.Rs1(), instr.ImmI()
		next := c.PC + INSTRUCTION_SIZE
		target := (c.Reg[rs1] + uint32(imm)) &^ 1
		if rd != 0 {
			c.Reg[rd] = next
		}
		c.PC = target
	default:
		return fmt.Errorf("unknown or unimplemented opcode: 0x%X (Opcode: 0x%X, type: %T)", instr.Opcode(), instr.Opcode(), instr)
	}
	return nil
}

func (c *CPU) Step(memory WordHandler) error {
	word, err := memory.ReadWord(c.PC)
	fmt.Printf("[CPU] Step: PC = %#x, Instruction = %#x\n", c.PC, word)
	if err != nil {
		return err
	}
	instr := assembler.Instruction(word)
	err = c.exec(instr, memory)
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
