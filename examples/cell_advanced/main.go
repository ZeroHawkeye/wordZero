package main

import (
	"fmt"
	"log"

	"github.com/ZeroHawkeye/wordZero/pkg/document"
)

func main() {
	// 创建新文档
	doc := document.New()

	fmt.Println("=== 演示高级单元格操作功能 ===")

	// 1. 创建表格
	fmt.Println("\n1. 创建基础表格...")
	config := &document.TableConfig{
		Rows:  4,
		Cols:  4,
		Width: 8000,
		Data: [][]string{
			{"学号", "姓名", "语文", "数学"},
			{"001", "张三", "85", "92"},
			{"002", "李四", "78", "88"},
			{"003", "王五", "90", "85"},
		},
	}

	table, _ := doc.AddTable(config)
	if table == nil {
		log.Fatal("创建表格失败")
	}

	// 2. 设置表头格式
	fmt.Println("2. 设置表头格式...")
	headerFormat := &document.CellFormat{
		TextFormat: &document.TextFormat{
			Bold:       true,
			FontSize:   14,
			FontColor:  "FFFFFF", // 白色文字
			FontFamily: "微软雅黑",
		},
		HorizontalAlign: document.CellAlignCenter,
		VerticalAlign:   document.CellVAlignCenter,
	}

	// 为第一行的每个单元格设置表头格式
	for col := 0; col < 4; col++ {
		err := table.SetCellFormat(0, col, headerFormat)
		if err != nil {
			log.Printf("设置表头格式失败: %v", err)
		}
	}

	// 3. 设置数据行格式
	fmt.Println("3. 设置数据行格式...")
	dataFormat := &document.CellFormat{
		TextFormat: &document.TextFormat{
			FontSize:   12,
			FontFamily: "宋体",
		},
		HorizontalAlign: document.CellAlignCenter,
		VerticalAlign:   document.CellVAlignCenter,
	}

	// 为数据行设置格式
	for row := 1; row < 4; row++ {
		for col := 0; col < 4; col++ {
			err := table.SetCellFormat(row, col, dataFormat)
			if err != nil {
				log.Printf("设置数据格式失败: %v", err)
			}
		}
	}

	// 4. 演示富文本单元格
	fmt.Println("4. 添加富文本单元格...")

	// 在表格下方添加一个新表格用于演示富文本
	richTextConfig := &document.TableConfig{
		Rows:  2,
		Cols:  3,
		Width: 8000,
	}

	richTable, _ := doc.AddTable(richTextConfig)

	// 在同一个单元格中添加不同格式的文本
	err := richTable.SetCellFormattedText(0, 0, "标题：", &document.TextFormat{
		Bold:     true,
		FontSize: 14,
	})
	if err != nil {
		log.Printf("设置富文本失败: %v", err)
	}

	err = richTable.AddCellFormattedText(0, 0, "重要信息", &document.TextFormat{
		Bold:      true,
		FontColor: "FF0000", // 红色
		FontSize:  12,
	})
	if err != nil {
		log.Printf("添加富文本失败: %v", err)
	}

	err = richTable.AddCellFormattedText(0, 0, "（普通文本）", &document.TextFormat{
		FontSize: 10,
	})
	if err != nil {
		log.Printf("添加富文本失败: %v", err)
	}

	// 5. 演示单元格合并
	fmt.Println("5. 演示单元格合并...")

	// 创建合并演示表格
	mergeConfig := &document.TableConfig{
		Rows:  5,
		Cols:  5,
		Width: 8000,
	}

	mergeTable, _ := doc.AddTable(mergeConfig)

	// 设置初始数据
	for row := 0; row < 5; row++ {
		for col := 0; col < 5; col++ {
			text := fmt.Sprintf("R%dC%d", row+1, col+1)
			err := mergeTable.SetCellText(row, col, text)
			if err != nil {
				log.Printf("设置单元格文本失败: %v", err)
			}
		}
	}

	// 水平合并：合并第一行的第2-4列
	err = mergeTable.MergeCellsHorizontal(0, 1, 3)
	if err != nil {
		log.Printf("水平合并失败: %v", err)
	}

	// 设置合并单元格的内容
	err = mergeTable.SetCellFormattedText(0, 1, "水平合并单元格", &document.TextFormat{
		Bold:      true,
		FontSize:  14,
		FontColor: "0000FF",
	})
	if err != nil {
		log.Printf("设置合并单元格内容失败: %v", err)
	}

	// 垂直合并：合并第一列的第2-4行
	err = mergeTable.MergeCellsVertical(1, 3, 0)
	if err != nil {
		log.Printf("垂直合并失败: %v", err)
	}

	// 设置垂直合并单元格的内容
	err = mergeTable.SetCellFormattedText(1, 0, "垂直合并", &document.TextFormat{
		Bold:      true,
		FontSize:  14,
		FontColor: "00FF00",
	})
	if err != nil {
		log.Printf("设置垂直合并单元格内容失败: %v", err)
	}

	// 区域合并：合并右下角2x2区域
	err = mergeTable.MergeCellsRange(3, 4, 3, 4)
	if err != nil {
		log.Printf("区域合并失败: %v", err)
	}

	// 设置区域合并单元格的内容
	err = mergeTable.SetCellFormattedText(3, 3, "区域合并", &document.TextFormat{
		Bold:      true,
		FontSize:  14,
		FontColor: "FF00FF",
	})
	if err != nil {
		log.Printf("设置区域合并单元格内容失败: %v", err)
	}

	// 6. 检查合并状态
	fmt.Println("6. 检查合并状态...")

	mergeInfo, err := mergeTable.GetMergedCellInfo(0, 1)
	if err != nil {
		log.Printf("获取合并信息失败: %v", err)
	} else {
		fmt.Printf("单元格(0,1)合并信息: %+v\n", mergeInfo)
	}

	// 7. 演示单元格内容和格式操作
	fmt.Println("7. 演示内容和格式操作...")

	// 创建操作演示表格
	opConfig := &document.TableConfig{
		Rows:  3,
		Cols:  3,
		Width: 6000,
	}

	opTable, _ := doc.AddTable(opConfig)

	// 设置带格式的内容
	err = opTable.SetCellFormattedText(0, 0, "原始内容", &document.TextFormat{
		Bold:      true,
		FontSize:  14,
		FontColor: "FF0000",
	})
	if err != nil {
		log.Printf("设置格式化内容失败: %v", err)
	}

	// 清空内容但保留格式
	err = opTable.ClearCellContent(0, 1)
	if err != nil {
		log.Printf("清空内容失败: %v", err)
	}

	// 设置新内容
	err = opTable.SetCellText(0, 1, "新内容")
	if err != nil {
		log.Printf("设置新内容失败: %v", err)
	}

	// 清空格式但保留内容
	err = opTable.ClearCellFormat(0, 2)
	if err != nil {
		log.Printf("清空格式失败: %v", err)
	}

	// 8. 保存文档
	fmt.Println("8. 保存文档...")

	// 添加说明段落
	doc.AddParagraph("本文档演示了wordZero库的高级单元格操作功能：")
	doc.AddParagraph("1. 第一个表格展示了基础表格创建和格式设置")
	doc.AddParagraph("2. 第二个表格展示了富文本单元格功能")
	doc.AddParagraph("3. 第三个表格展示了单元格合并功能")
	doc.AddParagraph("4. 第四个表格展示了内容和格式操作功能")

	// 9. 演示文字方向功能
	fmt.Println("9. 演示文字方向功能...")

	// 添加文字方向演示说明
	doc.AddParagraph("")
	doc.AddParagraph("=== 单元格文字方向演示 ===")

	// 创建文字方向演示表格
	directionConfig := &document.TableConfig{
		Rows:  3,
		Cols:  3,
		Width: 8000,
	}

	directionTable, _ := doc.AddTable(directionConfig)

	// 演示不同的文字方向
	directions := []struct {
		name      string
		direction document.CellTextDirection
		row       int
		col       int
	}{
		{"水平从左到右", document.TextDirectionLR, 0, 0},
		{"垂直从上到下", document.TextDirectionTB, 0, 1},
		{"垂直从下到上", document.TextDirectionBT, 0, 2},
		{"水平从右到左", document.TextDirectionRL, 1, 0},
		{"垂直显示上下", document.TextDirectionTBV, 1, 1},
		{"垂直显示下上", document.TextDirectionBTV, 1, 2},
		{"混合格式文字", document.TextDirectionTB, 2, 0},
		{"居中垂直文字", document.TextDirectionTB, 2, 1},
		{"默认方向", document.TextDirectionLR, 2, 2},
	}

	for _, dir := range directions {
		// 设置单元格文字
		err := directionTable.SetCellText(dir.row, dir.col, dir.name)
		if err != nil {
			log.Printf("设置单元格文字失败: %v", err)
		}

		// 设置文字方向
		err = directionTable.SetCellTextDirection(dir.row, dir.col, dir.direction)
		if err != nil {
			log.Printf("设置文字方向失败: %v", err)
		}

		// 为某些特殊单元格添加格式
		if dir.row == 2 && dir.col == 0 {
			// 混合格式演示
			err = directionTable.SetCellFormattedText(dir.row, dir.col, "混合", &document.TextFormat{
				Bold:      true,
				FontColor: "FF0000",
				FontSize:  14,
			})
			if err != nil {
				log.Printf("设置富文本失败: %v", err)
			}

			err = directionTable.AddCellFormattedText(dir.row, dir.col, "格式", &document.TextFormat{
				Italic:    true,
				FontColor: "0000FF",
				FontSize:  12,
			})
			if err != nil {
				log.Printf("添加富文本失败: %v", err)
			}

			err = directionTable.AddCellFormattedText(dir.row, dir.col, "文字", &document.TextFormat{
				FontColor: "00FF00",
				FontSize:  10,
			})
			if err != nil {
				log.Printf("添加富文本失败: %v", err)
			}
		}

		if dir.row == 2 && dir.col == 1 {
			// 居中垂直文字
			format := &document.CellFormat{
				TextFormat: &document.TextFormat{
					Bold:     true,
					FontSize: 16,
				},
				HorizontalAlign: document.CellAlignCenter,
				VerticalAlign:   document.CellVAlignCenter,
				TextDirection:   document.TextDirectionTB,
			}

			err = directionTable.SetCellFormat(dir.row, dir.col, format)
			if err != nil {
				log.Printf("设置单元格格式失败: %v", err)
			}
		}
	}

	// 验证文字方向设置
	fmt.Println("10. 验证文字方向设置...")
	for _, dir := range directions {
		actualDirection, err := directionTable.GetCellTextDirection(dir.row, dir.col)
		if err != nil {
			log.Printf("获取文字方向失败: %v", err)
			continue
		}

		if actualDirection == dir.direction {
			fmt.Printf("✓ 单元格(%d,%d) 文字方向设置正确: %s\n", dir.row, dir.col, dir.direction)
		} else {
			fmt.Printf("✗ 单元格(%d,%d) 文字方向不匹配，期望: %s，实际: %s\n",
				dir.row, dir.col, dir.direction, actualDirection)
		}
	}

	// 11. 保存文档
	fmt.Println("11. 保存文档...")

	// 更新说明
	doc.AddParagraph("• 单元格文字方向设置（支持6种方向）")
	doc.AddParagraph("• 文字方向与其他格式的组合使用")

	filename := "examples/output/cell_advanced_demo.docx"
	err = doc.Save(filename)
	if err != nil {
		log.Fatalf("保存文档失败: %v", err)
	}

	fmt.Printf("文档已保存为: %s\n", filename)
	fmt.Println("=== 演示完成 ===")
}
