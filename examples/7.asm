  addi	x1, x0, 7	# x1 = n
  addi	x2, x0, 0	# x2 = F(0)
  addi	x3, x0, 1	# x3 = F(1)
  addi	x4, x0, 0	# x4 = loop-register
  addi	x5, x0, 1	# x5 = i
loop_fib:
  add	x4, x2, x3	# x4 = F(i) = F(i-2) + F(i-1)
  add	x2, x3, x0	# x2 = old F(i-1)
  add	x3, x4, x0	# x3 = new F(i)
  addi	x5, x5, 1	# i++
  bne	x5, x1, loop_fib	# finaly x4 contains F(7) = 13
