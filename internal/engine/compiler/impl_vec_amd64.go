package compiler

import (
	"github.com/tetratelabs/wazero/internal/asm"
	"github.com/tetratelabs/wazero/internal/asm/amd64"
	"github.com/tetratelabs/wazero/internal/wazeroir"
)

// compileV128Const implements compiler.compileV128Const for amd64 architecture.
func (c *amd64Compiler) compileV128Const(o *wazeroir.OperationV128Const) error {
	c.maybeCompileMoveTopConditionalToFreeGeneralPurposeRegister()

	result, err := c.allocateRegister(registerTypeVector)
	if err != nil {
		return err
	}

	// We cannot directly load the value from memory to float regs,
	// so we move it to int reg temporarily.
	tmpReg, err := c.allocateRegister(registerTypeGeneralPurpose)
	if err != nil {
		return err
	}

	// Move the lower 64-bits.
	if o.Lo == 0 {
		c.assembler.CompileRegisterToRegister(amd64.XORQ, tmpReg, tmpReg)
	} else {
		c.assembler.CompileConstToRegister(amd64.MOVQ, int64(o.Lo), tmpReg)
	}
	c.assembler.CompileRegisterToRegister(amd64.MOVQ, tmpReg, result)

	if o.Lo != 0 && o.Hi == 0 {
		c.assembler.CompileRegisterToRegister(amd64.XORQ, tmpReg, tmpReg)
	} else if o.Hi != 0 {
		c.assembler.CompileConstToRegister(amd64.MOVQ, int64(o.Hi), tmpReg)
	}
	// Move the higher 64-bits with PINSRQ at the second element of 64x2 vector.
	c.assembler.CompileRegisterToRegisterWithArg(amd64.PINSRQ, tmpReg, result, 1)

	c.pushVectorRuntimeValueLocationOnRegister(result)
	return nil
}

// compileV128Add implements compiler.compileV128Add for amd64 architecture.
func (c *amd64Compiler) compileV128Add(o *wazeroir.OperationV128Add) error {
	c.locationStack.pop() // skip higher 64-bits.
	x2 := c.locationStack.pop()
	if err := c.compileEnsureOnGeneralPurposeRegister(x2); err != nil {
		return err
	}

	c.locationStack.pop() // skip higher 64-bits.
	x1 := c.locationStack.pop()
	if err := c.compileEnsureOnGeneralPurposeRegister(x1); err != nil {
		return err
	}
	var inst asm.Instruction
	switch o.Shape {
	case wazeroir.ShapeI8x16:
		inst = amd64.PADDB
	case wazeroir.ShapeI16x8:
		inst = amd64.PADDW
	case wazeroir.ShapeI32x4:
		inst = amd64.PADDL
	case wazeroir.ShapeI64x2:
		inst = amd64.PADDQ
	case wazeroir.ShapeF32x4:
		inst = amd64.ADDPS
	case wazeroir.ShapeF64x2:
		inst = amd64.ADDPD
	}
	c.assembler.CompileRegisterToRegister(inst, x2.register, x1.register)

	c.pushVectorRuntimeValueLocationOnRegister(x1.register)
	c.locationStack.markRegisterUnused(x2.register)
	return nil
}

func (c *amd64Compiler) compileV128Sub(o *wazeroir.OperationV128Sub) error {
	c.locationStack.pop() // skip higher 64-bits.
	x2 := c.locationStack.pop()
	if err := c.compileEnsureOnGeneralPurposeRegister(x2); err != nil {
		return err
	}

	c.locationStack.pop() // skip higher 64-bits.
	x1 := c.locationStack.pop()
	if err := c.compileEnsureOnGeneralPurposeRegister(x1); err != nil {
		return err
	}
	var inst asm.Instruction
	switch o.Shape {
	case wazeroir.ShapeI8x16:
		inst = amd64.PSUBB
	case wazeroir.ShapeI16x8:
		inst = amd64.PSUBW
	case wazeroir.ShapeI32x4:
		inst = amd64.PSUBL
	case wazeroir.ShapeI64x2:
		inst = amd64.PSUBQ
	case wazeroir.ShapeF32x4:
		inst = amd64.SUBPS
	case wazeroir.ShapeF64x2:
		inst = amd64.SUBPD
	}
	c.assembler.CompileRegisterToRegister(inst, x2.register, x1.register)

	c.pushVectorRuntimeValueLocationOnRegister(x1.register)
	c.locationStack.markRegisterUnused(x2.register)
	return nil
}

func (c *amd64Compiler) compileV128Load(o *wazeroir.OperationV128Load) error {

	switch o.Type {
	case wazeroir.LoadV128Type128:
		// MOVDQU
	case wazeroir.LoadV128Type8x8s:
		// PMOVSXBW
	case wazeroir.LoadV128Type8x8u:
		// PMOVZXBW
	case wazeroir.LoadV128Type16x4s:
		// PMOVSXWD
	case wazeroir.LoadV128Type16x4u:
		// PMOVZXWD
	case wazeroir.LoadV128Type32x2s:
		// PMOVSXDQ
	case wazeroir.LoadV128Type32x2u:
		// PMOVZXDQ
	case wazeroir.LoadV128Type8Splat:
		// https://stackoverflow.com/questions/36191748/difference-between-load1-and-broadcast-intrinsics
		// pinsrb	$0, %ecx, %xmm0
		// pxor	%xmm13, %xmm13
		// pshufb	%xmm13, %xmm0
	case wazeroir.LoadV128Type16Splat:
		// pinsrw $0, %ecx, %xmm0
		// pinsrw $1, %ecx, %xmm0
		// pshufd $0, %xmm0, %xmm0        # xmm0 = xmm0[0,0,0,0]
	case wazeroir.LoadV128Type32Splat:
		// pinsrd $0, (%rcx,%rax), %xmm0
		// pshufd $0, %xmm0, %xmm0        # xmm0 = xmm0[0,0,0,0]
	case wazeroir.LoadV128Type64Splat:
	// insrq	$0, (%rcx,%rax), %xmm0
	// pinsrq	$1, (%rcx,%rax), %xmm0
	case wazeroir.LoadV128Type32zero:
		// MOVL
	case wazeroir.LoadV128Type64zero:
		// MOVQ
	}
	return nil
}

func (c *amd64Compiler) compileV128LoadLane(o *wazeroir.OperationV128LoadLane) error {
	return nil
}

func (c *amd64Compiler) compileV128Store(o *wazeroir.OperationV128Store) error {
	return nil
}

func (c *amd64Compiler) compileV128StoreLane(o *wazeroir.OperationV128StoreLane) error {
	return nil
}

func (c *amd64Compiler) compileV128ExtractLane(o *wazeroir.OperationV128ExtractLane) error {
	return nil
}

func (c *amd64Compiler) compileV128ReplaceLane(o *wazeroir.OperationV128ReplaceLane) error {
	return nil
}

func (c *amd64Compiler) compileV128Splat(o *wazeroir.OperationV128Splat) error {
	return nil
}

func (c *amd64Compiler) compileV128Shuffle(o *wazeroir.OperationV128Shuffle) error {
	return nil
}

func (c *amd64Compiler) compileV128Swizzle(o *wazeroir.OperationV128Swizzle) error {
	return nil
}

func (c *amd64Compiler) compileV128AnyTrue(o *wazeroir.OperationV128AnyTrue) error {
	return nil
}

func (c *amd64Compiler) compileV128AllTrue(o *wazeroir.OperationV128AllTrue) error {
	return nil
}
