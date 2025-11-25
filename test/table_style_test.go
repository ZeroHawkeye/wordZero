package test

import (
	"fmt"
	"testing"

	"github.com/ZeroHawkeye/wordZero/pkg/document"
)

// TestTableBorders 测试表格边框功能
func TestTableBorders(t *testing.T) {
	// 创建文档
	doc := document.New()
	if doc == nil {
		t.Fatal("创建文档失败")
	}

	// 创建表格
	config := &document.TableConfig{
		Rows:  3,
		Cols:  3,
		Width: 5000,
		Data: [][]string{
			{"姓名", "年龄", "职业"},
			{"张三", "25", "工程师"},
			{"李四", "30", "设计师"},
		},
	}

	table, err := doc.AddTable(config)
	if err != nil {
		t.Fatalf("创建表格失败: %v", err)
	}
	if table == nil {
		t.Fatal("创建表格失败")
	}

	// 测试设置表格边框
	borderConfig := &document.TableBorderConfig{
		Top: &document.BorderConfig{
			Style: document.BorderStyleThick,
			Width: 12,
			Color: "FF0000",
			Space: 0,
		},
		Left: &document.BorderConfig{
			Style: document.BorderStyleSingle,
			Width: 6,
			Color: "0000FF",
			Space: 0,
		},
		Bottom: &document.BorderConfig{
			Style: document.BorderStyleDouble,
			Width: 6,
			Color: "00FF00",
			Space: 0,
		},
		Right: &document.BorderConfig{
			Style: document.BorderStyleDashed,
			Width: 6,
			Color: "FF00FF",
			Space: 0,
		},
		InsideH: &document.BorderConfig{
			Style: document.BorderStyleDotted,
			Width: 4,
			Color: "808080",
			Space: 0,
		},
		InsideV: &document.BorderConfig{
			Style: document.BorderStyleSingle,
			Width: 4,
			Color: "808080",
			Space: 0,
		},
	}

	err = table.SetTableBorders(borderConfig)
	if err != nil {
		t.Fatalf("设置表格边框失败: %v", err)
	}

	t.Log("表格边框设置成功")
}

// TestCellBorders 测试单元格边框功能
func TestCellBorders(t *testing.T) {
	// 创建文档
	doc := document.New()
	if doc == nil {
		t.Fatal("创建文档失败")
	}

	// 创建表格
	config := &document.TableConfig{
		Rows:  2,
		Cols:  2,
		Width: 3000,
		Data: [][]string{
			{"A1", "B1"},
			{"A2", "B2"},
		},
	}

	table, err := doc.AddTable(config)
	if err != nil {
		t.Fatalf("创建表格失败: %v", err)
	}
	if table == nil {
		t.Fatal("创建表格失败")
	}

	// 测试设置单元格边框
	cellBorderConfig := &document.CellBorderConfig{
		Top: &document.BorderConfig{
			Style: document.BorderStyleThick,
			Width: 8,
			Color: "FF0000",
			Space: 0,
		},
		Left: &document.BorderConfig{
			Style: document.BorderStyleSingle,
			Width: 4,
			Color: "0000FF",
			Space: 0,
		},
		Bottom: &document.BorderConfig{
			Style: document.BorderStyleDouble,
			Width: 6,
			Color: "00FF00",
			Space: 0,
		},
		Right: &document.BorderConfig{
			Style: document.BorderStyleDashed,
			Width: 4,
			Color: "FF00FF",
			Space: 0,
		},
		DiagDown: &document.BorderConfig{
			Style: document.BorderStyleSingle,
			Width: 2,
			Color: "FFFF00",
			Space: 0,
		},
	}

	err = table.SetCellBorders(0, 0, cellBorderConfig)
	if err != nil {
		t.Fatalf("设置单元格边框失败: %v", err)
	}

	t.Log("单元格边框设置成功")
}

// TestTableShading 测试表格背景功能
func TestTableShading(t *testing.T) {
	// 创建文档
	doc := document.New()
	if doc == nil {
		t.Fatal("创建文档失败")
	}

	// 创建表格
	config := &document.TableConfig{
		Rows:  3,
		Cols:  2,
		Width: 4000,
		Data: [][]string{
			{"产品", "价格"},
			{"产品A", "100"},
			{"产品B", "200"},
		},
	}

	table, err := doc.AddTable(config)
	if err != nil {
		t.Fatalf("创建表格失败: %v", err)
	}
	if table == nil {
		t.Fatal("创建表格失败")
	}

	// 测试设置表格背景
	shadingConfig := &document.ShadingConfig{
		Pattern:         document.ShadingPatternPct25,
		ForegroundColor: "000000",
		BackgroundColor: "E0E0E0",
	}

	err = table.SetTableShading(shadingConfig)
	if err != nil {
		t.Fatalf("设置表格背景失败: %v", err)
	}

	t.Log("表格背景设置成功")
}

// TestCellShading 测试单元格背景功能
func TestCellShading(t *testing.T) {
	// 创建文档
	doc := document.New()
	if doc == nil {
		t.Fatal("创建文档失败")
	}

	// 创建表格
	config := &document.TableConfig{
		Rows:  3,
		Cols:  3,
		Width: 4500,
		Data: [][]string{
			{"A", "B", "C"},
			{"1", "2", "3"},
			{"X", "Y", "Z"},
		},
	}

	table, err := doc.AddTable(config)
	if err != nil {
		t.Fatalf("创建表格失败: %v", err)
	}
	if table == nil {
		t.Fatal("创建表格失败")
	}

	// 测试设置单元格背景
	testCases := []struct {
		row             int
		col             int
		backgroundColor string
		pattern         document.ShadingPattern
	}{
		{0, 0, "FF0000", document.ShadingPatternSolid}, // 红色实色
		{0, 1, "00FF00", document.ShadingPatternPct50}, // 绿色50%
		{0, 2, "0000FF", document.ShadingPatternPct25}, // 蓝色25%
		{1, 0, "FFFF00", document.ShadingPatternSolid}, // 黄色实色
		{1, 1, "FF00FF", document.ShadingPatternPct75}, // 紫色75%
		{1, 2, "00FFFF", document.ShadingPatternPct10}, // 青色10%
	}

	for _, tc := range testCases {
		shadingConfig := &document.ShadingConfig{
			Pattern:         tc.pattern,
			BackgroundColor: tc.backgroundColor,
		}

		err = table.SetCellShading(tc.row, tc.col, shadingConfig)
		if err != nil {
			t.Fatalf("设置单元格(%d,%d)背景失败: %v", tc.row, tc.col, err)
		}
	}

	t.Log("单元格背景设置成功")
}

// TestAlternatingRowColors 测试奇偶行颜色交替
func TestAlternatingRowColors(t *testing.T) {
	// 创建文档
	doc := document.New()
	if doc == nil {
		t.Fatal("创建文档失败")
	}

	// 创建表格
	config := &document.TableConfig{
		Rows:  5,
		Cols:  3,
		Width: 4500,
		Data: [][]string{
			{"序号", "姓名", "分数"},
			{"1", "张三", "85"},
			{"2", "李四", "92"},
			{"3", "王五", "78"},
			{"4", "赵六", "95"},
		},
	}

	table, err := doc.AddTable(config)
	if err != nil {
		t.Fatalf("创建表格失败: %v", err)
	}
	if table == nil {
		t.Fatal("创建表格失败")
	}

	// 测试设置奇偶行颜色交替
	err = table.SetAlternatingRowColors("F0F0F0", "FFFFFF")
	if err != nil {
		t.Fatalf("设置奇偶行颜色交替失败: %v", err)
	}

	t.Log("奇偶行颜色交替设置成功")
}

// TestTableStyleTemplates 测试表格样式模板
func TestTableStyleTemplates(t *testing.T) {
	// 创建文档
	doc := document.New()
	if doc == nil {
		t.Fatal("创建文档失败")
	}

	// 创建表格
	config := &document.TableConfig{
		Rows:  4,
		Cols:  3,
		Width: 5000,
		Data: [][]string{
			{"项目", "预算", "实际"},
			{"开发", "10000", "9500"},
			{"测试", "5000", "5200"},
			{"总计", "15000", "14700"},
		},
	}

	table, err := doc.AddTable(config)
	if err != nil {
		t.Fatalf("创建表格失败: %v", err)
	}
	if table == nil {
		t.Fatal("创建表格失败")
	}

	// 测试应用表格样式模板
	styleConfig := &document.TableStyleConfig{
		Template:       document.TableStyleTemplateGrid,
		FirstRowHeader: true,
		LastRowTotal:   true,
		BandedRows:     true,
		BandedColumns:  false,
	}

	err = table.ApplyTableStyle(styleConfig)
	if err != nil {
		t.Fatalf("应用表格样式模板失败: %v", err)
	}

	t.Log("表格样式模板应用成功")
}

// TestCustomTableStyle 测试自定义表格样式
func TestCustomTableStyle(t *testing.T) {
	// 创建文档
	doc := document.New()
	if doc == nil {
		t.Fatal("创建文档失败")
	}

	// 创建表格
	config := &document.TableConfig{
		Rows:  3,
		Cols:  3,
		Width: 4500,
		Data: [][]string{
			{"功能", "状态", "备注"},
			{"登录", "完成", "已测试"},
			{"注册", "开发中", "50%"},
		},
	}

	table, err := doc.AddTable(config)
	if err != nil {
		t.Fatalf("创建表格失败: %v", err)
	}
	if table == nil {
		t.Fatal("创建表格失败")
	}

	// 创建自定义边框配置
	borderConfig := &document.TableBorderConfig{
		Top: &document.BorderConfig{
			Style: document.BorderStyleThick,
			Width: 12,
			Color: "2E75B6",
			Space: 0,
		},
		Left: &document.BorderConfig{
			Style: document.BorderStyleSingle,
			Width: 6,
			Color: "2E75B6",
			Space: 0,
		},
		Bottom: &document.BorderConfig{
			Style: document.BorderStyleThick,
			Width: 12,
			Color: "2E75B6",
			Space: 0,
		},
		Right: &document.BorderConfig{
			Style: document.BorderStyleSingle,
			Width: 6,
			Color: "2E75B6",
			Space: 0,
		},
		InsideH: &document.BorderConfig{
			Style: document.BorderStyleSingle,
			Width: 4,
			Color: "D0D0D0",
			Space: 0,
		},
		InsideV: &document.BorderConfig{
			Style: document.BorderStyleSingle,
			Width: 4,
			Color: "D0D0D0",
			Space: 0,
		},
	}

	// 创建自定义背景配置
	shadingConfig := &document.ShadingConfig{
		Pattern:         document.ShadingPatternPct10,
		BackgroundColor: "E7F3FF",
	}

	// 测试创建自定义表格样式
	err = table.CreateCustomTableStyle("CustomStyle1", "自定义样式1", borderConfig, shadingConfig, true)
	if err != nil {
		t.Fatalf("创建自定义表格样式失败: %v", err)
	}

	t.Log("自定义表格样式创建成功")
}

// TestRemoveBorders 测试移除边框功能
func TestRemoveBorders(t *testing.T) {
	// 创建文档
	doc := document.New()
	if doc == nil {
		t.Fatal("创建文档失败")
	}

	// 创建表格
	config := &document.TableConfig{
		Rows:  2,
		Cols:  2,
		Width: 3000,
		Data: [][]string{
			{"A", "B"},
			{"C", "D"},
		},
	}

	table, err := doc.AddTable(config)
	if err != nil {
		t.Fatalf("创建表格失败: %v", err)
	}
	if table == nil {
		t.Fatal("创建表格失败")
	}

	// 先设置边框
	borderConfig := &document.TableBorderConfig{
		Top: &document.BorderConfig{
			Style: document.BorderStyleSingle,
			Width: 6,
			Color: "000000",
			Space: 0,
		},
		Left: &document.BorderConfig{
			Style: document.BorderStyleSingle,
			Width: 6,
			Color: "000000",
			Space: 0,
		},
		Bottom: &document.BorderConfig{
			Style: document.BorderStyleSingle,
			Width: 6,
			Color: "000000",
			Space: 0,
		},
		Right: &document.BorderConfig{
			Style: document.BorderStyleSingle,
			Width: 6,
			Color: "000000",
			Space: 0,
		},
	}

	err = table.SetTableBorders(borderConfig)
	if err != nil {
		t.Fatalf("设置表格边框失败: %v", err)
	}

	// 测试移除表格边框
	err = table.RemoveTableBorders()
	if err != nil {
		t.Fatalf("移除表格边框失败: %v", err)
	}

	// 测试移除单元格边框
	err = table.RemoveCellBorders(0, 0)
	if err != nil {
		t.Fatalf("移除单元格边框失败: %v", err)
	}

	t.Log("移除边框功能测试成功")
}

// TestComplexTableStyle 测试复杂表格样式组合
func TestComplexTableStyle(t *testing.T) {
	// 创建文档
	doc := document.New()
	if doc == nil {
		t.Fatal("创建文档失败")
	}

	// 创建表格
	config := &document.TableConfig{
		Rows:  5,
		Cols:  4,
		Width: 6000,
		Data: [][]string{
			{"部门", "Q1", "Q2", "Q3"},
			{"销售部", "120", "135", "150"},
			{"技术部", "80", "90", "95"},
			{"市场部", "60", "70", "85"},
			{"总计", "260", "295", "330"},
		},
	}

	table, err := doc.AddTable(config)
	if err != nil {
		t.Fatalf("创建表格失败: %v", err)
	}
	if table == nil {
		t.Fatal("创建表格失败")
	}

	// 应用样式模板
	styleConfig := &document.TableStyleConfig{
		Template:          document.TableStyleTemplateColorful1,
		FirstRowHeader:    true,
		LastRowTotal:      true,
		FirstColumnHeader: true,
		BandedRows:        true,
	}

	err = table.ApplyTableStyle(styleConfig)
	if err != nil {
		t.Fatalf("应用表格样式失败: %v", err)
	}

	// 设置自定义边框
	borderConfig := &document.TableBorderConfig{
		Top: &document.BorderConfig{
			Style: document.BorderStyleDouble,
			Width: 8,
			Color: "2E75B6",
			Space: 0,
		},
		Bottom: &document.BorderConfig{
			Style: document.BorderStyleDouble,
			Width: 8,
			Color: "2E75B6",
			Space: 0,
		},
		InsideH: &document.BorderConfig{
			Style: document.BorderStyleSingle,
			Width: 4,
			Color: "B0B0B0",
			Space: 0,
		},
		InsideV: &document.BorderConfig{
			Style: document.BorderStyleSingle,
			Width: 4,
			Color: "B0B0B0",
			Space: 0,
		},
	}

	err = table.SetTableBorders(borderConfig)
	if err != nil {
		t.Fatalf("设置表格边框失败: %v", err)
	}

	// 为标题行设置特殊背景
	for j := 0; j < 4; j++ {
		shadingConfig := &document.ShadingConfig{
			Pattern:         document.ShadingPatternSolid,
			BackgroundColor: "2E75B6",
			ForegroundColor: "FFFFFF",
		}

		err = table.SetCellShading(0, j, shadingConfig)
		if err != nil {
			t.Fatalf("设置标题行背景失败: %v", err)
		}
	}

	// 为总计行设置特殊背景
	for j := 0; j < 4; j++ {
		shadingConfig := &document.ShadingConfig{
			Pattern:         document.ShadingPatternSolid,
			BackgroundColor: "FFFF99",
		}

		err = table.SetCellShading(4, j, shadingConfig)
		if err != nil {
			t.Fatalf("设置总计行背景失败: %v", err)
		}
	}

	t.Log("复杂表格样式组合测试成功")
}

// BenchmarkTableStyleOperations 表格样式操作性能测试
func BenchmarkTableStyleOperations(b *testing.B) {
	// 创建文档
	doc := document.New()
	if doc == nil {
		b.Fatal("创建文档失败")
	}

	// 创建表格
	config := &document.TableConfig{
		Rows:  10,
		Cols:  5,
		Width: 7500,
	}

	table, err := doc.AddTable(config)
	if err != nil {
		b.Fatalf("创建表格失败: %v", err)
	}

	borderConfig := &document.TableBorderConfig{
		Top: &document.BorderConfig{
			Style: document.BorderStyleSingle,
			Width: 6,
			Color: "000000",
			Space: 0,
		},
		Left: &document.BorderConfig{
			Style: document.BorderStyleSingle,
			Width: 6,
			Color: "000000",
			Space: 0,
		},
		Bottom: &document.BorderConfig{
			Style: document.BorderStyleSingle,
			Width: 6,
			Color: "000000",
			Space: 0,
		},
		Right: &document.BorderConfig{
			Style: document.BorderStyleSingle,
			Width: 6,
			Color: "000000",
			Space: 0,
		},
		InsideH: &document.BorderConfig{
			Style: document.BorderStyleSingle,
			Width: 4,
			Color: "C0C0C0",
			Space: 0,
		},
		InsideV: &document.BorderConfig{
			Style: document.BorderStyleSingle,
			Width: 4,
			Color: "C0C0C0",
			Space: 0,
		},
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err = table.SetTableBorders(borderConfig)
		if err != nil {
			b.Fatalf("设置表格边框失败: %v", err)
		}
	}
}

// TestBorderStyles 测试各种边框样式
func TestBorderStyles(t *testing.T) {
	// 创建文档
	doc := document.New()
	if doc == nil {
		t.Fatal("创建文档失败")
	}

	// 测试边框样式
	borderStyles := []document.BorderStyle{
		document.BorderStyleNone,
		document.BorderStyleSingle,
		document.BorderStyleThick,
		document.BorderStyleDouble,
		document.BorderStyleDotted,
		document.BorderStyleDashed,
		document.BorderStyleDotDash,
		document.BorderStyleWave,
	}

	for i, style := range borderStyles {
		// 为每种样式创建一个表格
		config := &document.TableConfig{
			Rows:  2,
			Cols:  2,
			Width: 2000,
			Data: [][]string{
				{fmt.Sprintf("样式%d", i+1), string(style)},
				{"测试", "数据"},
			},
		}

		table, err := doc.AddTable(config)
		if err != nil {
			t.Fatalf("创建表格失败: %v", err)
		}

		borderConfig := &document.TableBorderConfig{
			Top: &document.BorderConfig{
				Style: style,
				Width: 6,
				Color: "000000",
				Space: 0,
			},
			Left: &document.BorderConfig{
				Style: style,
				Width: 6,
				Color: "000000",
				Space: 0,
			},
			Bottom: &document.BorderConfig{
				Style: style,
				Width: 6,
				Color: "000000",
				Space: 0,
			},
			Right: &document.BorderConfig{
				Style: style,
				Width: 6,
				Color: "000000",
				Space: 0,
			},
		}

		err = table.SetTableBorders(borderConfig)
		if err != nil {
			t.Fatalf("设置表格%d边框失败: %v", i+1, err)
		}
	}

	t.Log("边框样式测试完成")
}

// TestShadingPatterns 测试各种底纹图案
func TestShadingPatterns(t *testing.T) {
	// 创建文档
	doc := document.New()
	if doc == nil {
		t.Fatal("创建文档失败")
	}

	// 测试底纹图案
	patterns := []document.ShadingPattern{
		document.ShadingPatternClear,
		document.ShadingPatternSolid,
		document.ShadingPatternPct25,
		document.ShadingPatternPct50,
		document.ShadingPatternPct75,
		document.ShadingPatternHorzStripe,
		document.ShadingPatternVertStripe,
		document.ShadingPatternDiagStripe,
	}

	// 创建表格
	config := &document.TableConfig{
		Rows:  len(patterns),
		Cols:  2,
		Width: 3000,
	}

	table, err := doc.AddTable(config)
	if err != nil {
		t.Fatalf("创建表格失败: %v", err)
	}
	if table == nil {
		t.Fatal("创建表格失败")
	}

	for i, pattern := range patterns {
		// 设置单元格内容
		err = table.SetCellText(i, 0, fmt.Sprintf("图案%d", i+1))
		if err != nil {
			t.Fatalf("设置单元格文本失败: %v", err)
		}

		err = table.SetCellText(i, 1, string(pattern))
		if err != nil {
			t.Fatalf("设置单元格文本失败: %v", err)
		}

		// 设置单元格背景
		shadingConfig := &document.ShadingConfig{
			Pattern:         pattern,
			BackgroundColor: "C0C0C0",
			ForegroundColor: "000000",
		}

		err = table.SetCellShading(i, 1, shadingConfig)
		if err != nil {
			t.Fatalf("设置单元格背景失败: %v", err)
		}
	}

	t.Log("底纹图案测试完成")
}
