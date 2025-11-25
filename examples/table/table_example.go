// Package main 演示WordZero表格功能
package main

import (
	"fmt"
	"log"

	"github.com/ZeroHawkeye/wordZero/pkg/document"
)

func main() {
	fmt.Println("=== WordZero 表格功能演示 ===")

	// 创建新文档
	doc := document.New()

	// 添加文档标题
	title := doc.AddParagraph("WordZero 表格功能演示")
	title.SetAlignment(document.AlignCenter)

	// 演示1：创建基础表格
	fmt.Println("1. 创建基础表格...")
	demonstrateBasicTable(doc)

	// 演示2：表格数据操作
	fmt.Println("2. 演示表格数据操作...")
	demonstrateTableDataOperations(doc)

	// 演示3：表格结构操作
	fmt.Println("3. 演示表格结构操作...")
	demonstrateTableStructureOperations(doc)

	// 演示4：表格复制和清空
	fmt.Println("4. 演示表格复制和清空...")
	demonstrateTableCopyAndClear(doc)

	// 演示5：表格删除操作
	fmt.Println("5. 演示表格删除操作...")
	demonstrateTableDeletion(doc)

	// 保存文档
	outputFile := "examples/output/table_demo.docx"
	err := doc.Save(outputFile)
	if err != nil {
		log.Fatalf("保存文档失败: %v", err)
	}

	fmt.Printf("表格演示文档已保存到: %s\n", outputFile)
	fmt.Println("=== 演示完成 ===")
}

// demonstrateBasicTable 演示基础表格创建
func demonstrateBasicTable(doc *document.Document) {
	doc.AddParagraph("1. 基础表格创建")

	// 创建一个3x4的表格，包含初始数据
	tableData := [][]string{
		{"姓名", "年龄", "职位", "部门"},
		{"张三", "28", "工程师", "技术部"},
		{"李四", "32", "经理", "销售部"},
	}

	config := &document.TableConfig{
		Rows:  3,
		Cols:  4,
		Width: 8000, // 8000磅宽度
		Data:  tableData,
	}

	table, _ := doc.AddTable(config)
	if table != nil {
		fmt.Printf("   创建表格成功：%dx%d\n", table.GetRowCount(), table.GetColumnCount())
	}

	doc.AddParagraph("") // 空行分隔
}

// demonstrateTableDataOperations 演示表格数据操作
func demonstrateTableDataOperations(doc *document.Document) {
	doc.AddParagraph("2. 表格数据操作")

	// 创建一个简单的2x3表格
	config := &document.TableConfig{
		Rows:  2,
		Cols:  3,
		Width: 6000,
		Data: [][]string{
			{"产品", "价格", "库存"},
			{"笔记本", "5000", "50"},
		},
	}

	table, _ := doc.AddTable(config)
	if table == nil {
		fmt.Println("   创建表格失败")
		return
	}

	// 设置单元格内容
	err := table.SetCellText(0, 0, "商品名称")
	if err != nil {
		fmt.Printf("   设置单元格失败: %v\n", err)
	} else {
		fmt.Println("   设置单元格内容成功")
	}

	// 获取单元格内容
	cellText, err := table.GetCellText(1, 1)
	if err != nil {
		fmt.Printf("   获取单元格失败: %v\n", err)
	} else {
		fmt.Printf("   单元格(1,1)内容: %s\n", cellText)
	}

	doc.AddParagraph("") // 空行分隔
}

// demonstrateTableStructureOperations 演示表格结构操作
func demonstrateTableStructureOperations(doc *document.Document) {
	doc.AddParagraph("3. 表格结构操作")

	// 创建基础表格
	config := &document.TableConfig{
		Rows:  2,
		Cols:  2,
		Width: 4000,
		Data: [][]string{
			{"A1", "B1"},
			{"A2", "B2"},
		},
	}

	table, _ := doc.AddTable(config)
	if table == nil {
		fmt.Println("   创建表格失败")
		return
	}

	fmt.Printf("   初始表格大小: %dx%d\n", table.GetRowCount(), table.GetColumnCount())

	// 插入行
	err := table.InsertRow(1, []string{"A1.5", "B1.5"})
	if err != nil {
		fmt.Printf("   插入行失败: %v\n", err)
	} else {
		fmt.Printf("   插入行后表格大小: %dx%d\n", table.GetRowCount(), table.GetColumnCount())
	}

	// 添加行到末尾
	err = table.AppendRow([]string{"A末", "B末"})
	if err != nil {
		fmt.Printf("   添加行失败: %v\n", err)
	} else {
		fmt.Printf("   添加行后表格大小: %dx%d\n", table.GetRowCount(), table.GetColumnCount())
	}

	// 插入列
	err = table.InsertColumn(1, []string{"C1", "C1.5", "C2", "C末"}, 1000)
	if err != nil {
		fmt.Printf("   插入列失败: %v\n", err)
	} else {
		fmt.Printf("   插入列后表格大小: %dx%d\n", table.GetRowCount(), table.GetColumnCount())
	}

	// 添加列到末尾
	err = table.AppendColumn([]string{"D1", "D1.5", "D2", "D末"}, 1000)
	if err != nil {
		fmt.Printf("   添加列失败: %v\n", err)
	} else {
		fmt.Printf("   添加列后表格大小: %dx%d\n", table.GetRowCount(), table.GetColumnCount())
	}

	doc.AddParagraph("") // 空行分隔
}

// demonstrateTableCopyAndClear 演示表格复制和清空
func demonstrateTableCopyAndClear(doc *document.Document) {
	doc.AddParagraph("4. 表格复制和清空")

	// 创建源表格
	config := &document.TableConfig{
		Rows:  2,
		Cols:  2,
		Width: 4000,
		Data: [][]string{
			{"原始1", "原始2"},
			{"原始3", "原始4"},
		},
	}

	sourceTable, _ := doc.AddTable(config)
	if sourceTable == nil {
		fmt.Println("   创建源表格失败")
		return
	}

	// 复制表格
	copiedTable := sourceTable.CopyTable()
	if copiedTable != nil {
		fmt.Println("   表格复制成功")

		// 修改复制的表格内容以区分
		copiedTable.SetCellText(0, 0, "复制1")
		copiedTable.SetCellText(0, 1, "复制2")
		copiedTable.SetCellText(1, 0, "复制3")
		copiedTable.SetCellText(1, 1, "复制4")

		// 将复制的表格添加到文档
		doc.Body.AddElement(copiedTable)
		fmt.Println("   复制的表格已添加到文档")
	}

	// 清空原表格内容
	sourceTable.ClearTable()
	fmt.Println("   原表格内容已清空")

	doc.AddParagraph("") // 空行分隔
}

// demonstrateTableDeletion 演示表格删除操作
func demonstrateTableDeletion(doc *document.Document) {
	doc.AddParagraph("5. 表格删除操作")

	// 创建测试表格
	config := &document.TableConfig{
		Rows:  4,
		Cols:  4,
		Width: 6000,
		Data: [][]string{
			{"1", "2", "3", "4"},
			{"5", "6", "7", "8"},
			{"9", "10", "11", "12"},
			{"13", "14", "15", "16"},
		},
	}

	table, _ := doc.AddTable(config)
	if table == nil {
		fmt.Println("   创建测试表格失败")
		return
	}

	fmt.Printf("   初始表格大小: %dx%d\n", table.GetRowCount(), table.GetColumnCount())

	// 删除第2行（索引1）
	err := table.DeleteRow(1)
	if err != nil {
		fmt.Printf("   删除行失败: %v\n", err)
	} else {
		fmt.Printf("   删除行后表格大小: %dx%d\n", table.GetRowCount(), table.GetColumnCount())
	}

	// 删除第2列（索引1）
	err = table.DeleteColumn(1)
	if err != nil {
		fmt.Printf("   删除列失败: %v\n", err)
	} else {
		fmt.Printf("   删除列后表格大小: %dx%d\n", table.GetRowCount(), table.GetColumnCount())
	}

	// 删除多行（索引1到2）
	err = table.DeleteRows(1, 2)
	if err != nil {
		fmt.Printf("   删除多行失败: %v\n", err)
	} else {
		fmt.Printf("   删除多行后表格大小: %dx%d\n", table.GetRowCount(), table.GetColumnCount())
	}

	doc.AddParagraph("") // 空行分隔
}
