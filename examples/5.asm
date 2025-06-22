  addi x1, x0, 5 # x1 = 5 (counter)
loop:
  addi	x1, x1, -1	# x1 -= 1
  bne	x1, x0, loop	# if x1 != 0 repeat
