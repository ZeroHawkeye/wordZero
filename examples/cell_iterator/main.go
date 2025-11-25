package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ZeroHawkeye/wordZero/pkg/document"
)

func main() {
	fmt.Println("=== 单元格迭代器功能演示 ===")

	// 创建新文档
	doc := document.New()

	// 创建一个3x4的测试表格
	config := &document.TableConfig{
		Rows:  3,
		Cols:  4,
		Width: 8000,
		Data: [][]string{
			{"产品", "价格", "数量", "总计"},
			{"苹果", "5.00", "10", "50.00"},
			{"橙子", "3.50", "15", "52.50"},
		},
	}

	table, _ := doc.AddTable(config)
	if table == nil {
		log.Fatal("创建表格失败")
	}

	fmt.Printf("创建了 %dx%d 的表格\n", table.GetRowCount(), table.GetColumnCount())

	// 演示1: 基本迭代器使用
	fmt.Println("\n--- 1. 基本迭代器遍历 ---")
	iterator := table.NewCellIterator()
	fmt.Printf("表格总共有 %d 个单元格\n", iterator.Total())

	cellCount := 0
	for iterator.HasNext() {
		cellInfo, err := iterator.Next()
		if err != nil {
			log.Printf("迭代器错误: %v", err)
			break
		}

		cellCount++
		fmt.Printf("单元格[%d,%d]: '%s'", cellInfo.Row, cellInfo.Col, cellInfo.Text)

		if cellInfo.IsLast {
			fmt.Print(" (最后一个)")
		}

		fmt.Printf(" - 进度: %.1f%%\n", iterator.Progress()*100)
	}

	// 演示2: 重置迭代器
	fmt.Println("\n--- 2. 重置迭代器演示 ---")
	iterator.Reset()
	row, col := iterator.Current()
	fmt.Printf("重置后，当前位置: (%d, %d)\n", row, col)

	// 只遍历前3个单元格
	for i := 0; i < 3 && iterator.HasNext(); i++ {
		cellInfo, _ := iterator.Next()
		fmt.Printf("单元格[%d,%d]: '%s'\n", cellInfo.Row, cellInfo.Col, cellInfo.Text)
	}

	// 演示3: ForEach方法
	fmt.Println("\n--- 3. ForEach 批量处理演示 ---")
	err := table.ForEach(func(row, col int, cell *document.TableCell, text string) error {
		if text != "" {
			fmt.Printf("处理单元格[%d,%d]: '%s' (长度: %d)\n", row, col, text, len(text))
		}
		return nil
	})

	if err != nil {
		log.Printf("ForEach执行失败: %v", err)
	}

	// 演示4: 按行遍历
	fmt.Println("\n--- 4. 按行遍历演示 ---")
	for row := 0; row < table.GetRowCount(); row++ {
		fmt.Printf("第%d行: ", row+1)
		err := table.ForEachInRow(row, func(col int, cell *document.TableCell, text string) error {
			fmt.Printf("'%s' ", text)
			return nil
		})
		if err != nil {
			log.Printf("按行遍历失败: %v", err)
		}
		fmt.Println()
	}

	// 演示5: 按列遍历
	fmt.Println("\n--- 5. 按列遍历演示 ---")
	for col := 0; col < table.GetColumnCount(); col++ {
		fmt.Printf("第%d列: ", col+1)
		err := table.ForEachInColumn(col, func(row int, cell *document.TableCell, text string) error {
			fmt.Printf("'%s' ", text)
			return nil
		})
		if err != nil {
			log.Printf("按列遍历失败: %v", err)
		}
		fmt.Println()
	}

	// 演示6: 获取范围单元格
	fmt.Println("\n--- 6. 获取单元格范围演示 ---")
	cells, err := table.GetCellRange(1, 1, 2, 3) // 获取价格、数量、总计的数据部分
	if err != nil {
		log.Printf("获取范围失败: %v", err)
	} else {
		fmt.Printf("范围 (1,1) 到 (2,3) 的单元格:\n")
		for _, cellInfo := range cells {
			fmt.Printf("  [%d,%d]: '%s'\n", cellInfo.Row, cellInfo.Col, cellInfo.Text)
		}
	}

	// 演示7: 查找单元格
	fmt.Println("\n--- 7. 查找单元格演示 ---")

	// 查找包含数字的单元格
	numberCells, err := table.FindCells(func(row, col int, cell *document.TableCell, text string) bool {
		// 简单检查是否包含数字字符
		for _, char := range text {
			if char >= '0' && char <= '9' {
				return true
			}
		}
		return false
	})

	if err != nil {
		log.Printf("查找失败: %v", err)
	} else {
		fmt.Printf("找到 %d 个包含数字的单元格:\n", len(numberCells))
		for _, cellInfo := range numberCells {
			fmt.Printf("  [%d,%d]: '%s'\n", cellInfo.Row, cellInfo.Col, cellInfo.Text)
		}
	}

	// 演示8: 按文本查找
	fmt.Println("\n--- 8. 按文本查找演示 ---")

	// 精确查找
	exactCells, err := table.FindCellsByText("苹果", true)
	if err != nil {
		log.Printf("精确查找失败: %v", err)
	} else {
		fmt.Printf("精确匹配 '苹果' 的单元格: %d 个\n", len(exactCells))
		for _, cellInfo := range exactCells {
			fmt.Printf("  [%d,%d]: '%s'\n", cellInfo.Row, cellInfo.Col, cellInfo.Text)
		}
	}

	// 模糊查找
	fuzzyCells, err := table.FindCellsByText("5", false)
	if err != nil {
		log.Printf("模糊查找失败: %v", err)
	} else {
		fmt.Printf("包含 '5' 的单元格: %d 个\n", len(fuzzyCells))
		for _, cellInfo := range fuzzyCells {
			fmt.Printf("  [%d,%d]: '%s'\n", cellInfo.Row, cellInfo.Col, cellInfo.Text)
		}
	}

	// 保存文档
	outputDir := "examples/output"
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		log.Printf("创建输出目录失败: %v", err)
	}

	filename := filepath.Join(outputDir, "cell_iterator_demo.docx")
	if err := doc.Save(filename); err != nil {
		log.Printf("保存文档失败: %v", err)
	} else {
		fmt.Printf("\n文档已保存到: %s\n", filename)
	}

	fmt.Println("\n=== 单元格迭代器演示完成 ===")
}
