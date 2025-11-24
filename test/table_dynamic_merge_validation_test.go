package test

import (
	"fmt"
	"testing"

	"github.com/ZeroHawkeye/wordZero/pkg/document"
)

// TestTableDynamicRowMergeValidation 验证动态添加行后合并单元格是否会导致错位
// 这是对issue中描述的问题的专门测试
func TestTableDynamicRowMergeValidation(t *testing.T) {
	// 开启日志
	document.SetGlobalLevel(document.LogLevelInfo)

	t.Run("验证28行表格动态添加第29行后合并单元格不会错位", func(t *testing.T) {
		doc := document.New()

		// 创建28行表格（如issue中描述的场景）
		config := &document.TableConfig{
			Rows:  28,
			Cols:  5,
			Width: 10000,
		}

		table := doc.AddTable(config)
		if table == nil {
			t.Fatal("创建表格失败")
		}

		// 填充前28行数据
		for i := 0; i < 28; i++ {
			for j := 0; j < 5; j++ {
				table.SetCellText(i, j, fmt.Sprintf("Cell-%d-%d", i+1, j+1))
			}
		}

		// 记录原始表格网格列数
		originalGridColCount := len(table.Grid.Cols)
		t.Logf("原始表格网格列数: %d", originalGridColCount)

		// 验证初始状态
		if table.GetRowCount() != 28 {
			t.Fatalf("初始行数不正确: 期望28，实际%d", table.GetRowCount())
		}
		if table.GetColumnCount() != 5 {
			t.Fatalf("初始列数不正确: 期望5，实际%d", table.GetColumnCount())
		}

		// 动态添加第29行
		err := table.AppendRow([]string{"Row29-1", "Row29-2", "Row29-3", "Row29-4", "Row29-5"})
		if err != nil {
			t.Fatalf("添加第29行失败: %v", err)
		}

		// 验证添加行后的状态
		if table.GetRowCount() != 29 {
			t.Fatalf("添加行后行数不正确: 期望29，实际%d", table.GetRowCount())
		}

		// 验证第29行的单元格数量
		row29CellCount := len(table.Rows[28].Cells)
		if row29CellCount != 5 {
			t.Errorf("第29行单元格数不正确: 期望5，实际%d", row29CellCount)
		}

		// 验证表格网格列数没有改变
		gridColCountAfterInsert := len(table.Grid.Cols)
		if gridColCountAfterInsert != originalGridColCount {
			t.Errorf("添加行后网格列数改变: 原始%d，现在%d", originalGridColCount, gridColCountAfterInsert)
		}

		// 验证所有行的单元格数量都是5
		for i, row := range table.Rows {
			if len(row.Cells) != 5 {
				t.Errorf("第%d行单元格数不正确: 期望5，实际%d", i+1, len(row.Cells))
			}
		}

		// 执行合并操作 - 这是issue中提到会导致错位的操作
		testMerges := []struct {
			name     string
			row      int
			startCol int
			endCol   int
		}{
			{"合并第1行的列2-4", 0, 1, 3},
			{"合并第15行的列1-3", 14, 0, 2},
			{"合并第29行（动态添加的）的列3-5", 28, 2, 4},
		}

		for _, tm := range testMerges {
			t.Run(tm.name, func(t *testing.T) {
				// 记录合并前该行的单元格数
				cellCountBefore := len(table.Rows[tm.row].Cells)
				
				// 执行合并
				err := table.MergeCellsHorizontal(tm.row, tm.startCol, tm.endCol)
				if err != nil {
					t.Fatalf("%s失败: %v", tm.name, err)
				}

				// 验证合并后的单元格数
				cellCountAfter := len(table.Rows[tm.row].Cells)
				expectedCellCount := cellCountBefore - (tm.endCol - tm.startCol)
				if cellCountAfter != expectedCellCount {
					t.Errorf("%s后单元格数不正确: 期望%d，实际%d", 
						tm.name, expectedCellCount, cellCountAfter)
				}

				// 验证合并的起始单元格有GridSpan属性
				mergedCell := table.Rows[tm.row].Cells[tm.startCol]
				if mergedCell.Properties == nil || mergedCell.Properties.GridSpan == nil {
					t.Errorf("%s后GridSpan未设置", tm.name)
				} else {
					expectedSpan := fmt.Sprintf("%d", tm.endCol-tm.startCol+1)
					if mergedCell.Properties.GridSpan.Val != expectedSpan {
						t.Errorf("%s后GridSpan值不正确: 期望%s，实际%s",
							tm.name, expectedSpan, mergedCell.Properties.GridSpan.Val)
					}
				}
			})
		}

		// 验证表格网格列数在所有合并操作后仍然保持不变
		finalGridColCount := len(table.Grid.Cols)
		if finalGridColCount != originalGridColCount {
			t.Errorf("合并后网格列数改变: 原始%d，现在%d", originalGridColCount, finalGridColCount)
		}

		// 验证每个单元格的宽度属性都正确设置
		for i, row := range table.Rows {
			for j, cell := range row.Cells {
				// 跳过已合并的单元格（它们有GridSpan或VMerge属性）
				if cell.Properties != nil && (cell.Properties.GridSpan != nil || cell.Properties.VMerge != nil) {
					continue
				}
				if cell.Properties == nil || cell.Properties.TableCellW == nil {
					t.Errorf("行%d列%d缺少宽度属性", i+1, j+1)
				}
			}
		}

		// 保存文档供手工验证
		err = doc.Save("test/output/dynamic_row_merge_validation.docx")
		if err != nil {
			t.Logf("警告：保存文档失败: %v（这不影响测试结果）", err)
		} else {
			t.Logf("测试文档已保存到: test/output/dynamic_row_merge_validation.docx")
		}
	})

	t.Run("对比：完整创建表格后合并单元格", func(t *testing.T) {
		doc := document.New()

		// 直接创建29行表格（不动态添加）
		config := &document.TableConfig{
			Rows:  29,
			Cols:  5,
			Width: 10000,
		}

		table := doc.AddTable(config)
		if table == nil {
			t.Fatal("创建表格失败")
		}

		// 填充所有29行数据
		for i := 0; i < 29; i++ {
			for j := 0; j < 5; j++ {
				table.SetCellText(i, j, fmt.Sprintf("Cell-%d-%d", i+1, j+1))
			}
		}

		// 记录原始网格列数
		originalGridColCount := len(table.Grid.Cols)

		// 执行相同的合并操作
		table.MergeCellsHorizontal(0, 1, 3)
		table.MergeCellsHorizontal(14, 0, 2)
		table.MergeCellsHorizontal(28, 2, 4)

		// 验证表格网格列数没有改变
		finalGridColCount := len(table.Grid.Cols)
		if finalGridColCount != originalGridColCount {
			t.Errorf("合并后网格列数改变: 原始%d，现在%d", originalGridColCount, finalGridColCount)
		}

		// 验证每个单元格的宽度属性
		for i, row := range table.Rows {
			for j, cell := range row.Cells {
				if cell.Properties == nil || cell.Properties.TableCellW == nil {
					if cell.Properties != nil && (cell.Properties.GridSpan != nil || cell.Properties.VMerge != nil) {
						continue
					}
					t.Errorf("行%d列%d缺少宽度属性", i+1, j+1)
				}
			}
		}

		// 保存文档供对比
		err := doc.Save("test/output/static_table_merge_validation.docx")
		if err != nil {
			t.Logf("警告：保存文档失败: %v", err)
		} else {
			t.Logf("对比文档已保存到: test/output/static_table_merge_validation.docx")
		}
	})

	t.Run("验证单元格宽度一致性", func(t *testing.T) {
		doc := document.New()

		// 创建带自定义列宽的表格
		customWidths := []int{1500, 2000, 2500, 2000, 2000}
		config := &document.TableConfig{
			Rows:      3,
			Cols:      5,
			Width:     10000,
			ColWidths: customWidths,
		}

		table := doc.AddTable(config)
		if table == nil {
			t.Fatal("创建表格失败")
		}

		// 记录每列的预期宽度
		expectedWidths := make([]string, len(table.Grid.Cols))
		for i, col := range table.Grid.Cols {
			expectedWidths[i] = col.W
		}

		// 动态添加多行
		for i := 0; i < 5; i++ {
			err := table.AppendRow([]string{
				fmt.Sprintf("A%d", i+4),
				fmt.Sprintf("B%d", i+4),
				fmt.Sprintf("C%d", i+4),
				fmt.Sprintf("D%d", i+4),
				fmt.Sprintf("E%d", i+4),
			})
			if err != nil {
				t.Fatalf("添加第%d行失败: %v", i+4, err)
			}
		}

		// 验证所有新添加的行的单元格宽度与网格定义一致
		for i := 3; i < table.GetRowCount(); i++ {
			row := table.Rows[i]
			for j, cell := range row.Cells {
				if cell.Properties == nil || cell.Properties.TableCellW == nil {
					t.Errorf("行%d列%d缺少宽度属性", i+1, j+1)
					continue
				}

				actualWidth := cell.Properties.TableCellW.W
				expectedWidth := expectedWidths[j]
				if actualWidth != expectedWidth {
					t.Errorf("行%d列%d宽度不一致: 期望%s，实际%s",
						i+1, j+1, expectedWidth, actualWidth)
				}
			}
		}

		// 执行合并后再次验证
		const (
			mergeRow       = 5
			mergeStartCol  = 1
			mergeEndCol    = 3
			mergedCellSpan = mergeEndCol - mergeStartCol + 1
		)
		table.MergeCellsHorizontal(mergeRow, mergeStartCol, mergeEndCol)

		// 验证未合并的单元格宽度仍然正确
		for i, row := range table.Rows {
			for j, cell := range row.Cells {
				// 跳过合并的单元格
				if cell.Properties != nil && cell.Properties.GridSpan != nil {
					continue
				}

				if cell.Properties == nil || cell.Properties.TableCellW == nil {
					t.Errorf("合并后行%d列%d缺少宽度属性", i+1, j+1)
					continue
				}

				// 对于未合并的单元格，验证宽度
				// 第5行（索引5）的列1-3已被合并，需要调整后续列的索引映射
				colIndex := j
				if i == mergeRow && j > mergeStartCol {
					// 第5行合并了列1-3，所以后续列的原始索引需要加上被合并掉的列数
					colIndex = j + (mergedCellSpan - 1)
				}
				if colIndex < len(expectedWidths) {
					expectedWidth := expectedWidths[colIndex]
					actualWidth := cell.Properties.TableCellW.W
					if actualWidth != expectedWidth {
						t.Errorf("合并后行%d列%d宽度不一致: 期望%s，实际%s",
							i+1, j+1, expectedWidth, actualWidth)
					}
				}
			}
		}
	})
}

// TestTableGridColCountAfterOperations 验证表格网格列数在各种操作后保持一致
func TestTableGridColCountAfterOperations(t *testing.T) {
	document.SetGlobalLevel(document.LogLevelInfo)

	t.Run("网格列数在添加行后不变", func(t *testing.T) {
		doc := document.New()
		
		config := &document.TableConfig{
			Rows:  5,
			Cols:  4,
			Width: 8000,
		}

		table := doc.AddTable(config)
		originalGridColCount := len(table.Grid.Cols)

		// 添加10行
		for i := 0; i < 10; i++ {
			table.AppendRow([]string{"A", "B", "C", "D"})
		}

		currentGridColCount := len(table.Grid.Cols)
		if currentGridColCount != originalGridColCount {
			t.Errorf("添加行后网格列数改变: 原始%d，现在%d", 
				originalGridColCount, currentGridColCount)
		}

		// 验证Grid中的列数与实际表格列数一致
		if currentGridColCount != table.GetColumnCount() {
			t.Errorf("网格列数(%d)与表格列数(%d)不一致",
				currentGridColCount, table.GetColumnCount())
		}
	})

	t.Run("网格列数在合并单元格后不变", func(t *testing.T) {
		doc := document.New()
		
		config := &document.TableConfig{
			Rows:  5,
			Cols:  6,
			Width: 9000,
		}

		table := doc.AddTable(config)
		originalGridColCount := len(table.Grid.Cols)

		// 合并多个单元格
		table.MergeCellsHorizontal(1, 0, 2)
		table.MergeCellsHorizontal(2, 3, 5)
		table.MergeCellsVertical(0, 2, 1)

		currentGridColCount := len(table.Grid.Cols)
		if currentGridColCount != originalGridColCount {
			t.Errorf("合并单元格后网格列数改变: 原始%d，现在%d",
				originalGridColCount, currentGridColCount)
		}

		// Grid列数应该始终等于原始列数，不受合并影响
		if currentGridColCount != 6 {
			t.Errorf("网格列数应该保持为6，实际为%d", currentGridColCount)
		}
	})
}
