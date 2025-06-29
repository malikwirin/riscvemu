# Calculate tribonacci(n) for a given n
# x1 = n (target index)
# x2 = trib(i-2)
# x3 = trib(i-3)
# x4 = trib(i-4)
# x5 = i (loop counter)
# x6 = temp

addi x1, x0, 7      # n = 7 (target index)
addi x2, x0, 1      # trib(2) = 1
addi x3, x0, 0      # trib(1) = 0
addi x4, x0, 0      # trib(0) = 0
addi x5, x0, 3      # i = 3 (start at trib(3))

loop:
  add x6, x2, x3    # temp = trib(i-2) + trib(i-3)
  add x6, x6, x4    # temp = temp + trib(i-4)
  add x4, x3, x0    # trib(i-4) = trib(i-3)
  add x3, x2, x0    # trib(i-3) = trib(i-2)
  add x2, x6, x0    # trib(i-2) = temp (new value)
  addi x5, x5, 1    # i++
  blt x5, x1, loop  # while i < n, repeat

# Result: tribonacci(n) is in x2
