; Calculates the Fibonacci sequence until it overflows.

    LD      r0, 0x00        ; Initialize registers
    LD      r2, r0
    LD      r0, 0x01
loop:
    LD      r1, r0
    ADD     r0, r2          ; r0 is an implied operand, it can be omitted.
    JC      halt            ; If there was overflow, stop the loop

    DSPLY   r0              ; Display the next number in the sequence

    LD      r2, r1
    JP      loop            ; Continue the loop
halt:
    HLT                     ; Halt the processor