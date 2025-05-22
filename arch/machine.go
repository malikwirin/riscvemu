package arch

type Machine struct {
     CPU    *CPU
     Memory *Memory
}

func NewMachine(memSize int) *Machine {
    return &Machine{
        CPU:    NewCPU(),
        Memory: NewMemory(memSize),
    }
}

func (m *Machine) Step() error {
    return m.CPU.Step(m.Memory)
}

func (m *Machine) Reset() error {
	return nil
}
