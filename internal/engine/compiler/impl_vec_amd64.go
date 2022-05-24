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

// compileV128Sub implements compiler.compileV128Sub for amd64 architecture.
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

// compileV128Load implements compiler.compileV128Load for amd64 architecture.
func (c *amd64Compiler) compileV128Load(o *wazeroir.OperationV128Load) error {
	result, err := c.allocateRegister(registerTypeVector)
	if err != nil {
		return err
	}

	switch o.Type {
	case wazeroir.LoadV128Type128:
		err = c.compileV128LoadImpl(amd64.MOVDQU, o.Arg.Offset, 16, result)
	case wazeroir.LoadV128Type8x8s:
		err = c.compileV128LoadImpl(amd64.PMOVSXBW, o.Arg.Offset, 8, result)
	case wazeroir.LoadV128Type8x8u:
		err = c.compileV128LoadImpl(amd64.PMOVZXBW, o.Arg.Offset, 8, result)
	case wazeroir.LoadV128Type16x4s:
		err = c.compileV128LoadImpl(amd64.PMOVSXWD, o.Arg.Offset, 8, result)
	case wazeroir.LoadV128Type16x4u:
		err = c.compileV128LoadImpl(amd64.PMOVZXWD, o.Arg.Offset, 8, result)
	case wazeroir.LoadV128Type32x2s:
		err = c.compileV128LoadImpl(amd64.PMOVSXDQ, o.Arg.Offset, 8, result)
	case wazeroir.LoadV128Type32x2u:
		err = c.compileV128LoadImpl(amd64.PMOVZXDQ, o.Arg.Offset, 8, result)
	case wazeroir.LoadV128Type8Splat:
		reg, err := c.compileMemoryAccessCeilSetup(o.Arg.Offset, 1)
		if err != nil {
			return err
		}
		c.assembler.CompileMemoryWithIndexToRegister(amd64.MOVBQZX, amd64ReservedRegisterForMemory, -1,
			reg, 1, reg)
		// pinsrb   $0, reg, result
		// pxor	    tmpVReg, tmpVReg
		// pshufb   tmpVReg, result
		c.locationStack.markRegisterUsed(result)
		tmpVReg, err := c.allocateRegister(registerTypeVector)
		if err != nil {
			return err
		}
		c.assembler.CompileRegisterToRegisterWithArg(amd64.PINSRB, reg, result, 0)
		c.assembler.CompileRegisterToRegister(amd64.PXOR, tmpVReg, tmpVReg)
		c.assembler.CompileRegisterToRegister(amd64.PSHUFB, tmpVReg, result)
	case wazeroir.LoadV128Type16Splat:
		reg, err := c.compileMemoryAccessCeilSetup(o.Arg.Offset, 2)
		if err != nil {
			return err
		}
		c.assembler.CompileMemoryWithIndexToRegister(amd64.MOVWQZX, amd64ReservedRegisterForMemory, -2,
			reg, 1, reg)
		// pinsrw $0, reg, result
		// pinsrw $1, reg, result
		// pshufd $0, result, result (result = result[0,0,0,0])
		c.assembler.CompileRegisterToRegisterWithArg(amd64.PINSRW, reg, result, 0)
		c.assembler.CompileRegisterToRegisterWithArg(amd64.PINSRW, reg, result, 1)
		c.assembler.CompileRegisterToRegisterWithArg(amd64.PSHUFD, result, result, 0)
	case wazeroir.LoadV128Type32Splat:
		reg, err := c.compileMemoryAccessCeilSetup(o.Arg.Offset, 4)
		if err != nil {
			return err
		}
		c.assembler.CompileMemoryWithIndexToRegister(amd64.MOVLQZX, amd64ReservedRegisterForMemory, -4,
			reg, 1, reg)
		// pinsrd $0, reg, result
		// pshufd $0, result, result (result = result[0,0,0,0])
		c.assembler.CompileRegisterToRegisterWithArg(amd64.PINSRD, reg, result, 0)
		c.assembler.CompileRegisterToRegisterWithArg(amd64.PSHUFD, result, result, 0)
	case wazeroir.LoadV128Type64Splat:
		reg, err := c.compileMemoryAccessCeilSetup(o.Arg.Offset, 8)
		if err != nil {
			return err
		}
		c.assembler.CompileMemoryWithIndexToRegister(amd64.MOVQ, amd64ReservedRegisterForMemory, -8,
			reg, 1, reg)
		// pinsrq $0, reg, result
		// pinsrq $1, reg, result
		c.assembler.CompileRegisterToRegisterWithArg(amd64.PINSRQ, reg, result, 0)
		c.assembler.CompileRegisterToRegisterWithArg(amd64.PINSRQ, reg, result, 1)
	case wazeroir.LoadV128Type32zero:
		err = c.compileV128LoadImpl(amd64.MOVL, o.Arg.Offset, 4, result)
	case wazeroir.LoadV128Type64zero:
		err = c.compileV128LoadImpl(amd64.MOVQ, o.Arg.Offset, 8, result)
	}

	if err != nil {
		return err
	}

	c.pushVectorRuntimeValueLocationOnRegister(result)
	return nil
}

func (c *amd64Compiler) compileV128LoadImpl(inst asm.Instruction, offset uint32, targetSizeInBytes int64, dst asm.Register) error {
	offsetReg, err := c.compileMemoryAccessCeilSetup(offset, targetSizeInBytes)
	if err != nil {
		return err
	}
	c.assembler.CompileMemoryWithIndexToRegister(inst, amd64ReservedRegisterForMemory, -targetSizeInBytes,
		offsetReg, 1, dst)
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
