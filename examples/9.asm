  addi	x1, x0 ,21	# x1 = 21
  jal	x5, double	# after return x6 contains the result
  j	end

double:
  slli	x6, x1, 1	# x6 = x1 << 1 = x1 ∗ 2
  jalr	x0, 0(x5)	# return over ra
end:
  # x6 = 42
