# Test optimization patterns for the peephole optimizer
# This file tests both zero register optimization and redundant move elimination

.text

# Pattern 1: Zero register operations (should be optimized)
test_zero:
    add  t0, t1, zero    # Should become addi t0, t1, 0 (mv t0, t1)
    add  t2, x0, t3      # Should become addi t2, t3, 0 (mv t2, t3)
    or   t4, t5, zero    # Should become addi t4, t5, 0 (mv t4, t5)
    or   t6, x0, a0      # Should become addi t6, a0, 0 (mv t6, a0)

# Pattern 2: Redundant moves (should be optimized)
test_redundant_move:
    mv   t0, t1          # First move: addi t0, t1, 0
    mv   t2, t0          # Should combine to: addi t2, t1, 0

    addi a0, a1, 0       # First move (explicit)
    addi a2, a0, 0       # Should combine to: addi a2, a1, 0

# Pattern 3: No optimization (keep as-is)
test_no_opt:
    add  t0, t1, t2      # Normal add, keep as-is
    mv   t3, t0          # Keep (t0 actually used above)

    mv   a0, a1          # First move
    add  a2, a0, a3      # Keep both (a0 used here)

# Pattern 4: Optimization boundary - label
    mv   s0, s1          # First move
loop:                    # Label boundary - should NOT optimize across this
    mv   s2, s0          # Keep separate (control flow target)

# Pattern 5: Optimization boundary - directive
    mv   s3, s4          # First move
    .align 4             # Directive boundary
    mv   s5, s3          # Keep separate
