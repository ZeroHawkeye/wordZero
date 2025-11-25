package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/ZeroHawkeye/wordZero/pkg/document"
)

func main() {
	fmt.Println("WordZero 表格布局和尺寸功能演示")
	fmt.Println("==============================")

	// 创建新文档
	doc := document.New()

	// 添加文档标题
	title := doc.AddParagraph("表格布局和尺寸功能演示")
	title.SetAlignment(document.AlignCenter)
	titleFormat := &document.TextFormat{
		Bold:       true,
		FontSize:   18,
		FontColor:  "2F5496",
		FontFamily: "微软雅黑",
	}
	title.AddFormattedText("", titleFormat)

	// 1. 行高设置演示
	fmt.Println("1. 创建行高设置演示表格...")
	doc.AddParagraph("1. 行高设置演示").SetStyle("Heading1")

	heightTable, _ := doc.AddTable(&document.TableConfig{
		Rows:  4,
		Cols:  3,
		Width: 8000,
		Data: [][]string{
			{"行类型", "高度设置", "说明"},
			{"自动行高", "默认", "根据内容自动调整高度"},
			{"固定行高", "40磅", "精确的固定高度"},
			{"最小行高", "30磅", "至少30磅，内容多时可以更高"},
		},
	})

	// 设置表头格式
	headerFormat := &document.CellFormat{
		TextFormat: &document.TextFormat{
			Bold:       true,
			FontSize:   12,
			FontColor:  "FFFFFF",
			FontFamily: "微软雅黑",
		},
		HorizontalAlign: document.CellAlignCenter,
		VerticalAlign:   document.CellVAlignCenter,
	}

	for col := 0; col < 3; col++ {
		heightTable.SetCellFormat(0, col, headerFormat)
	}

	// 应用不同的行高设置
	// 第2行：固定高度40磅
	heightTable.SetRowHeight(2, &document.RowHeightConfig{
		Height: 40,
		Rule:   document.RowHeightExact,
	})

	// 第3行：最小高度30磅
	heightTable.SetRowHeight(3, &document.RowHeightConfig{
		Height: 30,
		Rule:   document.RowHeightMinimum,
	})

	fmt.Println("   - 设置了不同的行高规则")

	// 2. 表格对齐演示
	fmt.Println("2. 创建表格对齐演示...")
	doc.AddParagraph("2. 表格对齐演示").SetStyle("Heading1")

	// 左对齐表格
	doc.AddParagraph("2.1 左对齐表格").SetStyle("Heading2")
	leftTable, _ := doc.AddTable(&document.TableConfig{
		Rows:  2,
		Cols:  2,
		Width: 4000,
		Data: [][]string{
			{"左对齐", "表格"},
			{"Left", "Aligned"},
		},
	})
	leftTable.SetTableAlignment(document.TableAlignLeft)

	// 居中对齐表格
	doc.AddParagraph("2.2 居中对齐表格").SetStyle("Heading2")
	centerTable, _ := doc.AddTable(&document.TableConfig{
		Rows:  2,
		Cols:  2,
		Width: 4000,
		Data: [][]string{
			{"居中对齐", "表格"},
			{"Center", "Aligned"},
		},
	})
	centerTable.SetTableAlignment(document.TableAlignCenter)

	// 右对齐表格
	doc.AddParagraph("2.3 右对齐表格").SetStyle("Heading2")
	rightTable, _ := doc.AddTable(&document.TableConfig{
		Rows:  2,
		Cols:  2,
		Width: 4000,
		Data: [][]string{
			{"右对齐", "表格"},
			{"Right", "Aligned"},
		},
	})
	rightTable.SetTableAlignment(document.TableAlignRight)

	fmt.Println("   - 创建了左对齐、居中、右对齐三种表格")

	// 3. 分页控制演示
	fmt.Println("3. 创建分页控制演示表格...")
	doc.AddParagraph("3. 分页控制演示").SetStyle("Heading1")

	pageBreakTable, _ := doc.AddTable(&document.TableConfig{
		Rows:  6,
		Cols:  4,
		Width: 9000,
		Data: [][]string{
			{"序号", "姓名", "部门", "职位"},
			{"001", "张三", "技术部", "工程师"},
			{"002", "李四", "产品部", "产品经理"},
			{"003", "王五", "设计部", "UI设计师"},
			{"004", "赵六", "市场部", "市场专员"},
			{"005", "钱七", "人事部", "HR专员"},
		},
	})

	// 设置表头格式
	for col := 0; col < 4; col++ {
		pageBreakTable.SetCellFormat(0, col, &document.CellFormat{
			TextFormat: &document.TextFormat{
				Bold:      true,
				FontSize:  12,
				FontColor: "FFFFFF",
			},
			HorizontalAlign: document.CellAlignCenter,
			VerticalAlign:   document.CellVAlignCenter,
		})
	}

	// 设置第一行为重复的标题行
	pageBreakTable.SetRowAsHeader(0, true)
	fmt.Println("   - 设置第一行为重复标题行")

	// 设置某些行禁止跨页分割
	pageBreakTable.SetRowKeepTogether(1, true)
	pageBreakTable.SetRowKeepTogether(2, true)
	fmt.Println("   - 设置前两个数据行禁止跨页分割")

	// 设置表格分页配置
	pageBreakTable.SetTablePageBreak(&document.TablePageBreakConfig{
		KeepWithNext:    false,
		KeepLines:       true,
		PageBreakBefore: false,
		WidowControl:    true,
	})

	// 4. 复杂布局演示
	fmt.Println("4. 创建复杂布局演示表格...")
	doc.AddParagraph("4. 复杂布局演示").SetStyle("Heading1")

	complexTable, _ := doc.AddTable(&document.TableConfig{
		Rows:      5,
		Cols:      4,
		Width:     9000,
		ColWidths: []int{1500, 3000, 2000, 2500},
		Data: [][]string{
			{"项目", "描述", "状态", "负责人"},
			{"WordZero核心", "Word文档操作库核心功能", "进行中", "开发团队"},
			{"表格功能", "完整的表格操作和格式化", "已完成", "张工程师"},
			{"样式系统", "18种预定义样式和自定义样式", "已完成", "李设计师"},
			{"测试套件", "完整的单元测试和集成测试", "进行中", "QA团队"},
		},
	})

	// 设置表头为重复标题行
	complexTable.SetHeaderRows(0, 0)

	// 设置不同的行高
	complexTable.SetRowHeight(0, &document.RowHeightConfig{
		Height: 35,
		Rule:   document.RowHeightExact,
	}) // 表头固定高度

	// 批量设置数据行为最小高度
	complexTable.SetRowHeightRange(1, 4, &document.RowHeightConfig{
		Height: 25,
		Rule:   document.RowHeightMinimum,
	})

	// 设置表头格式
	for col := 0; col < 4; col++ {
		complexTable.SetCellFormat(0, col, &document.CellFormat{
			TextFormat: &document.TextFormat{
				Bold:       true,
				FontSize:   14,
				FontColor:  "FFFFFF",
				FontFamily: "微软雅黑",
			},
			HorizontalAlign: document.CellAlignCenter,
			VerticalAlign:   document.CellVAlignCenter,
		})
	}

	// 设置状态列的特殊格式
	for row := 1; row <= 4; row++ {
		status, _ := complexTable.GetCellText(row, 2)
		var color string
		switch status {
		case "已完成":
			color = "00AA00" // 绿色
		case "进行中":
			color = "FF8800" // 橙色
		default:
			color = "666666" // 灰色
		}

		complexTable.SetCellFormat(row, 2, &document.CellFormat{
			TextFormat: &document.TextFormat{
				Bold:      true,
				FontColor: color,
				FontSize:  11,
			},
			HorizontalAlign: document.CellAlignCenter,
			VerticalAlign:   document.CellVAlignCenter,
		})
	}

	// 设置表格居中对齐
	complexTable.SetTableAlignment(document.TableAlignCenter)

	fmt.Println("   - 设置了自定义列宽")
	fmt.Println("   - 应用了不同的行高规则")
	fmt.Println("   - 添加了状态标识颜色")

	// 5. 输出统计信息
	fmt.Println("5. 生成统计信息...")
	doc.AddParagraph("5. 表格统计信息").SetStyle("Heading1")

	// 获取分页信息
	breakInfo := pageBreakTable.GetTableBreakInfo()
	infoText := fmt.Sprintf("分页控制表格统计：总行数 %d，标题行数 %d，禁止分割行数 %d",
		breakInfo["total_rows"], breakInfo["header_rows"], breakInfo["keep_together_rows"])
	doc.AddParagraph(infoText)

	// 获取表格布局信息
	layout := complexTable.GetTableLayout()
	layoutText := fmt.Sprintf("复杂表格布局：对齐方式 %s，环绕类型 %s，定位类型 %s",
		layout.Alignment, layout.TextWrap, layout.Position)
	doc.AddParagraph(layoutText)

	// 验证行高设置
	heightConfig, _ := complexTable.GetRowHeight(0)
	heightText := fmt.Sprintf("表头行高设置：%d磅，规则 %s", heightConfig.Height, heightConfig.Rule)
	doc.AddParagraph(heightText)

	// 6. 保存文档
	outputPath := filepath.Join("examples", "output", "table_layout_demo.docx")
	err := doc.Save(outputPath)
	if err != nil {
		log.Fatalf("保存文档失败: %v", err)
	}

	fmt.Printf("6. 文档已保存到: %s\n", outputPath)
	fmt.Println("\n演示完成！")
	fmt.Println("\n新实现的功能包括：")
	fmt.Println("✅ 行高设置（固定高度、最小高度、自动调整）")
	fmt.Println("✅ 表格对齐方式（左对齐、居中、右对齐）")
	fmt.Println("✅ 分页控制（标题行重复、禁止跨页分割）")
	fmt.Println("✅ 表格布局配置和查询")
	fmt.Println("✅ 批量行高设置和统计信息获取")
}
