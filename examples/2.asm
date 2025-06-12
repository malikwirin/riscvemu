addi x1 , x0 , 42 # x1 = 42
addi x2 , x0 , 100 # x2 = Adresse 100
sw x1 , 0(x2) # Speicher[100] = 42
lw x3 , 0(x2) # x3 = Speicher[100] −> 42
