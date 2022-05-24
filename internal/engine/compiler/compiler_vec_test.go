package compiler

import (
	"encoding/binary"
	"github.com/tetratelabs/wazero/internal/testing/require"
	"github.com/tetratelabs/wazero/internal/wasm"
	"github.com/tetratelabs/wazero/internal/wazeroir"
	"testing"
)

func TestCompiler_compileV128Add(t *testing.T) {}
func TestCompiler_compileV128Sub(t *testing.T) {}

func TestCompiler_compileV128Load(t *testing.T) {
	tests := []struct {
		name       string
		memSetupFn func(buf []byte)
		loadType   wazeroir.LoadV128Type
		offset     uint32
		exp        [16]byte
	}{
		{
			name: "v128 offset=0", loadType: wazeroir.LoadV128Type128, offset: 0,
			memSetupFn: func(buf []byte) {
				copy(buf, []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20})
			},
			exp: [16]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16},
		},
		{
			name: "v128 offset=2", loadType: wazeroir.LoadV128Type128, offset: 2,
			memSetupFn: func(buf []byte) {
				copy(buf, []byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20})
			},
			exp: [16]byte{3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18},
		},
		{
			name: "8x8s offset=0", loadType: wazeroir.LoadV128Type8x8s, offset: 0,
			memSetupFn: func(buf []byte) {
				copy(buf, []byte{
					1, 0xff, 3, 0xff, 5, 0xff, 7, 0xff, 9, 10,
					11, 12, 13, 14, 15, 16, 17, 18, 19, 20,
				})
			},
			exp: [16]byte{
				1, 0, 0xff, 0xff, 3, 0, 0xff, 0xff, 5, 0, 0xff, 0xff, 7, 0, 0xff, 0xff,
			},
		},
		{
			name: "8x8s offset=3", loadType: wazeroir.LoadV128Type8x8s, offset: 3,
			memSetupFn: func(buf []byte) {
				copy(buf, []byte{
					1, 0xff, 3, 0xff, 5, 0xff, 7, 0xff, 9, 10,
					11, 12, 13, 14, 15, 16, 17, 18, 19, 20,
				})
			},
			exp: [16]byte{
				0xff, 0xff, 5, 0, 0xff, 0xff, 7, 0, 0xff, 0xff, 9, 0, 10, 0, 11, 0,
			},
		},
		{
			name: "8x8u offset=0", loadType: wazeroir.LoadV128Type8x8u, offset: 0,
			memSetupFn: func(buf []byte) {
				copy(buf, []byte{
					1, 0xff, 3, 0xff, 5, 0xff, 7, 0xff, 9, 10,
					11, 12, 13, 14, 15, 16, 17, 18, 19, 20,
				})
			},
			exp: [16]byte{
				1, 0, 0xff, 0, 3, 0, 0xff, 0, 5, 0, 0xff, 0, 7, 0, 0xff, 0,
			},
		},
		{
			name: "8x8i offset=3", loadType: wazeroir.LoadV128Type8x8u, offset: 3,
			memSetupFn: func(buf []byte) {
				copy(buf, []byte{
					1, 0xff, 3, 0xff, 5, 0xff, 7, 0xff, 9, 10,
					11, 12, 13, 14, 15, 16, 17, 18, 19, 20,
				})
			},
			exp: [16]byte{
				0xff, 0, 5, 0, 0xff, 0, 7, 0, 0xff, 0, 9, 0, 10, 0, 11, 0,
			},
		},
		{
			name: "16x4s offset=0", loadType: wazeroir.LoadV128Type16x4s, offset: 0,
			memSetupFn: func(buf []byte) {
				copy(buf, []byte{
					1, 0xff, 3, 0xff, 5, 0xff, 7, 0xff, 9, 10,
					11, 12, 13, 14, 15, 16, 17, 18, 19, 20,
				})
			},
			exp: [16]byte{
				1, 0xff, 0xff, 0xff,
				3, 0xff, 0xff, 0xff,
				5, 0xff, 0xff, 0xff,
				7, 0xff, 0xff, 0xff,
			},
		},
		{
			name: "16x4s offset=3", loadType: wazeroir.LoadV128Type16x4s, offset: 3,
			memSetupFn: func(buf []byte) {
				copy(buf, []byte{
					1, 0xff, 3, 0xff, 5, 6, 0xff, 0xff, 9, 10,
					11, 12, 13, 14, 15, 16, 17, 18, 19, 20,
				})
			},
			exp: [16]byte{
				0xff, 5, 0, 0,
				6, 0xff, 0xff, 0xff,
				0xff, 9, 0, 0,
				10, 11, 0, 0,
			},
		},
		{
			name: "16x4u offset=0", loadType: wazeroir.LoadV128Type16x4u, offset: 0,
			memSetupFn: func(buf []byte) {
				copy(buf, []byte{
					1, 0xff, 3, 0xff, 5, 0xff, 7, 0xff, 9, 10,
					11, 12, 13, 14, 15, 16, 17, 18, 19, 20,
				})
			},
			exp: [16]byte{
				1, 0xff, 0, 0,
				3, 0xff, 0, 0,
				5, 0xff, 0, 0,
				7, 0xff, 0, 0,
			},
		},
		{
			name: "16x4u offset=3", loadType: wazeroir.LoadV128Type16x4u, offset: 3,
			memSetupFn: func(buf []byte) {
				copy(buf, []byte{
					1, 0xff, 3, 0xff, 5, 6, 0xff, 0xff, 9, 10,
					11, 12, 13, 14, 15, 16, 17, 18, 19, 20,
				})
			},
			exp: [16]byte{
				0xff, 5, 0, 0,
				6, 0xff, 0, 0,
				0xff, 9, 0, 0,
				10, 11, 0, 0,
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			env := newCompilerEnvironment()
			tc.memSetupFn(env.memory())

			compiler := env.requireNewCompiler(t, newCompiler,
				&wazeroir.CompilationResult{HasMemory: true, Signature: &wasm.FunctionType{}})

			err := compiler.compilePreamble()
			require.NoError(t, err)

			err = compiler.compileConstI32(&wazeroir.OperationConstI32{Value: tc.offset})
			require.NoError(t, err)

			err = compiler.compileV128Load(&wazeroir.OperationV128Load{
				Type: tc.loadType, Arg: &wazeroir.MemoryArg{},
			})
			require.NoError(t, err)

			require.Equal(t, uint64(2), compiler.runtimeValueLocationStack().sp)
			require.Equal(t, 1, len(compiler.runtimeValueLocationStack().usedRegisters))
			loadedLocation := compiler.runtimeValueLocationStack().peek()
			require.True(t, loadedLocation.onRegister())

			err = compiler.compileReturnFunction()
			require.NoError(t, err)

			// Generate and run the code under test.
			code, _, _, err := compiler.compile()
			require.NoError(t, err)
			env.exec(code)

			require.Equal(t, uint64(2), env.stackPointer())
			lo, hi := env.stackTopAsV128()

			var actual [16]byte
			binary.LittleEndian.PutUint64(actual[:8], lo)
			binary.LittleEndian.PutUint64(actual[8:], hi)
			require.Equal(t, tc.exp, actual)
		})
	}
}
