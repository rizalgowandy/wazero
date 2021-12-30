package jit

import "github.com/tetratelabs/wazero/wasm/wazeroir"

// compiler is the interface of architecture-specific native code compiler,
// and this is responsible for compiling native code for all wazeroir operations.
type compiler interface {
	// emitPreamble is called before compiling any wazeroir operation.
	// This is used, for example, to initilize the reserved registers, etc.
	emitPreamble()
	// Finilizes the compilation, and returns the byte slice of native codes.
	// maxStackPointer is the max stack pointer that the target function would reach.
	compile() (code []byte, maxStackPointer uint64, err error)
	// Followings are resinposible for compiling each wazeroir operation.
	compileUnreachable()
	compileSwap(o *wazeroir.OperationSwap) error
	compileGlobalGet(o *wazeroir.OperationGlobalGet) error
	compileGlobalSet(o *wazeroir.OperationGlobalSet) error
	compileBr(o *wazeroir.OperationBr) error
	compileBrIf(o *wazeroir.OperationBrIf) error
	compileLabel(o *wazeroir.OperationLabel) error
	compileCall(o *wazeroir.OperationCall) error
	compileDrop(o *wazeroir.OperationDrop) error
	compileSelect() error
	compilePick(o *wazeroir.OperationPick) error
	compileAdd(o *wazeroir.OperationAdd) error
	compileSub(o *wazeroir.OperationSub) error
	compileLe(o *wazeroir.OperationLe) error
	compileLoad(o *wazeroir.OperationLoad) error
	compileLoad8(o *wazeroir.OperationLoad8) error
	compileLoad16(o *wazeroir.OperationLoad16) error
	compileLoad32(o *wazeroir.OperationLoad32) error
	compileStore(o *wazeroir.OperationStore) error
	compileStore8(o *wazeroir.OperationStore8) error
	compileStore16(o *wazeroir.OperationStore16) error
	compileStore32(o *wazeroir.OperationStore32) error
	compileMemoryGrow()
	compileMemorySize()
	compileConstI32(o *wazeroir.OperationConstI32) error
	compileConstI64(o *wazeroir.OperationConstI64) error
	compileConstF32(o *wazeroir.OperationConstF32) error
	compileConstF64(o *wazeroir.OperationConstF64) error
}
