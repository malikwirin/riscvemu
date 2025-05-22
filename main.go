package main

import (
    "fmt"
	"os"

    "github.com/malikwirin/riscvemu/arch"
    "github.com/malikwirin/riscvemu/cli"
    "github.com/malikwirin/riscvemu/assembler"
)

func main() {
    if len(os.Args) < 2 {
        fmt.Println("Usage: riscvemu <program.asm>")
        return
    }

    instructions, err := assembler.ParseAssemblyFile(os.Args[1])
    if err != nil {
        fmt.Println("Failed to parse assembly:", err)
        return
    }

    cpu := arch.NewCPU()
    mem := arch.NewMemory(4096)

    cli.StartREPL(cpu, mem, instructions)
}
