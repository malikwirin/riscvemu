package assembler

type Instruction uint32

func (i Instruction) Opcode() uint32 {
    return uint32(i) & 0x7F
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

func (i *Instruction) SetOpcode(opcode uint32) {
    *i = Instruction((uint32(*i) &^ 0x7F) | (opcode & 0x7F))
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

func (i Instruction) Type() string {
    switch i.Opcode() {
    case 0x33:
        return "R"
    case 0x13, 0x03, 0x67:
        return "I"
    case 0x23:
        return "S"
    case 0x63:
        return "B"
    case 0x37, 0x17:
        return "U"
    case 0x6F:
        return "J"
    default:
        return "unknown"
    }
}
