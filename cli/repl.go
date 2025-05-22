package cli

import (
    "bufio"
    "fmt"
    "os"
    "strings"
)

func StartREPL(cpu *CPU, mem *Memory, program []Instruction) {
    reader := bufio.NewReader(os.Stdin)
    for {
        fmt.Print("riscv> ")
        input, _ := reader.ReadString('\n')
        input = strings.TrimSpace(input)
        switch input {
        case "step":
            // Einen Befehl ausführen
        case "regs":
            // Register anzeigen
        case "dump":
            // Speicher anzeigen
        case "exit":
            return
        default:
            fmt.Println("Unknown command")
        }
    }
}
