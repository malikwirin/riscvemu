package cpu

import (
	"testing"

	"github.com/malikwirin/riscvemu/arch/cpu/cputest"
)

func TestMULComputesLowWord(t *testing.T) {
	core := makeCPU(t)
	fetchProgram(t, core,
		"addi x1, x0, 6",
		"addi x2, x0, 7",
		"mul x3, x1, x2",
	)
	for i := 0; i < 5; i++ {
		core.RunCycle()
	}
	if got := core.Reg(3); got != 42 {
		t.Errorf("x3 = %d, want 42 (6*7)", got)
	}
}

func TestMULHandlesLargeProducts(t *testing.T) {
	// 0x10000 * 0x10000 = 0x100000000, low 32 bits = 0
	core := makeCPU(t)
	fetchProgram(t, core,
		"addi x1, x0, -1", // x1 = 0xFFFFFFFF (unsigned: huge)
		"addi x2, x0, 2",
		"mul x3, x1, x2",
	)
	for i := 0; i < 5; i++ {
		core.RunCycle()
	}
	if got := core.Reg(3); got != 0xFFFFFFFE {
		t.Errorf("x3 = %#x, want 0xFFFFFFFE (low 32 bits of -1*2)", got)
	}
}

func TestDIVAndREM(t *testing.T) {
	cases := []struct {
		name     string
		x1, x2   int32
		op       string
		expected uint32
	}{
		// signed: 10 / 3 = 3, 10 mod 3 = 1
		{"div_signed", 10, 3, "div", 3},
		{"rem_signed", 10, 3, "rem", 1},
		// -10 / 3: RISC-V rounds toward zero: -3, -10 - (-3*3) = -1
		{"div_negative", -10, 3, "div", 0xFFFFFFFD}, // -3
		{"rem_negative", -10, 3, "rem", 0xFFFFFFFF}, // -1 (low 32 bits)
		// unsigned division: 0xFFFFFFFF / 2 = 0x7FFFFFFF
		{"divu_unsigned", -1, 2, "divu", 0x7FFFFFFF},
		// unsigned remainder: 0xFFFFFFFF mod 2 = 1
		{"remu_unsigned", -1, 2, "remu", 1},
		// divide by -1: -1 / -1 = 1
		{"div_by_neg1", -1, -1, "div", 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			core := makeCPU(t, withDivLatency(2))
			fetchProgram(t, core,
				"addi x1, x0, "+itoa(tc.x1),
				"addi x2, x0, "+itoa(tc.x2),
				tc.op+" x3, x1, x2",
			)
			for i := 0; i < 10; i++ {
				core.RunCycle()
			}
			if got := core.Reg(3); got != tc.expected {
				t.Errorf("x3 = %#x, want %#x", got, tc.expected)
			}
		})
	}
}

func TestDIVByZero(t *testing.T) {
	// RISC-V: division by zero returns -1, remainder by zero returns the dividend.
	cases := []struct {
		op       string
		x1, want uint32
	}{
		{"div", 42, 0xFFFFFFFF},  // -1
		{"divu", 42, 0xFFFFFFFF}, // 2^32-1
		{"rem", 42, 42},          // dividend
		{"remu", 42, 42},         // dividend
	}
	for _, tc := range cases {
		t.Run(tc.op, func(t *testing.T) {
			core := makeCPU(t, withDivLatency(2))
			fetchProgram(t, core,
				"addi x1, x0, "+itoa(int32(tc.x1)),
				"addi x2, x0, 0",
				tc.op+" x3, x1, x2",
			)
			for i := 0; i < 10; i++ {
				core.RunCycle()
			}
			if got := core.Reg(3); got != tc.want {
				t.Errorf("x3 = %#x, want %#x", got, tc.want)
			}
		})
	}
}

func TestMULHHighWord(t *testing.T) {
	// 0x10000 * 0x10000 = 0x1_00000000; high word = 1.
	// Build 0x10000 = 2048 + 2048 + 2048 + 2048 + 2048 + 2048 + 2048 + 2048
	// via chained addis (each addi is I-type, imm range -2048..2047).
	core := makeCPU(t, withMulLatency(3), withIQSize(16))
	fetchProgram(t, core,
		"addi x1, x0, 2047",
		"addi x1, x1, 1", // x1 = 2048
		"slli x1, x1, 4", // x1 = 0x8000
		"slli x1, x1, 1", // x1 = 0x10000
		"addi x2, x0, 2047",
		"addi x2, x2, 1", // x2 = 2048
		"slli x2, x2, 4", // x2 = 0x8000
		"slli x2, x2, 1", // x2 = 0x10000
		"mulh x3, x1, x2",
	)
	for i := 0; i < 20; i++ {
		core.RunCycle()
	}
	if got := core.Reg(3); got != 1 {
		t.Errorf("x3 = %d, want 1 (high word of 0x10000*0x10000)", got)
	}
}

func TestMULTracksFUBusyCycles(t *testing.T) {
	core := makeCPU(t, withMulLatency(3))
	if !core.Fetch(cputest.Encode(t, "addi x1, x0, 6"), 0) {
		t.Fatal("Fetch 1 failed")
	}
	if !core.Fetch(cputest.Encode(t, "addi x2, x0, 7"), 0) {
		t.Fatal("Fetch 2 failed")
	}
	if !core.Fetch(cputest.Encode(t, "mul x3, x1, x2"), 0) {
		t.Fatal("Fetch 3 failed")
	}
	for i := 0; i < 5; i++ {
		core.RunCycle()
	}
	if got := core.Stats().FunctionalBusyCycles[OpMUL]; got == 0 {
		t.Errorf("MUL busy cycles = 0, want > 0 (MUL FU was used)")
	}
}
