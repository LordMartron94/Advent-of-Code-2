// Package tests is the entrypoint to my specific tests.
package tests

import (
	"blaze/elementwise"
	"blaze/scalar"
	"blaze/structure"
	"fmt"
	"memarch"
	"memcore"

	foundationtesting "foundation/testing"
	"memforge"
	"memstruct"
	"testing"
	"unsafe"
)

func TestEntryPoint(t *testing.T) {
	defer memforge.MemforgeMemoryDebug()
	foundationtesting.EnableOkMessagesSet(false)

	fmt.Println("Testing matrix: ")
	testMatrix(t)
}

func testMatrix(t *testing.T) {
	allocator := memforge.DynamicLinearAllocatorCreateFunction(uint64(memcore.KiloByte), func(currentCap, neededCap uint64) uint64 {
		newSize := max(currentCap*2, neededCap)
		if newSize > uint64(memcore.GigaByte) {
			panic("too much memory for a test")
		}

		return newSize
	})

	defer memforge.DynamicLinearAllocatorDestroy(allocator)

	allocFn := func(sizeBytes, alignment uint64) memcore.MarkRaw {
		return memforge.DynamicLinearAllocatorMallocUnsafe(allocator, sizeBytes, alignment)
	}

	// Test 1: Basic initialization and dimensions
	t.Run("initialization", func(t *testing.T) {
		matrixMark, matrixPtr := memarch.MemArchMatrixCreate[int](allocFn, 5, 7)
		assertTrue(t, matrixPtr != nil, "matrix pointer should not be nil", "matrix pointer valid")

		rows := memstruct.MatrixRowsGet[int](matrixMark)
		cols := memstruct.MatrixColsGet[int](matrixMark)
		capacity := memstruct.MatrixCapacityGet[int](matrixMark)

		assertEqual(t, rows, uint64(5), "rows correct")
		assertEqual(t, cols, uint64(7), "cols correct")
		assertEqual(t, capacity, uint64(35), "capacity correct")
	})

	// Test 2: Get/Set operations
	t.Run("get_set", func(t *testing.T) {
		matrixMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 3, 4)

		// Test SetAt and GetAt
		err := memstruct.MatrixSetAt(matrixMark, 1, 2, 42)
		assertTrue(t, err == nil, "SetAt should succeed", "SetAt succeeded")

		val, err := memstruct.MatrixItemGetAt[int](matrixMark, 1, 2)
		assertTrue(t, err == nil, "GetAt should succeed", "GetAt succeeded")
		assertEqual(t, val, 42, "value correct")

		// Test SetAtUnsafe and GetAtUnsafe
		memstruct.MatrixSetAtUnsafe(matrixMark, 0, 0, 100)
		val = memstruct.MatrixItemGetAtUnsafe[int](matrixMark, 0, 0)
		assertEqual(t, val, 100, "unsafe value correct")

		// Test bounds checking
		_, err = memstruct.MatrixItemGetAt[int](matrixMark, 10, 10)
		assertTrue(t, err != nil, "GetAt should fail for out of bounds", "bounds check works")

		err = memstruct.MatrixSetAt(matrixMark, 10, 10, 99)
		assertTrue(t, err != nil, "SetAt should fail for out of bounds", "bounds check works")
	})

	// Test 3: SetAll and ZeroAll
	t.Run("set_all_zero_all", func(t *testing.T) {
		matrixMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 3, 3)

		// Set all to 5
		memstruct.MatrixSetAll(matrixMark, 5)
		val := memstruct.MatrixItemGetAtUnsafe[int](matrixMark, 1, 1)
		assertEqual(t, val, 5, "SetAll works")

		// Zero all
		memstruct.MatrixZeroAll[int](matrixMark)
		val = memstruct.MatrixItemGetAtUnsafe[int](matrixMark, 1, 1)
		assertEqual(t, val, 0, "ZeroAll works")
	})

	// Test 4: CopyFrom
	t.Run("copy_from", func(t *testing.T) {
		srcMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 2, 3)
		dstMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 4, 5)

		// Fill source with pattern
		for row := uint64(0); row < 2; row++ {
			for col := uint64(0); col < 3; col++ {
				memstruct.MatrixSetAtUnsafe(srcMark, row, col, int(row*3+col))
			}
		}

		// Copy to destination
		err := memstruct.MatrixCopyFrom[int](dstMark, srcMark, 1, 1)
		assertTrue(t, err == nil, "CopyFrom should succeed", "CopyFrom succeeded")

		// Verify copied values
		val := memstruct.MatrixItemGetAtUnsafe[int](dstMark, 1, 1)
		assertEqual(t, val, 0, "copy correct")

		val = memstruct.MatrixItemGetAtUnsafe[int](dstMark, 2, 3)
		assertEqual(t, val, 5, "copy correct")
	})

	// Test 5: CopyFromRange
	t.Run("copy_from_range", func(t *testing.T) {
		srcMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 4, 4)
		dstMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 4, 4)

		// Fill source with pattern
		for row := uint64(0); row < 4; row++ {
			for col := uint64(0); col < 4; col++ {
				memstruct.MatrixSetAtUnsafe(srcMark, row, col, int(row*4+col))
			}
		}

		// Copy range [1:3, 1:3] to destination at [0, 0]
		err := memstruct.MatrixCopyFromRange[int](dstMark, srcMark, 1, 3, 1, 3, 0, 0)
		assertTrue(t, err == nil, "CopyFromRange should succeed", "CopyFromRange succeeded")

		// Verify: src[1,1] = 5 should be at dst[0,0]
		val := memstruct.MatrixItemGetAtUnsafe[int](dstMark, 0, 0)
		assertEqual(t, val, 5, "copy range correct")
	})

	// Test 6: Snapshot operations
	t.Run("snapshot", func(t *testing.T) {
		srcMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 3, 3)

		// Fill source
		memstruct.MatrixSetAtUnsafe(srcMark, 1, 1, 99)

		// Create snapshot
		dstMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 3, 3)
		memstruct.MatrixSnapshotCreate[int](dstMark, srcMark)

		// Verify snapshot
		val := memstruct.MatrixItemGetAtUnsafe[int](dstMark, 1, 1)
		assertEqual(t, val, 99, "snapshot correct")

		// Modify source, verify snapshot unchanged
		memstruct.MatrixSetAtUnsafe(srcMark, 1, 1, 0)
		val = memstruct.MatrixItemGetAtUnsafe[int](dstMark, 1, 1)
		assertEqual(t, val, 99, "snapshot independent")
	})

	// Test 7: SnapshotRestore
	t.Run("snapshot_restore", func(t *testing.T) {
		srcMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 2, 2)
		dstMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 2, 2)

		// Fill source
		memstruct.MatrixSetAtUnsafe(srcMark, 0, 0, 50)
		memstruct.MatrixSetAtUnsafe(srcMark, 1, 1, 75)

		// Restore to destination
		err := memstruct.MatrixSnapshotRestore[int](dstMark, srcMark)
		assertTrue(t, err == nil, "SnapshotRestore should succeed", "SnapshotRestore succeeded")

		val := memstruct.MatrixItemGetAtUnsafe[int](dstMark, 0, 0)
		assertEqual(t, val, 50, "restore correct")

		val = memstruct.MatrixItemGetAtUnsafe[int](dstMark, 1, 1)
		assertEqual(t, val, 75, "restore correct")
	})

	// Test 8: Clear
	t.Run("clear", func(t *testing.T) {
		matrixMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 2, 2)

		// Fill with values
		memstruct.MatrixSetAll(matrixMark, 42)
		val := memstruct.MatrixItemGetAtUnsafe[int](matrixMark, 0, 0)
		assertEqual(t, val, 42, "set works")

		// Clear
		memstruct.MatrixClear[int](matrixMark)
		val = memstruct.MatrixItemGetAtUnsafe[int](matrixMark, 0, 0)
		assertEqual(t, val, 0, "clear works")
	})

	// Test 9: ForEach
	t.Run("for_each", func(t *testing.T) {
		matrixMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 3, 3)

		// Fill with pattern
		for row := uint64(0); row < 3; row++ {
			for col := uint64(0); col < 3; col++ {
				memstruct.MatrixSetAtUnsafe(matrixMark, row, col, int(row*3+col))
			}
		}

		// Test ForEachUnsafe
		count := 0
		memstruct.MatrixForEachUnsafe[int](matrixMark, func(ptr unsafe.Pointer, row, col uint64) {
			count++
		})
		assertEqual(t, count, 9, "ForEach count correct")
	})

	// Test 10: ReplaceInternal
	t.Run("replace_internal", func(t *testing.T) {
		matrixMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 2, 2)

		memstruct.MatrixSetAtUnsafe(matrixMark, 0, 0, 10)
		memstruct.MatrixSetAtUnsafe(matrixMark, 1, 1, 20)

		err := memstruct.MatrixReplaceInternal[int](matrixMark, 0, 0, 1, 1)
		assertTrue(t, err == nil, "ReplaceInternal should succeed", "ReplaceInternal succeeded")

		val := memstruct.MatrixItemGetAtUnsafe[int](matrixMark, 1, 1)
		assertEqual(t, val, 10, "replace correct")
	})

	// Test 11: IsIdxValid
	t.Run("is_idx_valid", func(t *testing.T) {
		matrixMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 3, 4)

		valid := memstruct.MatrixIsIdxValid[int](matrixMark, 0, 0)
		assertTrue(t, valid, "(0,0) should be valid", "valid index works")

		valid = memstruct.MatrixIsIdxValid[int](matrixMark, 2, 3)
		assertTrue(t, valid, "(2,3) should be valid", "valid index works")

		valid = memstruct.MatrixIsIdxValid[int](matrixMark, 3, 0)
		assertFalse(t, valid, "(3,0) should be invalid", "invalid index detected")

		valid = memstruct.MatrixIsIdxValid[int](matrixMark, 0, 4)
		assertFalse(t, valid, "(0,4) should be invalid", "invalid index detected")
	})

	// Test 12: MatrixView
	t.Run("view", func(t *testing.T) {
		matrixMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 5, 5)

		// Fill with pattern
		for row := uint64(0); row < 5; row++ {
			for col := uint64(0); col < 5; col++ {
				memstruct.MatrixSetAtUnsafe(matrixMark, row, col, int(row*5+col))
			}
		}

		// Create view [1:3, 1:3]
		view := memstruct.MatrixViewGet[int](matrixMark, 1, 3, 1, 3, false)

		rows := memstruct.MatrixViewRowsGet(view)
		cols := memstruct.MatrixViewColsGet(view)
		assertEqual(t, rows, uint64(2), "view rows correct")
		assertEqual(t, cols, uint64(2), "view cols correct")

		// Access view element (relative to view)
		val, err := memstruct.MatrixViewItemGetAt(view, 0, 0)
		assertTrue(t, err == nil, "ViewItemGetAt should succeed", "view get succeeded")
		assertEqual(t, val, 6, "view value correct")

		// Modify through view
		err = memstruct.MatrixViewItemSetAt(view, 0, 0, 999)
		assertTrue(t, err == nil, "ViewItemSetAt should succeed", "view set succeeded")

		val = memstruct.MatrixItemGetAtUnsafe[int](matrixMark, 1, 1)
		assertEqual(t, val, 999, "view modification works")
	})

	// Test 13: Readonly view
	t.Run("readonly_view", func(t *testing.T) {
		matrixMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 3, 3)
		memstruct.MatrixSetAtUnsafe(matrixMark, 1, 1, 42)

		view := memstruct.MatrixViewGet[int](matrixMark, 0, 2, 0, 2, true)

		readonly := memstruct.MatrixViewIsReadonly(view)
		assertTrue(t, readonly, "view should be readonly", "readonly flag correct")

		// Should be able to read
		val, err := memstruct.MatrixViewItemGetAt(view, 1, 1)
		assertTrue(t, err == nil, "readonly view should allow read", "readonly read works")
		assertEqual(t, val, 42, "readonly value correct")

		// Should not be able to get pointer
		_, err = memstruct.MatrixViewItemPtrGetAt(view, 1, 1)
		assertTrue(t, err != nil, "readonly view should not allow pointer", "readonly pointer blocked")

		// Should not be able to set
		err = memstruct.MatrixViewItemSetAt(view, 1, 1, 99)
		assertTrue(t, err != nil, "readonly view should not allow set", "readonly set blocked")
	})

	// Test 14: Nested view
	t.Run("nested_view", func(t *testing.T) {
		matrixMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 5, 5)

		// Fill with pattern
		for row := uint64(0); row < 5; row++ {
			for col := uint64(0); col < 5; col++ {
				memstruct.MatrixSetAtUnsafe(matrixMark, row, col, int(row*5+col))
			}
		}

		// Create outer view [1:4, 1:4]
		outerView := memstruct.MatrixViewGet[int](matrixMark, 1, 4, 1, 4, false)

		// Create nested view [0:2, 0:2] relative to outer view
		nestedView := memstruct.MatrixViewNestedGet(outerView, 0, 2, 0, 2)

		val, err := memstruct.MatrixViewItemGetAt(nestedView, 0, 0)
		assertTrue(t, err == nil, "nested view get should succeed", "nested view works")
		assertEqual(t, val, 6, "nested view correct")
	})

	// Test 15: UnaryExecute
	t.Run("unary_execute", func(t *testing.T) {
		srcMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 3, 3)
		dstMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 3, 3)

		// Fill source
		memstruct.MatrixSetAll(srcMark, 5)

		// Execute unary operation: multiply by 2
		memstruct.MatrixUnaryExecute(srcMark, dstMark, func(item int) int {
			return item * 2
		}, 1)

		val := memstruct.MatrixItemGetAtUnsafe[int](dstMark, 1, 1)
		assertEqual(t, val, 10, "unary execute works")
	})

	// Test 16: BinaryExecute
	t.Run("binary_execute", func(t *testing.T) {
		aMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 2, 2)
		bMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 2, 2)
		dstMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 2, 2)

		// Fill matrices
		memstruct.MatrixSetAll(aMark, 3)
		memstruct.MatrixSetAll(bMark, 4)

		// Execute binary operation: add
		memstruct.MatrixBinaryExecute(aMark, bMark, dstMark, func(a, b int) int {
			return a + b
		}, 1)

		val := memstruct.MatrixItemGetAtUnsafe[int](dstMark, 0, 0)
		assertEqual(t, val, 7, "binary execute works")
	})

	// Test 17: BinaryReadOnlyExecute
	t.Run("binary_readonly_execute", func(t *testing.T) {
		aMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 2, 2)
		bMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 2, 2)

		memstruct.MatrixSetAll(aMark, 5)
		memstruct.MatrixSetAll(bMark, 3)

		sum := 0
		memstruct.MatrixBinaryReadOnlyExecute(aMark, bMark, func(a, b int) {
			sum += a + b
		}, 1)

		assertEqual(t, sum, 32, "binary readonly works")
	})

	// Test 18: UnaryReadOnlyExecute
	t.Run("unary_readonly_execute", func(t *testing.T) {
		matrixMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 3, 3)
		memstruct.MatrixSetAll(matrixMark, 7)

		count := 0
		memstruct.MatrixUnaryReadOnlyExecute(matrixMark, func(item int) {
			count++
		}, 1)

		assertEqual(t, count, 9, "unary readonly works")
	})

	// Test 19: Different numeric types
	t.Run("different_types", func(t *testing.T) {
		// Test float32
		f32Mark, _ := memarch.MemArchMatrixCreate[float32](allocFn, 2, 2)
		memstruct.MatrixSetAtUnsafe[float32](f32Mark, 0, 0, 3.14)
		val := memstruct.MatrixItemGetAtUnsafe[float32](f32Mark, 0, 0)
		assertEqualFloat(t, float64(val), 3.14, 0.001, "float32 works")

		// Test int64
		i64Mark, _ := memarch.MemArchMatrixCreate[int64](allocFn, 2, 2)
		memstruct.MatrixSetAtUnsafe[int64](i64Mark, 1, 1, 123456789)
		val64 := memstruct.MatrixItemGetAtUnsafe[int64](i64Mark, 1, 1)
		assertEqual(t, val64, int64(123456789), "int64 works")
	})

	// Test 20: Edge cases - 1x1 matrix
	t.Run("edge_case_1x1", func(t *testing.T) {
		matrixMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 1, 1)

		memstruct.MatrixSetAtUnsafe(matrixMark, 0, 0, 42)
		val := memstruct.MatrixItemGetAtUnsafe[int](matrixMark, 0, 0)
		assertEqual(t, val, 42, "1x1 matrix works")
	})

	// Test 21: Edge cases - single row
	t.Run("edge_case_single_row", func(t *testing.T) {
		matrixMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 1, 5)

		for col := uint64(0); col < 5; col++ {
			memstruct.MatrixSetAtUnsafe(matrixMark, 0, col, int(col))
		}

		val := memstruct.MatrixItemGetAtUnsafe[int](matrixMark, 0, 3)
		assertEqual(t, val, 3, "single row works")
	})

	// Test 22: Edge cases - single column
	t.Run("edge_case_single_col", func(t *testing.T) {
		matrixMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 5, 1)

		for row := uint64(0); row < 5; row++ {
			memstruct.MatrixSetAtUnsafe(matrixMark, row, 0, int(row))
		}

		val := memstruct.MatrixItemGetAtUnsafe[int](matrixMark, 3, 0)
		assertEqual(t, val, 3, "single column works")
	})

	// Test 23: CreateFrom
	t.Run("create_from", func(t *testing.T) {
		srcMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 3, 3)

		// Fill source
		for row := uint64(0); row < 3; row++ {
			for col := uint64(0); col < 3; col++ {
				memstruct.MatrixSetAtUnsafe(srcMark, row, col, int(row*3+col))
			}
		}

		dstMark, _ := memarch.MemArchMatrixCreateFrom[int](allocFn, srcMark, 3, 3)

		// Verify copy
		val := memstruct.MatrixItemGetAtUnsafe[int](dstMark, 1, 2)
		assertEqual(t, val, 5, "CreateFrom works")
	})

	// ────────────────────────────────────────────────────────────────
	//   BLAZE MATRIX OPERATIONS TESTS
	// ────────────────────────────────────────────────────────────────

	// Test 24: Elementwise Add
	t.Run("blaze_elementwise_add", func(t *testing.T) {
		aMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 2, 2)
		bMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 2, 2)
		cMark, _ := memarch.MemArchMatrixCreate[float64](allocFn, 2, 2)

		memstruct.MatrixSetAtUnsafe(aMark, 0, 0, 1)
		memstruct.MatrixSetAtUnsafe(aMark, 0, 1, 2)
		memstruct.MatrixSetAtUnsafe(aMark, 1, 0, 3)
		memstruct.MatrixSetAtUnsafe(aMark, 1, 1, 4)

		memstruct.MatrixSetAtUnsafe(bMark, 0, 0, 5)
		memstruct.MatrixSetAtUnsafe(bMark, 0, 1, 6)
		memstruct.MatrixSetAtUnsafe(bMark, 1, 0, 7)
		memstruct.MatrixSetAtUnsafe(bMark, 1, 1, 8)

		elementwise.BlazeElementWiseMatrixAddF64[int, int](aMark, bMark, cMark)

		val := memstruct.MatrixItemGetAtUnsafe[float64](cMark, 0, 0)
		assertEqualFloat(t, val, 6, 0.001, "elementwise add works")

		val = memstruct.MatrixItemGetAtUnsafe[float64](cMark, 1, 1)
		assertEqualFloat(t, val, 12, 0.001, "elementwise add works")
	})

	// Test 25: Elementwise Subtract
	t.Run("blaze_elementwise_subtract", func(t *testing.T) {
		aMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 2, 2)
		bMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 2, 2)
		cMark, _ := memarch.MemArchMatrixCreate[float64](allocFn, 2, 2)

		memstruct.MatrixSetAtUnsafe(aMark, 0, 0, 10)
		memstruct.MatrixSetAtUnsafe(aMark, 1, 1, 20)

		memstruct.MatrixSetAtUnsafe(bMark, 0, 0, 3)
		memstruct.MatrixSetAtUnsafe(bMark, 1, 1, 5)

		elementwise.BlazeElementWiseMatrixSubtractF64[int, int](aMark, bMark, cMark)

		val := memstruct.MatrixItemGetAtUnsafe[float64](cMark, 0, 0)
		assertEqual(t, val, 7, "elementwise subtract works")

		val = memstruct.MatrixItemGetAtUnsafe[float64](cMark, 1, 1)
		assertEqual(t, val, 15, "elementwise subtract works")
	})

	// Test 26: Elementwise Multiply (Hadamard product)
	t.Run("blaze_elementwise_multiply", func(t *testing.T) {
		aMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 2, 2)
		bMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 2, 2)
		cMark, _ := memarch.MemArchMatrixCreate[float64](allocFn, 2, 2)

		memstruct.MatrixSetAtUnsafe(aMark, 0, 0, 3)
		memstruct.MatrixSetAtUnsafe(aMark, 1, 1, 4)

		memstruct.MatrixSetAtUnsafe(bMark, 0, 0, 2)
		memstruct.MatrixSetAtUnsafe(bMark, 1, 1, 5)

		elementwise.BlazeElementWiseMatrixMultiplyF64[int, int](aMark, bMark, cMark)

		val := memstruct.MatrixItemGetAtUnsafe[float64](cMark, 0, 0)
		assertEqual(t, val, 6, "elementwise multiply works")

		val = memstruct.MatrixItemGetAtUnsafe[float64](cMark, 1, 1)
		assertEqual(t, val, 20, "elementwise multiply works")
	})

	// Test 27: Elementwise Divide
	t.Run("blaze_elementwise_divide", func(t *testing.T) {
		aMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 2, 2)
		bMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 2, 2)
		cMark, _ := memarch.MemArchMatrixCreate[float64](allocFn, 2, 2)

		memstruct.MatrixSetAtUnsafe(aMark, 0, 0, 20)
		memstruct.MatrixSetAtUnsafe(aMark, 1, 1, 15)

		memstruct.MatrixSetAtUnsafe(bMark, 0, 0, 4)
		memstruct.MatrixSetAtUnsafe(bMark, 1, 1, 3)

		elementwise.BlazeElementWiseMatrixDivideF64[int, int](aMark, bMark, cMark)

		val := memstruct.MatrixItemGetAtUnsafe[float64](cMark, 0, 0)
		assertEqualFloat(t, val, 5.0, 0.001, "elementwise divide works")

		val = memstruct.MatrixItemGetAtUnsafe[float64](cMark, 1, 1)
		assertEqualFloat(t, val, 5.0, 0.001, "elementwise divide works")
	})

	// Test 28: Scalar Multiply
	t.Run("blaze_scalar_multiply", func(t *testing.T) {
		srcMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 2, 2)
		dstMark, _ := memarch.MemArchMatrixCreate[float64](allocFn, 2, 2)

		memstruct.MatrixSetAtUnsafe(srcMark, 0, 0, 5)
		memstruct.MatrixSetAtUnsafe(srcMark, 1, 1, 10)

		scalar.BlazeScalarMatrixMultiplyF64[int](srcMark, dstMark, 2.5)

		val := memstruct.MatrixItemGetAtUnsafe[float64](dstMark, 0, 0)
		assertEqualFloat(t, val, 12.5, 0.001, "scalar multiply works")

		val = memstruct.MatrixItemGetAtUnsafe[float64](dstMark, 1, 1)
		assertEqualFloat(t, val, 25.0, 0.001, "scalar multiply works")
	})

	// Test 29: Scalar Divide
	t.Run("blaze_scalar_divide", func(t *testing.T) {
		srcMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 2, 2)
		dstMark, _ := memarch.MemArchMatrixCreate[float64](allocFn, 2, 2)

		memstruct.MatrixSetAtUnsafe(srcMark, 0, 0, 20)
		memstruct.MatrixSetAtUnsafe(srcMark, 1, 1, 30)

		scalar.BlazeScalarMatrixDivideF64[int](srcMark, dstMark, 4.0)

		val := memstruct.MatrixItemGetAtUnsafe[float64](dstMark, 0, 0)
		assertEqualFloat(t, val, 5.0, 0.001, "scalar divide works")

		val = memstruct.MatrixItemGetAtUnsafe[float64](dstMark, 1, 1)
		assertEqualFloat(t, val, 7.5, 0.001, "scalar divide works")
	})

	// Test 30: Scalar Add
	t.Run("blaze_scalar_add", func(t *testing.T) {
		srcMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 2, 2)
		dstMark, _ := memarch.MemArchMatrixCreate[float64](allocFn, 2, 2)

		memstruct.MatrixSetAtUnsafe(srcMark, 0, 0, 5)
		memstruct.MatrixSetAtUnsafe(srcMark, 1, 1, 10)

		scalar.BlazeScalarMatrixAddF64[int](srcMark, dstMark, 3.5)

		val := memstruct.MatrixItemGetAtUnsafe[float64](dstMark, 0, 0)
		assertEqualFloat(t, val, 8.5, 0.001, "scalar add works")

		val = memstruct.MatrixItemGetAtUnsafe[float64](dstMark, 1, 1)
		assertEqualFloat(t, val, 13.5, 0.001, "scalar add works")
	})

	// Test 31: Scalar Subtract
	t.Run("blaze_scalar_subtract", func(t *testing.T) {
		srcMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 2, 2)
		dstMark, _ := memarch.MemArchMatrixCreate[float64](allocFn, 2, 2)

		memstruct.MatrixSetAtUnsafe(srcMark, 0, 0, 10)
		memstruct.MatrixSetAtUnsafe(srcMark, 1, 1, 15)

		scalar.BlazeScalarMatrixSubtractF64[int](srcMark, dstMark, 2.5)

		val := memstruct.MatrixItemGetAtUnsafe[float64](dstMark, 0, 0)
		assertEqualFloat(t, val, 7.5, 0.001, "scalar subtract works")

		val = memstruct.MatrixItemGetAtUnsafe[float64](dstMark, 1, 1)
		assertEqualFloat(t, val, 12.5, 0.001, "scalar subtract works")
	})

	// Test 32: Scalar Clamp
	t.Run("blaze_scalar_clamp", func(t *testing.T) {
		matrixMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 3, 3)

		memstruct.MatrixSetAtUnsafe(matrixMark, 0, 0, 5)
		memstruct.MatrixSetAtUnsafe(matrixMark, 0, 1, 15)
		memstruct.MatrixSetAtUnsafe(matrixMark, 1, 0, -5)
		memstruct.MatrixSetAtUnsafe(matrixMark, 1, 1, 25)

		scalar.BlazeScalarMatrixClamp(matrixMark, 0, 20)

		val := memstruct.MatrixItemGetAtUnsafe[int](matrixMark, 0, 0)
		assertEqual(t, val, 5, "clamp works")

		val = memstruct.MatrixItemGetAtUnsafe[int](matrixMark, 0, 1)
		assertEqual(t, val, 15, "clamp works")

		val = memstruct.MatrixItemGetAtUnsafe[int](matrixMark, 1, 0)
		assertEqual(t, val, 0, "clamp min works")

		val = memstruct.MatrixItemGetAtUnsafe[int](matrixMark, 1, 1)
		assertEqual(t, val, 20, "clamp max works")
	})

	// Test 33: Scalar SetAllSequence
	t.Run("blaze_scalar_set_all_sequence", func(t *testing.T) {
		matrixMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 3, 3)

		scalar.BlazeScalarMatrixSetAllSequence(matrixMark, 10, 2)

		// Check first few values: 10, 12, 14, 16, ...
		val := memstruct.MatrixItemGetAtUnsafe[int](matrixMark, 0, 0)
		assertEqual(t, val, 10, "sequence works")

		val = memstruct.MatrixItemGetAtUnsafe[int](matrixMark, 0, 1)
		assertEqual(t, val, 12, "sequence works")

		val = memstruct.MatrixItemGetAtUnsafe[int](matrixMark, 0, 2)
		assertEqual(t, val, 14, "sequence works")

		val = memstruct.MatrixItemGetAtUnsafe[int](matrixMark, 1, 0)
		assertEqual(t, val, 16, "sequence works")
	})

	// Test 34: Transpose (non-square)
	t.Run("blaze_structure_transpose_non_square", func(t *testing.T) {
		srcMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 2, 3)
		dstMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 3, 2)

		// Fill source: 2x3 matrix
		// [1, 2, 3]
		// [4, 5, 6]
		memstruct.MatrixSetAtUnsafe(srcMark, 0, 0, 1)
		memstruct.MatrixSetAtUnsafe(srcMark, 0, 1, 2)
		memstruct.MatrixSetAtUnsafe(srcMark, 0, 2, 3)
		memstruct.MatrixSetAtUnsafe(srcMark, 1, 0, 4)
		memstruct.MatrixSetAtUnsafe(srcMark, 1, 1, 5)
		memstruct.MatrixSetAtUnsafe(srcMark, 1, 2, 6)

		structure.BlazeStructureMatrixTranspose[int](srcMark, dstMark)

		// Verify transpose: 3x2 matrix
		// [1, 4]
		// [2, 5]
		// [3, 6]
		val := memstruct.MatrixItemGetAtUnsafe[int](dstMark, 0, 0)
		assertEqual(t, val, 1, "transpose works")

		val = memstruct.MatrixItemGetAtUnsafe[int](dstMark, 0, 1)
		assertEqual(t, val, 4, "transpose works")

		val = memstruct.MatrixItemGetAtUnsafe[int](dstMark, 1, 0)
		assertEqual(t, val, 2, "transpose works")

		val = memstruct.MatrixItemGetAtUnsafe[int](dstMark, 2, 1)
		assertEqual(t, val, 6, "transpose works")
	})

	// Test 35: Transpose (square, in-place)
	t.Run("blaze_structure_transpose_square_inplace", func(t *testing.T) {
		matrixMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 3, 3)

		// Fill matrix:
		// [1, 2, 3]
		// [4, 5, 6]
		// [7, 8, 9]
		for row := uint64(0); row < 3; row++ {
			for col := uint64(0); col < 3; col++ {
				memstruct.MatrixSetAtUnsafe(matrixMark, row, col, int(row*3+col+1))
			}
		}

		structure.BlazeStructureMatrixTranspose[int](matrixMark, matrixMark)

		// Verify transpose:
		// [1, 4, 7]
		// [2, 5, 8]
		// [3, 6, 9]
		val := memstruct.MatrixItemGetAtUnsafe[int](matrixMark, 0, 0)
		assertEqual(t, val, 1, "in-place transpose works")

		val = memstruct.MatrixItemGetAtUnsafe[int](matrixMark, 0, 1)
		assertEqual(t, val, 4, "in-place transpose works")

		val = memstruct.MatrixItemGetAtUnsafe[int](matrixMark, 1, 0)
		assertEqual(t, val, 2, "in-place transpose works")

		val = memstruct.MatrixItemGetAtUnsafe[int](matrixMark, 2, 2)
		assertEqual(t, val, 9, "in-place transpose works")
	})

	// Test 36: Transpose (square, separate matrices)
	t.Run("blaze_structure_transpose_square_separate", func(t *testing.T) {
		srcMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 2, 2)
		dstMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 2, 2)

		// Fill source:
		// [1, 2]
		// [3, 4]
		memstruct.MatrixSetAtUnsafe(srcMark, 0, 0, 1)
		memstruct.MatrixSetAtUnsafe(srcMark, 0, 1, 2)
		memstruct.MatrixSetAtUnsafe(srcMark, 1, 0, 3)
		memstruct.MatrixSetAtUnsafe(srcMark, 1, 1, 4)

		structure.BlazeStructureMatrixTranspose[int](srcMark, dstMark)

		// Verify transpose:
		// [1, 3]
		// [2, 4]
		val := memstruct.MatrixItemGetAtUnsafe[int](dstMark, 0, 0)
		assertEqual(t, val, 1, "transpose works")

		val = memstruct.MatrixItemGetAtUnsafe[int](dstMark, 0, 1)
		assertEqual(t, val, 3, "transpose works")

		val = memstruct.MatrixItemGetAtUnsafe[int](dstMark, 1, 0)
		assertEqual(t, val, 2, "transpose works")

		// Verify source unchanged
		val = memstruct.MatrixItemGetAtUnsafe[int](srcMark, 0, 1)
		assertEqual(t, val, 2, "transpose preserves source")
	})

	// Test 37: Float32 precision operations
	t.Run("blaze_float32_precision", func(t *testing.T) {
		aMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 2, 2)
		bMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 2, 2)
		cMark, _ := memarch.MemArchMatrixCreate[float32](allocFn, 2, 2)

		memstruct.MatrixSetAtUnsafe(aMark, 0, 0, 3)
		memstruct.MatrixSetAtUnsafe(bMark, 0, 0, 2)

		elementwise.BlazeElementWiseMatrixAddF32[int, int](aMark, bMark, cMark)

		val := memstruct.MatrixItemGetAtUnsafe[float32](cMark, 0, 0)
		assertEqualFloat(t, float64(val), 5.0, 0.001, "float32 works")
	})
}

func testWithTempOkMessages(test func(t *testing.T), t *testing.T) {
	foundationtesting.EnableOkMessagesSet(true)
	test(t)
	foundationtesting.EnableOkMessagesSet(false)
}

// assertEqual is a helper that asserts two values are equal and shows expected/got on failure.
func assertEqual[T comparable](t *testing.T, got, expected T, successMsg string) {
	if got != expected {
		foundationtesting.Assert(false, fmt.Sprintf("expected %v, got %v", expected, got), successMsg, t)
	} else {
		foundationtesting.Assert(true, "", successMsg, t)
	}
}

// assertEqualFloat is a helper for float comparisons with epsilon tolerance.
func assertEqualFloat(t *testing.T, got, expected, epsilon float64, successMsg string) {
	diff := got - expected
	if diff < 0 {
		diff = -diff
	}
	if diff > epsilon {
		foundationtesting.Assert(false, fmt.Sprintf("expected %v, got %v (diff: %v)", expected, got, diff), successMsg, t)
	} else {
		foundationtesting.Assert(true, "", successMsg, t)
	}
}

// assertTrue is a helper that asserts a condition is true.
func assertTrue(t *testing.T, condition bool, errorMsg, successMsg string) {
	foundationtesting.Assert(condition, errorMsg, successMsg, t)
}

// assertFalse is a helper that asserts a condition is false.
func assertFalse(t *testing.T, condition bool, errorMsg, successMsg string) {
	foundationtesting.Assert(!condition, errorMsg, successMsg, t)
}
