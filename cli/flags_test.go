package cli

import (
	"strings"
	"testing"

	"codeberg.org/malik/riscvemu/arch"
	"codeberg.org/malik/riscvemu/arch/cpu"
)

func TestConfigFromFlagsDefaults(t *testing.T) {
	cfg, memSize, err := ConfigFromFlags(nil)
	if err != nil {
		t.Fatalf("ConfigFromFlags(nil): %v", err)
	}
	if memSize != 64*1024 {
		t.Errorf("memSize = %d, want %d", memSize, 64*1024)
	}
	// SpecConfig defaults
	if cfg.ALURSCount != 4 {
		t.Errorf("ALURSCount = %d, want 4", cfg.ALURSCount)
	}
	if cfg.LSURSCount != 3 {
		t.Errorf("LSURSCount = %d, want 3", cfg.LSURSCount)
	}
	if cfg.RegisterCount != 8 {
		t.Errorf("RegisterCount = %d, want 8", cfg.RegisterCount)
	}
}

func TestConfigFromFlagsOverrides(t *testing.T) {
	cfg, memSize, err := ConfigFromFlags([]string{
		"-alu-rs", "8",
		"-alu-lat", "3",
		"-load-lat", "5",
		"-regs", "32",
		"-mem", "32768",
	})
	if err != nil {
		t.Fatalf("ConfigFromFlags: %v", err)
	}
	if cfg.ALURSCount != 8 {
		t.Errorf("ALURSCount = %d, want 8", cfg.ALURSCount)
	}
	if cfg.ALULatency != 3 {
		t.Errorf("ALULatency = %d, want 3", cfg.ALULatency)
	}
	if cfg.LoadLatency != 5 {
		t.Errorf("LoadLatency = %d, want 5", cfg.LoadLatency)
	}
	if cfg.RegisterCount != 32 {
		t.Errorf("RegisterCount = %d, want 32", cfg.RegisterCount)
	}
	if memSize != 32768 {
		t.Errorf("memSize = %d, want 32768", memSize)
	}
}

func TestConfigFromFlagsUnknown(t *testing.T) {
	_, _, err := ConfigFromFlags([]string{"-unknown", "1"})
	if err == nil {
		t.Fatal("expected error for unknown flag")
	}
	if !strings.Contains(err.Error(), "unknown flag") {
		t.Errorf("error = %q, want 'unknown flag'", err)
	}
}

func TestConfigFromFlagsInvalidValue(t *testing.T) {
	_, _, err := ConfigFromFlags([]string{"-alu-rs", "abc"})
	if err == nil {
		t.Fatal("expected error for invalid value")
	}
	if !strings.Contains(err.Error(), "invalid value") {
		t.Errorf("error = %q, want 'invalid value'", err)
	}
}

func TestConfigFromFlagsMissingValue(t *testing.T) {
	_, _, err := ConfigFromFlags([]string{"-alu-rs"})
	if err == nil {
		t.Fatal("expected error for missing value")
	}
	if !strings.Contains(err.Error(), "requires a value") {
		t.Errorf("error = %q, want 'requires a value'", err)
	}
}

func TestConfigFromFlagsHelp(t *testing.T) {
	_, _, err := ConfigFromFlags([]string{"-help"})
	if !HelpRequested(err) {
		t.Errorf("HelpRequested(err) = false, want true (err=%v)", err)
	}
}

func TestConfigFromFlagsValidation(t *testing.T) {
	// alu-rs = 0 must fail validation.
	_, _, err := ConfigFromFlags([]string{"-alu-rs", "0"})
	if err == nil {
		t.Fatal("expected error for alu-rs=0")
	}
	if !strings.Contains(err.Error(), "alu-rs must be >= 1") {
		t.Errorf("error = %q, want 'alu-rs must be >= 1'", err)
	}
}

func TestCmdConfigPrintsActiveConfig(t *testing.T) {
	cfg := cpu.Config{
		ALURSCount:           7,
		LSURSCount:           5,
		ALULatency:           2,
		LoadLatency:          3,
		StoreLatency:         4,
		MulRSCount:           2,
		DivRSCount:           1,
		MulLatency:           6,
		DivLatency:           9,
		InstructionQueueSize: 12,
		RegisterCount:        16,
	}
	withMachineConfig(1024, cfg, func(_ *arch.Machine, owner *testOwner) {
		out := captureOutput(func() {
			if err := cmdConfig(owner, nil); err != nil {
				t.Fatalf("cmdConfig: %v", err)
			}
		})
		for _, want := range []string{
			"ALURSCount:           7",
			"LSURSCount:           5",
			"ALULatency:           2",
			"LoadLatency:          3",
			"StoreLatency:         4",
			"MulRSCount:           2",
			"DivRSCount:           1",
			"MulLatency:           6",
			"DivLatency:           9",
			"InstructionQueueSize: 12",
			"RegisterCount:        16",
		} {
			if !strings.Contains(out, want) {
				t.Errorf("output missing %q\nfull output:\n%s", want, out)
			}
		}
	})
}

func TestCmdRegsRespectsRegisterCount(t *testing.T) {
	cfg := cpu.DefaultConfig() // 32 registers
	withMachineConfig(1024, cfg, func(_ *arch.Machine, owner *testOwner) {
		out := captureOutput(func() {
			if err := cmdRegs(owner, nil); err != nil {
				t.Fatalf("cmdRegs: %v", err)
			}
		})
		// 32-register layout must show x0..x31
		if !strings.Contains(out, "x0:") || !strings.Contains(out, "x31:") {
			t.Errorf("32-register layout missing x0/x31 markers\noutput:\n%s", out)
		}
	})

	cfg = cpu.SpecConfig() // 8 registers
	withMachineConfig(1024, cfg, func(_ *arch.Machine, owner *testOwner) {
		out := captureOutput(func() {
			if err := cmdRegs(owner, nil); err != nil {
				t.Fatalf("cmdRegs: %v", err)
			}
		})
		// 8-register layout must show x0..x7 but not x8
		if !strings.Contains(out, "x7:") {
			t.Errorf("8-register layout missing x7:\noutput:\n%s", out)
		}
		if strings.Contains(out, "x8:") {
			t.Errorf("8-register layout should not contain x8:\noutput:\n%s", out)
		}
	})
}
