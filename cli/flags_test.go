package cli

import (
	"strings"
	"testing"

	"codeberg.org/malik/riscvemu/arch"
	"codeberg.org/malik/riscvemu/arch/cpu"
)

// TestConfigFromFlags pins the flag parser: defaults, single
// overrides, and every error class (unknown flag, invalid
// value, missing value, validation). Each case is one row
// of the table, so adding a new flag is a new row, not a
// new test function.
func TestConfigFromFlags(t *testing.T) {
	cases := []struct {
		name        string
		args        []string
		wantErr     string
		wantHelp    bool
		wantALU     int
		wantRegs    int
		wantMemSize int
	}{
		{
			name:        "defaults",
			args:        nil,
			wantALU:     4,
			wantRegs:    8,
			wantMemSize: 64 * 1024,
		},
		{
			name:        "overrides multiple",
			args:        []string{"-alu-rs", "8", "-alu-lat", "3", "-load-lat", "5", "-regs", "32", "-mem", "32768"},
			wantALU:     8,
			wantRegs:    32,
			wantMemSize: 32768,
		},
		{
			name:    "unknown flag",
			args:    []string{"-unknown", "1"},
			wantErr: "unknown flag",
		},
		{
			name:    "invalid value",
			args:    []string{"-alu-rs", "abc"},
			wantErr: "invalid value",
		},
		{
			name:    "missing value",
			args:    []string{"-alu-rs"},
			wantErr: "requires a value",
		},
		{
			name:     "help requested",
			args:     []string{"-help"},
			wantHelp: true,
		},
		{
			name:    "validation rejects zero",
			args:    []string{"-alu-rs", "0"},
			wantErr: "alu-rs must be >= 1",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, memSize, err := ConfigFromFlags(tc.args)
			if tc.wantErr != "" {
				if err == nil {
					t.Fatalf("want error %q, got nil", tc.wantErr)
				}
				if !strings.Contains(err.Error(), tc.wantErr) {
					t.Errorf("error = %q, want substring %q", err, tc.wantErr)
				}
				return
			}
			if err != nil {
				if tc.wantHelp && !HelpRequested(err) {
					t.Fatalf("want help-requested error, got %v", err)
				}
				if !tc.wantHelp {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if tc.wantALU != 0 && cfg.ALURSCount != tc.wantALU {
				t.Errorf("ALURSCount = %d, want %d", cfg.ALURSCount, tc.wantALU)
			}
			if tc.wantRegs != 0 && cfg.RegisterCount != tc.wantRegs {
				t.Errorf("RegisterCount = %d, want %d", cfg.RegisterCount, tc.wantRegs)
			}
			if tc.wantMemSize != 0 && memSize != tc.wantMemSize {
				t.Errorf("memSize = %d, want %d", memSize, tc.wantMemSize)
			}
		})
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

// TestCmdRegsRespectsRegisterCount pins that the regs command
// honours the configured register count: 32-register layouts
// show x0..x31, 8-register Spec layouts show x0..x7 and stop.
func TestCmdRegsRespectsRegisterCount(t *testing.T) {
	cases := []struct {
		name     string
		cfg      cpu.Config
		mustHave []string
		mustMiss []string
	}{
		{
			name:     "32-register default shows x0..x31",
			cfg:      cpu.DefaultConfig(),
			mustHave: []string{"x0:", "x31:"},
		},
		{
			name:     "8-register Spec stops at x7",
			cfg:      cpu.SpecConfig(),
			mustHave: []string{"x7:"},
			mustMiss: []string{"x8:"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			withMachineConfig(1024, tc.cfg, func(_ *arch.Machine, owner *testOwner) {
				out := captureOutput(func() {
					if err := cmdRegs(owner, nil); err != nil {
						t.Fatalf("cmdRegs: %v", err)
					}
				})
				for _, w := range tc.mustHave {
					if !strings.Contains(out, w) {
						t.Errorf("output missing %q\noutput:\n%s", w, out)
					}
				}
				for _, w := range tc.mustMiss {
					if strings.Contains(out, w) {
						t.Errorf("output should not contain %q\noutput:\n%s", w, out)
					}
				}
			})
		})
	}
}
