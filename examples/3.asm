  addi x1, x0, 7	# x1 = 7
  addi x2, x0, 7	# x2 = 7
  beq x1, x2, equal	# if x1 == x2, branch to equal
  addi x1, x0, 1	# gets skipped
equal:
  addi x3, x0, 99	# x3 = 99
