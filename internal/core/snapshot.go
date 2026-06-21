package core

import "codeberg.org/malik/riscvemu/arch/cpu"

// Snapshot is a value-type view of the emulator's state. It is
// returned by App.Snapshot so frontends can render without holding
// locks or contending with the CPU's internal state. Mutating a
// Snapshot does not affect the underlying emulator.
type Snapshot struct {
	PC        uint32
	Registers [32]uint32
	Qi        [32]cpu.RSTag
	RS        cpu.RSSnapshot
	Stats     cpu.Statistics
}
