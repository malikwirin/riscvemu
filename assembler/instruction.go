package assembler

type Instruction uint32

func (i Instruction) Opcode() Opcode {
	return Opcode(uint32(i) & 0x7F)
}

func (i Instruction) Rd() uint32 {
	return (uint32(i) >> 7) & 0x1F
}

func (i Instruction) Funct3() uint32 {
	return (uint32(i) >> 12) & 0x7
}

func (i Instruction) Rs1() uint32 {
	return (uint32(i) >> 15) & 0x1F
}

func (i Instruction) Rs2() uint32 {
	return (uint32(i) >> 20) & 0x1F
}

func (i Instruction) Funct7() uint32 {
	return (uint32(i) >> 25) & 0x7F
}

func (i *Instruction) SetOpcode(opcode Opcode) {
	*i = Instruction((uint32(*i) &^ 0x7F) | (uint32(opcode) & 0x7F))
}

func (i *Instruction) SetRd(rd uint32) {
	*i = Instruction((uint32(*i) &^ (0x1F << 7)) | ((rd & 0x1F) << 7))
}

func (i *Instruction) SetFunct3(funct3 uint32) {
	*i = Instruction((uint32(*i) &^ (0x7 << 12)) | ((funct3 & 0x7) << 12))
}

func (i *Instruction) SetRs1(rs1 uint32) {
	*i = Instruction((uint32(*i) &^ (0x1F << 15)) | ((rs1 & 0x1F) << 15))
}

func (i *Instruction) SetRs2(rs2 uint32) {
	*i = Instruction((uint32(*i) &^ (0x1F << 20)) | ((rs2 & 0x1F) << 20))
}

func (i *Instruction) SetFunct7(funct7 uint32) {
	*i = Instruction((uint32(*i) &^ (0x7F << 25)) | ((funct7 & 0x7F) << 25))
}

func (i Instruction) ImmI() int32 {
	// 12 Bit signed immediate (bits 20-31)
	imm := int32(uint32(i) >> 20)
	if imm&0x800 != 0 {
		imm |= ^0xFFF // sign extension
	}
	return imm
}

func (i Instruction) ImmS() int32 {
	// S-type: bits 7-11 | 25-31
	imm := int32(((uint32(i) >> 7) & 0x1F) | (((uint32(i) >> 25) & 0x7F) << 5))
	if imm&0x800 != 0 {
		imm |= ^0xFFF
	}
	return imm
}

func (i Instruction) ImmB() int32 {
	// B-type: [31] [7] [30:25] [11:8] << 1
	imm := int32(
		(((uint32(i) >> 31) & 0x1) << 12) |
			(((uint32(i) >> 7) & 0x1) << 11) |
			(((uint32(i) >> 25) & 0x3F) << 5) |
			(((uint32(i) >> 8) & 0xF) << 1),
	)
	if imm&0x1000 != 0 {
		imm |= ^0x1FFF
	}
	return imm
}

func (i Instruction) ImmJ() int32 {
	// J-type: [31] [19:12] [20] [30:21] << 1
	imm := int32(
		(((uint32(i) >> 31) & 0x1) << 20) |
			(((uint32(i) >> 12) & 0xFF) << 12) |
			(((uint32(i) >> 20) & 0x1) << 11) |
			(((uint32(i) >> 21) & 0x3FF) << 1),
	)
	if imm&0x100000 != 0 {
		imm |= ^0xFFFFF
	}
	return imm
}

func (i *Instruction) SetImmI(imm int32) {
	// 12-bit signed immediate at bits 20-31
	ui := uint32(*i) &^ (0xFFF << 20)
	*i = Instruction(ui | ((uint32(imm) & 0xFFF) << 20))
}

func (i *Instruction) SetImmS(imm int32) {
	// S-type: bits 7-11 | 25-31
	ui := uint32(*i)
	ui &^= (0x1F << 7)  // clear bits 7-11
	ui &^= (0x7F << 25) // clear bits 25-31
	ui |= (uint32(imm) & 0x1F) << 7
	ui |= ((uint32(imm) >> 5) & 0x7F) << 25
	*i = Instruction(ui)
}

func (i *Instruction) SetImmB(imm int32) {
	// B-type: [31] [7] [30:25] [11:8]
	ui := uint32(*i)
	ui &^= (1 << 31)    // bit 31 (imm[12])
	ui &^= (1 << 7)     // bit 7  (imm[11])
	ui &^= (0x3F << 25) // bits 25-30 (imm[10:5])
	ui &^= (0xF << 8)   // bits 8-11 (imm[4:1])

	ui |= ((uint32(imm) >> 12) & 0x1) << 31
	ui |= ((uint32(imm) >> 11) & 0x1) << 7
	ui |= ((uint32(imm) >> 5) & 0x3F) << 25
	ui |= ((uint32(imm) >> 1) & 0xF) << 8
	*i = Instruction(ui)
}

func (i *Instruction) SetImmJ(imm int32) {
	// J-type: [31] [19:12] [20] [30:21]
	ui := uint32(*i)
	ui &^= (1 << 31)     // bit 31 (imm[20])
	ui &^= (0xFF << 12)  // bits 12-19 (imm[19:12])
	ui &^= (1 << 20)     // bit 20 (imm[11])
	ui &^= (0x3FF << 21) // bits 21-30 (imm[10:1])

	ui |= ((uint32(imm) >> 20) & 0x1) << 31
	ui |= ((uint32(imm) >> 12) & 0xFF) << 12
	ui |= ((uint32(imm) >> 11) & 0x1) << 20
	ui |= ((uint32(imm) >> 1) & 0x3FF) << 21
	*i = Instruction(ui)
}

func (i Instruction) Type() string {
	switch i.Opcode() {
	case OPCODE_R_TYPE:
		return "R"
	case OPCODE_I_TYPE, OPCODE_LOAD, OPCODE_JALR:
		return "I"
	case OPCODE_STORE:
		return "S"
	case OPCODE_BRANCH:
		return "B"
	case OPCODE_JAL:
		return "J"
	default:
		return "unknown"
	}
}
