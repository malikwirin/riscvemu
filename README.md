# riscvemu

A simple, test-driven RISC-V emulator for educational purposes, written in Go.

## Overview

**riscvemu** is a minimal RISC-V emulator designed to help students and hobbyists understand the internals of the RISC-V architecture. The project focuses on clarity and transparency, following the principles of Test Driven Development (TDD). All essential CPU features are implemented with comprehensive unit tests, and a simple command-line REPL interface allows interactive experimentation.

## Features

- Implements a subset of the RISC-V RV32I instruction set
- Interactive REPL for loading, running, and inspecting programs
- Memory and register inspection and manipulation
- Assembler for a small set of supported instructions
- Test-driven, with extensive unit and integration tests
- Easily extensible for new instructions or features

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
- `step 5` – execute 5 instructions
- `regs` – print all registers
- `mem 0 16` – dump the first 16 words of memory
- `randstore 100 10` – fill memory at address 100 with 10 random 32-bit words

### 3. Writing and Running Programs

Write your RISC-V assembly programs (see the provided `.asm` files as templates in `examples/`).  
Load your program in the REPL with `load <filename>`.  
You can also use the `store` and `randstore` commands to initialize memory before running your program.

## Project Structure

- `arch/` – Core emulator logic (CPU, memory, machine)
- `assembler/` – Assembly parsing and encoding
- `cli/` – REPL and command-line interface
- `examples/` – Example assembly programs

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
