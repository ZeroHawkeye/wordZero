package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ZeroHawkeye/wordZero/pkg/document"
)

func main() {
	// 创建输出目录
	if err := os.MkdirAll("examples/output", 0755); err != nil {
		log.Fatalf("创建输出目录失败: %v", err)
	}

	// 创建文档
	doc := document.New()
	if doc == nil {
		log.Fatal("创建文档失败")
	}

	// 添加标题
	doc.AddParagraph("WordZero - 表格样式和外观功能演示")
	doc.AddParagraph("")

	// 演示1：基础表格边框
	demonstrateBorders(doc)

	// 演示2：单元格边框
	demonstrateCellBorders(doc)

	// 演示3：表格背景
	demonstrateTableShading(doc)

	// 演示4：单元格背景
	demonstrateCellShading(doc)

	// 演示5：奇偶行颜色交替
	demonstrateAlternatingRows(doc)

	// 演示6：预定义样式模板
	demonstrateStyleTemplates(doc)

	// 演示7：自定义表格样式
	demonstrateCustomStyle(doc)

	// 演示8：复杂样式组合
	demonstrateComplexStyle(doc)

	// 演示9：无边框表格
	demonstrateNoBorders(doc)

	// 演示10：各种边框样式
	demonstrateBorderStyles(doc)

	// 保存文档
	outputFile := "examples/output/table_style_demo.docx"
	if err := doc.Save(outputFile); err != nil {
		log.Fatalf("保存文档失败: %v", err)
	}

	fmt.Printf("表格样式演示文档已保存到: %s\n", outputFile)
}

// demonstrateBorders 演示表格边框设置
func demonstrateBorders(doc *document.Document) {
	doc.AddParagraph("1. 表格边框设置演示")

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

	table, _ := doc.AddTable(config)
	if table == nil {
		log.Fatal("创建表格失败")
	}

	// 设置不同的边框样式
	borderConfig := &document.TableBorderConfig{
		Top: &document.BorderConfig{
			Style: document.BorderStyleThick,
			Width: 12,
			Color: "FF0000", // 红色粗上边框
			Space: 0,
		},
		Left: &document.BorderConfig{
			Style: document.BorderStyleSingle,
			Width: 6,
			Color: "0000FF", // 蓝色左边框
			Space: 0,
		},
		Bottom: &document.BorderConfig{
			Style: document.BorderStyleDouble,
			Width: 6,
			Color: "00FF00", // 绿色双线下边框
			Space: 0,
		},
		Right: &document.BorderConfig{
			Style: document.BorderStyleDashed,
			Width: 6,
			Color: "FF00FF", // 紫色虚线右边框
			Space: 0,
		},
		InsideH: &document.BorderConfig{
			Style: document.BorderStyleDotted,
			Width: 4,
			Color: "808080", // 灰色点线内部水平边框
			Space: 0,
		},
		InsideV: &document.BorderConfig{
			Style: document.BorderStyleSingle,
			Width: 4,
			Color: "808080", // 灰色内部垂直边框
			Space: 0,
		},
	}

	err := table.SetTableBorders(borderConfig)
	if err != nil {
		log.Fatalf("设置表格边框失败: %v", err)
	}

	doc.AddParagraph("")
}

// demonstrateCellBorders 演示单元格边框设置
func demonstrateCellBorders(doc *document.Document) {
	doc.AddParagraph("2. 单元格边框设置演示")

	config := &document.TableConfig{
		Rows:  2,
		Cols:  3,
		Width: 4500,
		Data: [][]string{
			{"单元格A", "单元格B", "单元格C"},
			{"数据1", "数据2", "数据3"},
		},
	}

	table, _ := doc.AddTable(config)
	if table == nil {
		log.Fatal("创建表格失败")
	}

	// 为第一个单元格设置特殊边框
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
			Color: "FFFF00", // 黄色对角线
			Space: 0,
		},
	}

	err := table.SetCellBorders(0, 0, cellBorderConfig)
	if err != nil {
		log.Fatalf("设置单元格边框失败: %v", err)
	}

	doc.AddParagraph("")
}

// demonstrateTableShading 演示表格背景设置
func demonstrateTableShading(doc *document.Document) {
	doc.AddParagraph("3. 表格背景设置演示")

	config := &document.TableConfig{
		Rows:  3,
		Cols:  2,
		Width: 4000,
		Data: [][]string{
			{"产品", "价格"},
			{"产品A", "￥100"},
			{"产品B", "￥200"},
		},
	}

	table, _ := doc.AddTable(config)
	if table == nil {
		log.Fatal("创建表格失败")
	}

	// 设置表格整体背景
	shadingConfig := &document.ShadingConfig{
		Pattern:         document.ShadingPatternPct25,
		ForegroundColor: "000000",
		BackgroundColor: "E0E0E0",
	}

	err := table.SetTableShading(shadingConfig)
	if err != nil {
		log.Fatalf("设置表格背景失败: %v", err)
	}

	doc.AddParagraph("")
}

// demonstrateCellShading 演示单元格背景设置
func demonstrateCellShading(doc *document.Document) {
	doc.AddParagraph("4. 单元格背景设置演示")

	config := &document.TableConfig{
		Rows:  3,
		Cols:  3,
		Width: 4500,
		Data: [][]string{
			{"红色", "绿色", "蓝色"},
			{"黄色", "紫色", "青色"},
			{"橙色", "粉色", "灰色"},
		},
	}

	table, _ := doc.AddTable(config)
	if table == nil {
		log.Fatal("创建表格失败")
	}

	// 为不同单元格设置不同颜色
	colors := [][]string{
		{"FF0000", "00FF00", "0000FF"}, // 红绿蓝
		{"FFFF00", "FF00FF", "00FFFF"}, // 黄紫青
		{"FFA500", "FFC0CB", "808080"}, // 橙粉灰
	}

	patterns := [][]document.ShadingPattern{
		{document.ShadingPatternSolid, document.ShadingPatternPct50, document.ShadingPatternPct25},
		{document.ShadingPatternSolid, document.ShadingPatternPct75, document.ShadingPatternPct10},
		{document.ShadingPatternPct60, document.ShadingPatternPct40, document.ShadingPatternSolid},
	}

	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			shadingConfig := &document.ShadingConfig{
				Pattern:         patterns[i][j],
				BackgroundColor: colors[i][j],
			}

			err := table.SetCellShading(i, j, shadingConfig)
			if err != nil {
				log.Fatalf("设置单元格(%d,%d)背景失败: %v", i, j, err)
			}
		}
	}

	doc.AddParagraph("")
}

// demonstrateAlternatingRows 演示奇偶行颜色交替
func demonstrateAlternatingRows(doc *document.Document) {
	doc.AddParagraph("5. 奇偶行颜色交替演示")

	config := &document.TableConfig{
		Rows:  6,
		Cols:  3,
		Width: 4500,
		Data: [][]string{
			{"序号", "姓名", "分数"},
			{"1", "张三", "85"},
			{"2", "李四", "92"},
			{"3", "王五", "78"},
			{"4", "赵六", "95"},
			{"5", "钱七", "88"},
		},
	}

	table, _ := doc.AddTable(config)
	if table == nil {
		log.Fatal("创建表格失败")
	}

	// 设置奇偶行颜色交替
	err := table.SetAlternatingRowColors("F0F0F0", "FFFFFF")
	if err != nil {
		log.Fatalf("设置奇偶行颜色交替失败: %v", err)
	}

	doc.AddParagraph("")
}

// demonstrateStyleTemplates 演示预定义样式模板
func demonstrateStyleTemplates(doc *document.Document) {
	doc.AddParagraph("6. 预定义样式模板演示")

	// 演示Grid样式
	config1 := &document.TableConfig{
		Rows:  4,
		Cols:  3,
		Width: 4500,
		Data: [][]string{
			{"项目", "预算", "实际"},
			{"开发", "10000", "9500"},
			{"测试", "5000", "5200"},
			{"总计", "15000", "14700"},
		},
	}

	table1, _ := doc.AddTable(config1)
	if table1 == nil {
		log.Fatal("创建表格失败")
	}

	styleConfig1 := &document.TableStyleConfig{
		Template:       document.TableStyleTemplateGrid,
		FirstRowHeader: true,
		LastRowTotal:   true,
		BandedRows:     true,
		BandedColumns:  false,
	}

	err := table1.ApplyTableStyle(styleConfig1)
	if err != nil {
		log.Fatalf("应用Grid样式失败: %v", err)
	}

	doc.AddParagraph("")

	// 演示Colorful样式
	config2 := &document.TableConfig{
		Rows:  4,
		Cols:  3,
		Width: 4500,
		Data: [][]string{
			{"部门", "人数", "预算"},
			{"技术部", "20", "500万"},
			{"销售部", "15", "300万"},
			{"市场部", "10", "200万"},
		},
	}

	table2, _ := doc.AddTable(config2)
	if table2 == nil {
		log.Fatal("创建表格失败")
	}

	styleConfig2 := &document.TableStyleConfig{
		Template:          document.TableStyleTemplateColorful1,
		FirstRowHeader:    true,
		FirstColumnHeader: true,
		BandedRows:        true,
	}

	err = table2.ApplyTableStyle(styleConfig2)
	if err != nil {
		log.Fatalf("应用Colorful样式失败: %v", err)
	}

	doc.AddParagraph("")
}

// demonstrateCustomStyle 演示自定义表格样式
func demonstrateCustomStyle(doc *document.Document) {
	doc.AddParagraph("7. 自定义表格样式演示")

	config := &document.TableConfig{
		Rows:  4,
		Cols:  3,
		Width: 4500,
		Data: [][]string{
			{"功能", "状态", "备注"},
			{"登录", "完成", "已测试"},
			{"注册", "开发中", "50%"},
			{"支付", "计划中", "下个版本"},
		},
	}

	table, _ := doc.AddTable(config)
	if table == nil {
		log.Fatal("创建表格失败")
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

	// 应用自定义样式
	err := table.CreateCustomTableStyle("CustomBlue", "蓝色主题", borderConfig, shadingConfig, true)
	if err != nil {
		log.Fatalf("创建自定义表格样式失败: %v", err)
	}

	doc.AddParagraph("")
}

// demonstrateComplexStyle 演示复杂样式组合
func demonstrateComplexStyle(doc *document.Document) {
	doc.AddParagraph("8. 复杂样式组合演示")

	config := &document.TableConfig{
		Rows:  5,
		Cols:  4,
		Width: 6000,
		Data: [][]string{
			{"部门", "Q1", "Q2", "Q3"},
			{"销售部", "120万", "135万", "150万"},
			{"技术部", "80万", "90万", "95万"},
			{"市场部", "60万", "70万", "85万"},
			{"总计", "260万", "295万", "330万"},
		},
	}

	table, _ := doc.AddTable(config)
	if table == nil {
		log.Fatal("创建表格失败")
	}

	// 应用基础样式
	styleConfig := &document.TableStyleConfig{
		Template:          document.TableStyleTemplateColorful2,
		FirstRowHeader:    true,
		LastRowTotal:      true,
		FirstColumnHeader: true,
		BandedRows:        false,
	}

	err := table.ApplyTableStyle(styleConfig)
	if err != nil {
		log.Fatalf("应用基础样式失败: %v", err)
	}

	// 设置特殊边框
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
		log.Fatalf("设置边框失败: %v", err)
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
			log.Fatalf("设置标题行背景失败: %v", err)
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
			log.Fatalf("设置总计行背景失败: %v", err)
		}
	}

	doc.AddParagraph("")
}

// demonstrateNoBorders 演示无边框表格
func demonstrateNoBorders(doc *document.Document) {
	doc.AddParagraph("9. 无边框表格演示")

	config := &document.TableConfig{
		Rows:  3,
		Cols:  3,
		Width: 4500,
		Data: [][]string{
			{"项目A", "项目B", "项目C"},
			{"说明1", "说明2", "说明3"},
			{"结果1", "结果2", "结果3"},
		},
	}

	table, _ := doc.AddTable(config)
	if table == nil {
		log.Fatal("创建表格失败")
	}

	// 移除所有边框
	err := table.RemoveTableBorders()
	if err != nil {
		log.Fatalf("移除表格边框失败: %v", err)
	}

	// 设置轻微的背景色以区分单元格
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			var bgColor string
			if (i+j)%2 == 0 {
				bgColor = "F8F8F8"
			} else {
				bgColor = "FFFFFF"
			}

			shadingConfig := &document.ShadingConfig{
				Pattern:         document.ShadingPatternSolid,
				BackgroundColor: bgColor,
			}

			err = table.SetCellShading(i, j, shadingConfig)
			if err != nil {
				log.Fatalf("设置单元格背景失败: %v", err)
			}
		}
	}

	doc.AddParagraph("")
}

// demonstrateBorderStyles 演示各种边框样式
func demonstrateBorderStyles(doc *document.Document) {
	doc.AddParagraph("10. 各种边框样式演示")

	borderStyles := []struct {
		style document.BorderStyle
		name  string
	}{
		{document.BorderStyleSingle, "单线"},
		{document.BorderStyleThick, "粗线"},
		{document.BorderStyleDouble, "双线"},
		{document.BorderStyleDotted, "点线"},
		{document.BorderStyleDashed, "虚线"},
		{document.BorderStyleDotDash, "点划线"},
		{document.BorderStyleWave, "波浪线"},
	}

	for i, styleInfo := range borderStyles {
		config := &document.TableConfig{
			Rows:  2,
			Cols:  2,
			Width: 3000,
			Data: [][]string{
				{fmt.Sprintf("样式%d", i+1), styleInfo.name},
				{"演示", "数据"},
			},
		}

		table, _ := doc.AddTable(config)
		if table == nil {
			log.Fatalf("创建表格%d失败", i+1)
		}

		borderConfig := &document.TableBorderConfig{
			Top: &document.BorderConfig{
				Style: styleInfo.style,
				Width: 6,
				Color: "000000",
				Space: 0,
			},
			Left: &document.BorderConfig{
				Style: styleInfo.style,
				Width: 6,
				Color: "000000",
				Space: 0,
			},
			Bottom: &document.BorderConfig{
				Style: styleInfo.style,
				Width: 6,
				Color: "000000",
				Space: 0,
			},
			Right: &document.BorderConfig{
				Style: styleInfo.style,
				Width: 6,
				Color: "000000",
				Space: 0,
			},
		}

		err := table.SetTableBorders(borderConfig)
		if err != nil {
			log.Fatalf("设置表格%d边框失败: %v", i+1, err)
		}

		if i < len(borderStyles)-1 {
			doc.AddParagraph("")
		}
	}
}
