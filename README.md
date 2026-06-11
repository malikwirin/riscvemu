# riscvemu

A simple, test-driven RISC-V emulator for educational purposes, written in Go.

## Overview

**riscvemu** is a minimal RISC-V emulator designed to help students and hobbyists understand the internals of the RISC-V architecture. The project focuses on clarity and transparency, following the principles of Test Driven Development (TDD). All essential CPU features are implemented with comprehensive unit tests, and a simple command-line REPL interface allows interactive experimentation.

## Features

- Implements a subset of the RISC-V RV32I instruction set
- Tomasulo-style out-of-order execution with reservation stations, common data bus, and implicit register renaming
- Interactive REPL for loading, running, and inspecting programs
- Per-FU execution statistics (cycles, IPC, stalls, utilisation)
- Memory and register inspection and manipulation
- Assembler for a small set of supported instructions
- Test-driven, with extensive unit and integration tests
- Easily extensible for new instructions or features

## Architecture

`riscvemu` models the CPU as a Tomasulo implementation with the classic
data flow:

- **In-order issue.** Instructions are dispatched to reservation
  stations (RS) in program order. A full RS or an unresolved conditional
  branch stalls the issue stage.
- **Out-of-order execute.** The ALU and LSU pools each contain one or
  more functional units (FUs). An RS entry is dispatched to a free FU
  as soon as its operands are ready, regardless of program order.
- **Out-of-order completion.** Completed results are broadcast on a
  single common data bus (CDB) and commit to the architectural register
  file. One CDB broadcast per cycle.
- **Implicit register renaming.** `Qi[32]` records the rename tag of
  the in-flight producer of each architectural register. A free RS
  receives a monotonically increasing tag, which is the only identity
  the rest of the machine uses.
- **Stall-on-branch.** Conditional branches block issue until they
  resolve. JAL and JALR are unconditional and do not block.
- **Configurable.** RS counts and per-FU latencies live in `Config`
  (`arch/cpu/config.go`); see `DefaultConfig` for the defaults
  (3 ALU RS, 2 LSU RS, ALU latency 1, load latency 2, store latency 2,
  instruction queue 8).

## Quick Start

### 1. Build & Run

```sh
git clone https://github.com/malikwirin/riscvemu.git
cd riscvemu
go build -o riscvemu
./riscvemu
```

### 2. Using the REPL

After starting, you'll see a prompt. Try commands like:

- `help` – list available commands
- `load examples/1.asm` – load an example RISC-V assembly program
- `step 5` – execute 5 cycles
- `regs` – print all registers
- `mem 0 16` – dump the first 16 words of memory
- `stats` – print execution statistics
- `randstore 100 10` – fill memory at address 100 with 10 random 32-bit words

### 3. REPL Commands

The REPL exposes the following commands. `step` and `reset` are the only
ones that change machine state. Everything else is read-only.

| Command | Syntax | Description |
| --- | --- | --- |
| `help` | `help [command]` | List all commands, or show help for a specific one. |
| `load` | `load <filename> [address]` | Assemble a `.asm` file and load it at `address` (default `0`). |
| `step` | `step [n]` | Advance the machine by `n` cycles (default `1`). Under the Tomasulo model this is **cycles**, not instructions; a chain of dependent ADDs needs several cycles per instruction. |
| `regs` | `regs` | Print `x0`..`x31`. |
| `pc` | `pc` | Print the current program counter. |
| `peek` | `peek` | Print the next instruction word at the current PC as hex. |
| `mem` | `mem [start [length]]` | Dump memory words starting at `start` (default `0`) for `length` words (default `16`). |
| `store` | `store <address> <v1> [v2 ...]` | Write one or more 32-bit values to memory. |
| `randstore` | `randstore <address> <count>` | Write `count` random 32-bit values to memory. |
| `reset` | `reset` | Reset the CPU and memory to the initial state. |
| `stats` | `stats` | Print execution statistics: cycles, retired, IPC, issued, structural and branch stalls, RAW resolutions, and a per-FU utilisation table. |
| `quit` / `exit` | `quit` / `exit` | Leave the REPL. |

### 4. Writing and Running Programs

Write your RISC-V assembly programs (see the provided `.asm` files as templates in `examples/`).  
Load your program in the REPL with `load <filename>`.  
You can also use the `store` and `randstore` commands to initialize memory before running your program.

## Project Structure

- `arch/` – Core emulator logic (CPU, memory, machine)
  - `arch/cpu/` – Tomasulo core: reservation stations, functional units, common data bus, register renaming, statistics
- `assembler/` – Assembly parsing and encoding
- `cli/` – REPL and command-line interface
- `examples/` – Example assembly programs
- `tests/` – End-to-end integration tests

## Test Driven Development

This project is developed following TDD principles.  
You can run all tests using:

```sh
go test ./...
```

## Example

```asm
# examples/1.asm
addi x1, x0, 5     # x1 = 5
addi x2, x0, 10    # x2 = 10
add  x3, x1, x2    # x3 = x1 + x2 = 15
```

In the REPL:

```
load examples/1.asm
step 3
regs
```

## Requirements

- Go 1.20 or newer

## License

This project is licensed under the GNU Affero General Public License v3.0.  
See [LICENCE.md](LICENCE.md) for details.
[![License: AGPL v3](https://img.shields.io/badge/License-AGPL_v3-blue.svg)](https://www.gnu.org/licenses/agpl-3.0)

---

**Educational use and contributions are welcome!**
