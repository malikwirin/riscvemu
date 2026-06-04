package cpu

import (
	"testing"

	"github.com/malikwirin/riscvemu/assembler"
)

// helper: encode an instruction line into a word.
func encode(t *testing.T, line string) uint32 {
	t.Helper()
	instr, err := assembler.ParseInstruction(line)
	if err != nil {
		t.Fatalf("ParseInstruction(%q): %v", line, err)
	}
	return uint32(instr)
}

func TestIssueADDIPlacesInALURS(t *testing.T) {
	core := NewCPU(DefaultConfig())
	word := encode(t, "addi x1, x0, 5")
	if err := core.ReceiveInstruction(word); err != nil {
		t.Fatalf("ReceiveInstruction: %v", err)
	}
	if got := core.rs.ALUCountBusy(); got != 1 {
		t.Errorf("ALU RS busy = %d, want 1", got)
	}
	if got := core.rs.LSUCountBusy(); got != 0 {
		t.Errorf("LSU RS busy = %d, want 0", got)
	}
}

func TestIssueLWPPlacesInLSURS(t *testing.T) {
	core := NewCPU(DefaultConfig())
	word := encode(t, "lw x3, 0(x2)")
	if err := core.ReceiveInstruction(word); err != nil {
		t.Fatalf("ReceiveInstruction: %v", err)
	}
	if got := core.rs.LSUCountBusy(); got != 1 {
		t.Errorf("LSU RS busy = %d, want 1", got)
	}
	if got := core.rs.ALUCountBusy(); got != 0 {
		t.Errorf("ALU RS busy = %d, want 0", got)
	}
}

func TestIssueSetsQiForDestination(t *testing.T) {
	core := NewCPU(DefaultConfig())
	word := encode(t, "addi x1, x0, 5")
	if err := core.ReceiveInstruction(word); err != nil {
		t.Fatalf("ReceiveInstruction: %v", err)
	}
	if core.rf.Qi[1] == NoTag {
		t.Errorf("Qi[x1] = NoTag after issuing addi to x1")
	}
}

func TestIssueDoesNotSetQiForX0(t *testing.T) {
	core := NewCPU(DefaultConfig())
	word := encode(t, "addi x0, x0, 0")
	if err := core.ReceiveInstruction(word); err != nil {
		t.Fatalf("ReceiveInstruction: %v", err)
	}
	if core.rf.Qi[0] != NoTag {
		t.Errorf("Qi[x0] must stay NoTag, got %d", core.rf.Qi[0])
	}
}

func TestStructuralStallWhenALUFull(t *testing.T) {
	cfg := DefaultConfig()
	cfg.ALURSCount = 2
	core := NewCPU(cfg)
	// Fill the ALU RS.
	if err := core.ReceiveInstruction(encode(t, "addi x1, x0, 1")); err != nil {
		t.Fatalf("issue 1: %v", err)
	}
	if err := core.ReceiveInstruction(encode(t, "addi x2, x0, 2")); err != nil {
		t.Fatalf("issue 2: %v", err)
	}
	// This one must stall.
	if err := core.ReceiveInstruction(encode(t, "addi x3, x0, 3")); err != nil {
		t.Fatalf("issue 3: %v", err)
	}
	if got := core.Stats().StructuralStalls; got != 1 {
		t.Errorf("StructuralStalls = %d, want 1", got)
	}
}

func TestIssueCapturesImmediateOperands(t *testing.T) {
	core := NewCPU(DefaultConfig())
	word := encode(t, "addi x1, x0, 42")
	if err := core.ReceiveInstruction(word); err != nil {
		t.Fatalf("ReceiveInstruction: %v", err)
	}
	// The RS entry for x1 should have Imm = 42.
	var found *RSEntry
	for i := range core.rs.alu {
		if core.rs.alu[i].Busy && core.rs.alu[i].Rd == 1 {
			found = &core.rs.alu[i]
			break
		}
	}
	if found == nil {
		t.Fatal("RS entry for x1 not found")
	}
	if found.Imm != 42 {
		t.Errorf("RS entry Imm = %d, want 42", found.Imm)
	}
}

func TestIssueTakesOperandValueWhenReady(t *testing.T) {
	core := NewCPU(DefaultConfig())
	// x1 <- 7 first
	if err := core.ReceiveInstruction(encode(t, "addi x1, x0, 7")); err != nil {
		t.Fatalf("issue 1: %v", err)
	}
	// Commit x1 directly to the register file (simulating writeback completion)
	core.rf.V[1] = 7
	core.rf.Qi[1] = NoTag
	// Now issue add x2, x1, x0 — should capture Vj = 7
	if err := core.ReceiveInstruction(encode(t, "add x2, x1, x0")); err != nil {
		t.Fatalf("issue 2: %v", err)
	}
	var found *RSEntry
	for i := range core.rs.alu {
		if core.rs.alu[i].Busy && core.rs.alu[i].Rd == 2 {
			found = &core.rs.alu[i]
			break
		}
	}
	if found == nil {
		t.Fatal("RS entry for x2 not found")
	}
	if found.Vj != 7 {
		t.Errorf("RS entry Vj = %d, want 7", found.Vj)
	}
	if found.Qj != NoTag {
		t.Errorf("RS entry Qj = %d, want NoTag (operand was ready)", found.Qj)
	}
}

func TestIssueSetsTagWhenOperandPending(t *testing.T) {
	core := NewCPU(DefaultConfig())
	// First: addi x1, x0, 7 (sets Qi[x1] to the ALU tag)
	if err := core.ReceiveInstruction(encode(t, "addi x1, x0, 7")); err != nil {
		t.Fatalf("issue 1: %v", err)
	}
	// Second: add x2, x1, x0 — should capture Qj = Qi[x1] (operand pending)
	if err := core.ReceiveInstruction(encode(t, "add x2, x1, x0")); err != nil {
		t.Fatalf("issue 2: %v", err)
	}
	var found *RSEntry
	for i := range core.rs.alu {
		if core.rs.alu[i].Busy && core.rs.alu[i].Rd == 2 {
			found = &core.rs.alu[i]
			break
		}
	}
	if found == nil {
		t.Fatal("RS entry for x2 not found")
	}
	if found.Qj == NoTag {
		t.Errorf("RS entry Qj = NoTag, want a tag (operand was pending)")
	}
}
