package main

import (
	"fmt"
	"log"

	"github.com/ZeroHawkeye/wordZero/pkg/document"
)

func main() {
	// 创建文档
	doc := document.New()
	if doc == nil {
		log.Fatal("创建文档失败")
	}

	doc.AddParagraph("Word文档表格默认样式演示")
	doc.AddParagraph("")

	// 1. 展示新的默认表格样式
	doc.AddParagraph("1. 新的默认表格样式（参考tmp_test的单线边框）")

	config1 := &document.TableConfig{
		Rows:  3,
		Cols:  3,
		Width: 5000,
		Data: [][]string{
			{"姓名", "年龄", "职业"},
			{"张三", "25", "工程师"},
			{"李四", "30", "设计师"},
		},
	}

	table1, _ := doc.AddTable(config1)
	if table1 == nil {
		log.Fatal("创建表格1失败")
	}

	// 2. 对比原来无边框的效果
	doc.AddParagraph("2. 移除边框的表格样式（原来的默认效果）")

	config2 := &document.TableConfig{
		Rows:  3,
		Cols:  3,
		Width: 5000,
		Data: [][]string{
			{"产品", "价格", "库存"},
			{"笔记本", "5000元", "50台"},
			{"台式机", "3000元", "30台"},
		},
	}

	table2, _ := doc.AddTable(config2)
	if table2 == nil {
		log.Fatal("创建表格2失败")
	}

	// 手动移除边框来展示原来的效果
	err := table2.RemoveTableBorders()
	if err != nil {
		log.Fatalf("移除表格边框失败: %v", err)
	}

	doc.AddParagraph("")
	doc.AddParagraph("☝️ 该表格手动移除了边框，展示了原来的默认效果")
	doc.AddParagraph("")

	// 3. 展示可以覆盖默认样式
	doc.AddParagraph("3. 自定义样式覆盖默认样式")

	config3 := &document.TableConfig{
		Rows:  4,
		Cols:  3,
		Width: 6000,
		Data: [][]string{
			{"部门", "预算", "实际支出"},
			{"技术部", "100万", "95万"},
			{"销售部", "80万", "85万"},
			{"总计", "180万", "180万"},
		},
	}

	table3, _ := doc.AddTable(config3)
	if table3 == nil {
		log.Fatal("创建表格3失败")
	}

	// 应用自定义边框样式
	borderConfig := &document.TableBorderConfig{
		Top: &document.BorderConfig{
			Style: document.BorderStyleDouble,
			Width: 8,
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
			Style: document.BorderStyleDouble,
			Width: 8,
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

	err = table3.SetTableBorders(borderConfig)
	if err != nil {
		log.Fatalf("设置自定义边框失败: %v", err)
	}

	// 为标题行设置特殊背景
	for j := 0; j < 3; j++ {
		shadingConfig := &document.ShadingConfig{
			Pattern:         document.ShadingPatternSolid,
			BackgroundColor: "2E75B6",
			ForegroundColor: "FFFFFF",
		}

		err = table3.SetCellShading(0, j, shadingConfig)
		if err != nil {
			log.Fatalf("设置标题行背景失败: %v", err)
		}
	}

	// 为总计行设置特殊背景
	for j := 0; j < 3; j++ {
		shadingConfig := &document.ShadingConfig{
			Pattern:         document.ShadingPatternSolid,
			BackgroundColor: "FFFF99",
		}

		err = table3.SetCellShading(3, j, shadingConfig)
		if err != nil {
			log.Fatalf("设置总计行背景失败: %v", err)
		}
	}

	doc.AddParagraph("")
	doc.AddParagraph("☝️ 该表格使用了自定义样式，覆盖了默认的单线边框")
	doc.AddParagraph("")

	// 4. 展示与tmp_test参考表格相同的配置
	doc.AddParagraph("4. 与tmp_test参考表格相同的配置")

	config4 := &document.TableConfig{
		Rows:  3,
		Cols:  3,
		Width: 8522, // 与tmp_test中的总宽度匹配
		Data: [][]string{
			{"Cell A1", "Cell B1", "Cell C1"},
			{"Cell A2", "Cell B2", "Cell C2"},
			{"Cell A3", "Cell B3", "Cell C3"},
		},
	}

	table4, _ := doc.AddTable(config4)
	if table4 == nil {
		log.Fatal("创建表格4失败")
	}

	doc.AddParagraph("")
	doc.AddParagraph("☝️ 该表格完全匹配tmp_test参考表格的样式和尺寸")
	doc.AddParagraph("")

	// 保存文档
	outputPath := "examples/output/table_default_style_demo.docx"
	err = doc.Save(outputPath)
	if err != nil {
		log.Fatalf("保存文档失败: %v", err)
	}

	fmt.Printf("表格默认样式演示文档已保存到: %s\n", outputPath)
	fmt.Println("\n✅ 新的默认表格样式功能已成功实现：")
	fmt.Println("   - 默认使用单线边框（参考tmp_test）")
	fmt.Println("   - 边框粗细为4（1/8磅）")
	fmt.Println("   - 自动调整布局（autofit）")
	fmt.Println("   - 标准单元格边距（108 dxa）")
	fmt.Println("   - 仍然支持自定义样式覆盖")
}
