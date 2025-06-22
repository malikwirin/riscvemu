# −−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−
# Array lays at address 100
# length = 5 (4 Bytes per Element)
# solution (maximum) in x3
# −−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−−
  addi	x1, x0, 100	# x1 = base address of the array
  addi	x2, x0, 5	# x2 = number of elements
  addi	x4, x1, 0	# x4 = pointer to the first element

  lw	x3, 0(x4)	# x3 = first element (initial maximum)
  addi	x2, x2, −1	# x2 = remaining elements

loop_max:
  addi	x4, x4, 4	# ptr += 4 (next element)
  lw	x5, 0(x4)	# x5 = current element

  slt	x6, x3, x5	# x6 = 1, if x3 < x5
  bne	x6, x0, update	# if new element is greater: update

cont:
  addi	x2, x2, −1	# next element (RemainingCount−−)
  bne	x2, x0, loop_max	# repeat as long as there are elements

  # end: x3 contains the maximum
  j	end

update:
  add	x3, x5, x0	# x3 = x5 (new maximum)
  j	cont

end:
  # x3 = maximum
