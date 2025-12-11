// Package tests is the entrypoint to my specific tests.
package tests

import (
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

	fmt.Println("Testing matrix: ")
	testWithTempOkMessages(testMatrix, t)
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
		foundationtesting.Assert(matrixPtr != nil, "matrix pointer should not be nil", "matrix pointer valid", t)

		rows := memstruct.MatrixRowsGet[int](matrixMark)
		cols := memstruct.MatrixColsGet[int](matrixMark)
		capacity := memstruct.MatrixCapacityGet[int](matrixMark)

		foundationtesting.Assert(rows == 5, "rows should be 5", "rows correct", t)
		foundationtesting.Assert(cols == 7, "cols should be 7", "cols correct", t)
		foundationtesting.Assert(capacity == 35, "capacity should be 35 (5*7)", "capacity correct", t)
	})

	// Test 2: Get/Set operations
	t.Run("get_set", func(t *testing.T) {
		matrixMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 3, 4)

		// Test SetAt and GetAt
		err := memstruct.MatrixSetAt(matrixMark, 1, 2, 42)
		foundationtesting.Assert(err == nil, "SetAt should succeed", "SetAt succeeded", t)

		val, err := memstruct.MatrixItemGetAt[int](matrixMark, 1, 2)
		foundationtesting.Assert(err == nil, "GetAt should succeed", "GetAt succeeded", t)
		foundationtesting.Assert(val == 42, "value should be 42", "value correct", t)

		// Test SetAtUnsafe and GetAtUnsafe
		memstruct.MatrixSetAtUnsafe(matrixMark, 0, 0, 100)
		val = memstruct.MatrixItemGetAtUnsafe[int](matrixMark, 0, 0)
		foundationtesting.Assert(val == 100, "unsafe value should be 100", "unsafe value correct", t)

		// Test bounds checking
		_, err = memstruct.MatrixItemGetAt[int](matrixMark, 10, 10)
		foundationtesting.Assert(err != nil, "GetAt should fail for out of bounds", "bounds check works", t)

		err = memstruct.MatrixSetAt(matrixMark, 10, 10, 99)
		foundationtesting.Assert(err != nil, "SetAt should fail for out of bounds", "bounds check works", t)
	})

	// Test 3: SetAll and ZeroAll
	t.Run("set_all_zero_all", func(t *testing.T) {
		matrixMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 3, 3)

		// Set all to 5
		memstruct.MatrixSetAll(matrixMark, 5)
		val := memstruct.MatrixItemGetAtUnsafe[int](matrixMark, 1, 1)
		foundationtesting.Assert(val == 5, "all values should be 5", "SetAll works", t)

		// Zero all
		memstruct.MatrixZeroAll[int](matrixMark)
		val = memstruct.MatrixItemGetAtUnsafe[int](matrixMark, 1, 1)
		foundationtesting.Assert(val == 0, "all values should be 0", "ZeroAll works", t)
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
		foundationtesting.Assert(err == nil, "CopyFrom should succeed", "CopyFrom succeeded", t)

		// Verify copied values
		val := memstruct.MatrixItemGetAtUnsafe[int](dstMark, 1, 1)
		foundationtesting.Assert(val == 0, "copied value at (1,1) should be 0", "copy correct", t)

		val = memstruct.MatrixItemGetAtUnsafe[int](dstMark, 2, 3)
		foundationtesting.Assert(val == 5, "copied value at (2,3) should be 5", "copy correct", t)
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
		foundationtesting.Assert(err == nil, "CopyFromRange should succeed", "CopyFromRange succeeded", t)

		// Verify: src[1,1] = 5 should be at dst[0,0]
		val := memstruct.MatrixItemGetAtUnsafe[int](dstMark, 0, 0)
		foundationtesting.Assert(val == 5, "copied value should be 5", "copy range correct", t)
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
		foundationtesting.Assert(val == 99, "snapshot should contain 99", "snapshot correct", t)

		// Modify source, verify snapshot unchanged
		memstruct.MatrixSetAtUnsafe(srcMark, 1, 1, 0)
		val = memstruct.MatrixItemGetAtUnsafe[int](dstMark, 1, 1)
		foundationtesting.Assert(val == 99, "snapshot should still be 99", "snapshot independent", t)
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
		foundationtesting.Assert(err == nil, "SnapshotRestore should succeed", "SnapshotRestore succeeded", t)

		val := memstruct.MatrixItemGetAtUnsafe[int](dstMark, 0, 0)
		foundationtesting.Assert(val == 50, "restored value should be 50", "restore correct", t)

		val = memstruct.MatrixItemGetAtUnsafe[int](dstMark, 1, 1)
		foundationtesting.Assert(val == 75, "restored value should be 75", "restore correct", t)
	})

	// Test 8: Clear
	t.Run("clear", func(t *testing.T) {
		matrixMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 2, 2)

		// Fill with values
		memstruct.MatrixSetAll(matrixMark, 42)
		val := memstruct.MatrixItemGetAtUnsafe[int](matrixMark, 0, 0)
		foundationtesting.Assert(val == 42, "value should be 42", "set works", t)

		// Clear
		memstruct.MatrixClear[int](matrixMark)
		val = memstruct.MatrixItemGetAtUnsafe[int](matrixMark, 0, 0)
		foundationtesting.Assert(val == 0, "value should be 0 after clear", "clear works", t)
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
		foundationtesting.Assert(count == 9, "ForEach should visit 9 elements", "ForEach count correct", t)
	})

	// Test 10: ReplaceInternal
	t.Run("replace_internal", func(t *testing.T) {
		matrixMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 2, 2)

		memstruct.MatrixSetAtUnsafe(matrixMark, 0, 0, 10)
		memstruct.MatrixSetAtUnsafe(matrixMark, 1, 1, 20)

		err := memstruct.MatrixReplaceInternal[int](matrixMark, 0, 0, 1, 1)
		foundationtesting.Assert(err == nil, "ReplaceInternal should succeed", "ReplaceInternal succeeded", t)

		val := memstruct.MatrixItemGetAtUnsafe[int](matrixMark, 1, 1)
		foundationtesting.Assert(val == 10, "replaced value should be 10", "replace correct", t)
	})

	// Test 11: IsIdxValid
	t.Run("is_idx_valid", func(t *testing.T) {
		matrixMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 3, 4)

		valid := memstruct.MatrixIsIdxValid[int](matrixMark, 0, 0)
		foundationtesting.Assert(valid, "(0,0) should be valid", "valid index works", t)

		valid = memstruct.MatrixIsIdxValid[int](matrixMark, 2, 3)
		foundationtesting.Assert(valid, "(2,3) should be valid", "valid index works", t)

		valid = memstruct.MatrixIsIdxValid[int](matrixMark, 3, 0)
		foundationtesting.Assert(!valid, "(3,0) should be invalid", "invalid index detected", t)

		valid = memstruct.MatrixIsIdxValid[int](matrixMark, 0, 4)
		foundationtesting.Assert(!valid, "(0,4) should be invalid", "invalid index detected", t)
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
		foundationtesting.Assert(rows == 2, "view should have 2 rows", "view rows correct", t)
		foundationtesting.Assert(cols == 2, "view should have 2 cols", "view cols correct", t)

		// Access view element (relative to view)
		val, err := memstruct.MatrixViewItemGetAt(view, 0, 0)
		foundationtesting.Assert(err == nil, "ViewItemGetAt should succeed", "view get succeeded", t)
		foundationtesting.Assert(val == 6, "view[0,0] should be 6 (matrix[1,1])", "view value correct", t)

		// Modify through view
		err = memstruct.MatrixViewItemSetAt(view, 0, 0, 999)
		foundationtesting.Assert(err == nil, "ViewItemSetAt should succeed", "view set succeeded", t)

		val = memstruct.MatrixItemGetAtUnsafe[int](matrixMark, 1, 1)
		foundationtesting.Assert(val == 999, "matrix[1,1] should be 999", "view modification works", t)
	})

	// Test 13: Readonly view
	t.Run("readonly_view", func(t *testing.T) {
		matrixMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 3, 3)
		memstruct.MatrixSetAtUnsafe(matrixMark, 1, 1, 42)

		view := memstruct.MatrixViewGet[int](matrixMark, 0, 2, 0, 2, true)

		readonly := memstruct.MatrixViewIsReadonly(view)
		foundationtesting.Assert(readonly, "view should be readonly", "readonly flag correct", t)

		// Should be able to read
		val, err := memstruct.MatrixViewItemGetAt(view, 1, 1)
		foundationtesting.Assert(err == nil, "readonly view should allow read", "readonly read works", t)
		foundationtesting.Assert(val == 42, "readonly view value should be 42", "readonly value correct", t)

		// Should not be able to get pointer
		_, err = memstruct.MatrixViewItemPtrGetAt(view, 1, 1)
		foundationtesting.Assert(err != nil, "readonly view should not allow pointer", "readonly pointer blocked", t)

		// Should not be able to set
		err = memstruct.MatrixViewItemSetAt(view, 1, 1, 99)
		foundationtesting.Assert(err != nil, "readonly view should not allow set", "readonly set blocked", t)
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
		foundationtesting.Assert(err == nil, "nested view get should succeed", "nested view works", t)
		foundationtesting.Assert(val == 6, "nested view[0,0] should be 6", "nested view correct", t)
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
		foundationtesting.Assert(val == 10, "unary result should be 10", "unary execute works", t)
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
		foundationtesting.Assert(val == 7, "binary result should be 7", "binary execute works", t)
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

		foundationtesting.Assert(sum == 32, "sum should be 32 (4 elements * 8)", "binary readonly works", t)
	})

	// Test 18: UnaryReadOnlyExecute
	t.Run("unary_readonly_execute", func(t *testing.T) {
		matrixMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 3, 3)
		memstruct.MatrixSetAll(matrixMark, 7)

		count := 0
		memstruct.MatrixUnaryReadOnlyExecute(matrixMark, func(item int) {
			count++
		}, 1)

		foundationtesting.Assert(count == 9, "should visit 9 elements", "unary readonly works", t)
	})

	// Test 19: Different numeric types
	t.Run("different_types", func(t *testing.T) {
		// Test float32
		f32Mark, _ := memarch.MemArchMatrixCreate[float32](allocFn, 2, 2)
		memstruct.MatrixSetAtUnsafe[float32](f32Mark, 0, 0, 3.14)
		val := memstruct.MatrixItemGetAtUnsafe[float32](f32Mark, 0, 0)
		foundationtesting.Assert(val == 3.14, "float32 value should be 3.14", "float32 works", t)

		// Test int64
		i64Mark, _ := memarch.MemArchMatrixCreate[int64](allocFn, 2, 2)
		memstruct.MatrixSetAtUnsafe[int64](i64Mark, 1, 1, 123456789)
		val64 := memstruct.MatrixItemGetAtUnsafe[int64](i64Mark, 1, 1)
		foundationtesting.Assert(val64 == 123456789, "int64 value should be 123456789", "int64 works", t)
	})

	// Test 20: Edge cases - 1x1 matrix
	t.Run("edge_case_1x1", func(t *testing.T) {
		matrixMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 1, 1)

		memstruct.MatrixSetAtUnsafe(matrixMark, 0, 0, 42)
		val := memstruct.MatrixItemGetAtUnsafe[int](matrixMark, 0, 0)
		foundationtesting.Assert(val == 42, "1x1 matrix should work", "1x1 matrix works", t)
	})

	// Test 21: Edge cases - single row
	t.Run("edge_case_single_row", func(t *testing.T) {
		matrixMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 1, 5)

		for col := uint64(0); col < 5; col++ {
			memstruct.MatrixSetAtUnsafe(matrixMark, 0, col, int(col))
		}

		val := memstruct.MatrixItemGetAtUnsafe[int](matrixMark, 0, 3)
		foundationtesting.Assert(val == 3, "single row matrix should work", "single row works", t)
	})

	// Test 22: Edge cases - single column
	t.Run("edge_case_single_col", func(t *testing.T) {
		matrixMark, _ := memarch.MemArchMatrixCreate[int](allocFn, 5, 1)

		for row := uint64(0); row < 5; row++ {
			memstruct.MatrixSetAtUnsafe(matrixMark, row, 0, int(row))
		}

		val := memstruct.MatrixItemGetAtUnsafe[int](matrixMark, 3, 0)
		foundationtesting.Assert(val == 3, "single column matrix should work", "single column works", t)
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
		foundationtesting.Assert(val == 5, "CreateFrom should copy values", "CreateFrom works", t)
	})
}

func testWithTempOkMessages(test func(t *testing.T), t *testing.T) {
	foundationtesting.EnableOkMessagesSet(true)
	test(t)
	foundationtesting.EnableOkMessagesSet(false)
}
