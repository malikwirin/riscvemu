  jal	x0, target # jump without return
  addi	x1, x0, 1 # this instruction will be skipped
target:
  addi	x2, x0, 2 # x2 = 2
