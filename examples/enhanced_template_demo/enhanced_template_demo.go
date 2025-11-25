// Package main 演示增强的模板功能
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ZeroHawkeye/wordZero/pkg/document"
)

func main() {
	fmt.Println("=== WordZero 增强模板功能演示 ===")

	// 确保输出目录存在
	os.MkdirAll("examples/output", 0755)

	// 演示1：保持样式的变量替换
	demonstrateStyledVariableTemplate()

	// 演示2：表格模板功能
	demonstrateTableTemplate()

	// 演示3：复杂文档模板
	demonstrateComplexDocumentTemplate()

	fmt.Println("\n✅ 增强模板功能演示完成！")
}

// demonstrateStyledVariableTemplate 演示保持样式的变量替换
func demonstrateStyledVariableTemplate() {
	fmt.Println("\n📝 演示1：保持样式的变量替换")

	// 创建一个包含格式化内容的模板文档
	templateDoc := document.New()

	// 添加标题
	titlePara := templateDoc.AddParagraph("")
	titleRun := &document.Run{
		Properties: &document.RunProperties{
			Bold:     &document.Bold{},
			FontSize: &document.FontSize{Val: "32"}, // 16磅
			Color:    &document.Color{Val: "0066CC"},
		},
		Text: document.Text{Content: "{{title}}"},
	}
	titlePara.Runs = []document.Run{*titleRun}
	titlePara.SetAlignment(document.AlignCenter)

	// 添加副标题
	subtitlePara := templateDoc.AddParagraph("")
	subtitleRun := &document.Run{
		Properties: &document.RunProperties{
			Italic:   &document.Italic{},
			FontSize: &document.FontSize{Val: "24"}, // 12磅
			Color:    &document.Color{Val: "666666"},
		},
		Text: document.Text{Content: "作者：{{author}} | 日期：{{date}}"},
	}
	subtitlePara.Runs = []document.Run{*subtitleRun}
	subtitlePara.SetAlignment(document.AlignCenter)

	// 添加正文段落
	bodyPara := templateDoc.AddParagraph("")

	// 混合格式的正文
	normalRun := &document.Run{
		Text: document.Text{Content: "这是一个"},
	}
	boldRun := &document.Run{
		Properties: &document.RunProperties{
			Bold: &document.Bold{},
		},
		Text: document.Text{Content: "{{status}}"},
	}
	endRun := &document.Run{
		Text: document.Text{Content: "的项目，当前进度为"},
	}
	progressRun := &document.Run{
		Properties: &document.RunProperties{
			Bold:  &document.Bold{},
			Color: &document.Color{Val: "FF0000"},
		},
		Text: document.Text{Content: "{{progress}}%"},
	}
	finalRun := &document.Run{
		Text: document.Text{Content: "。"},
	}

	bodyPara.Runs = []document.Run{*normalRun, *boldRun, *endRun, *progressRun, *finalRun}

	// 保存模板文档
	templateFile := "examples/output/styled_template.docx"
	err := templateDoc.Save(templateFile)
	if err != nil {
		log.Fatalf("保存模板文档失败: %v", err)
	}
	fmt.Printf("✓ 创建样式模板文档: %s\n", templateFile)

	// 创建模板引擎并加载模板
	engine := document.NewTemplateEngine()
	_, err = engine.LoadTemplateFromDocument("styled_template", templateDoc)
	if err != nil {
		log.Fatalf("加载模板失败: %v", err)
	}

	// 准备数据
	data := document.NewTemplateData()
	data.SetVariable("title", "WordZero 项目报告")
	data.SetVariable("author", "张开发")
	data.SetVariable("date", time.Now().Format("2006年01月02日"))
	data.SetVariable("status", "进行中")
	data.SetVariable("progress", "85")

	// 使用新的渲染方法
	resultDoc, err := engine.RenderTemplateToDocument("styled_template", data)
	if err != nil {
		log.Fatalf("渲染模板失败: %v", err)
	}

	// 保存结果文档
	outputFile := "examples/output/styled_result_" + time.Now().Format("20060102_150405") + ".docx"
	err = resultDoc.Save(outputFile)
	if err != nil {
		log.Fatalf("保存结果文档失败: %v", err)
	}

	fmt.Printf("✓ 生成保持样式的文档: %s\n", outputFile)
}

// demonstrateTableTemplate 演示表格模板功能
func demonstrateTableTemplate() {
	fmt.Println("\n📊 演示2：表格模板功能")

	// 创建包含表格模板的文档
	templateDoc := document.New()

	// 添加标题
	templateDoc.AddHeadingParagraph("销售报表", 1)

	// 创建表格模板
	tableConfig := &document.TableConfig{
		Rows:  2, // 表头 + 模板行
		Cols:  4,
		Width: 9000, // 15cm
	}
	table, _ := templateDoc.CreateTable(tableConfig)

	// 设置表头
	table.SetCellText(0, 0, "产品名称")
	table.SetCellText(0, 1, "销售数量")
	table.SetCellText(0, 2, "单价")
	table.SetCellText(0, 3, "总金额")

	// 设置表头样式
	headerFormat := &document.TextFormat{
		Bold:      true,
		FontSize:  12,
		FontColor: "FFFFFF",
	}

	headerTexts := []string{"产品名称", "销售数量", "单价", "总金额"}
	for i := 0; i < 4; i++ {
		table.SetCellFormattedText(0, i, headerTexts[i], headerFormat)

		// 设置表头背景色
		table.SetCellShading(0, i, &document.ShadingConfig{
			Pattern:         document.ShadingPatternClear,
			BackgroundColor: "366092",
		})
	}

	// 设置模板行（包含循环语法）
	table.SetCellText(1, 0, "{{#each items}}{{name}}")
	table.SetCellText(1, 1, "{{quantity}}")
	table.SetCellText(1, 2, "{{price}}")
	table.SetCellText(1, 3, "{{total}}{{/each}}")

	// 添加表格到文档
	templateDoc.Body.AddElement(table)

	// 保存模板文档
	templateFile := "examples/output/table_template.docx"
	err := templateDoc.Save(templateFile)
	if err != nil {
		log.Fatalf("保存表格模板失败: %v", err)
	}
	fmt.Printf("✓ 创建表格模板文档: %s\n", templateFile)

	// 创建模板引擎并加载模板
	engine := document.NewTemplateEngine()
	_, err = engine.LoadTemplateFromDocument("table_template", templateDoc)
	if err != nil {
		log.Fatalf("加载表格模板失败: %v", err)
	}

	// 准备销售数据
	data := document.NewTemplateData()

	salesItems := []interface{}{
		map[string]interface{}{
			"name":     "WordZero专业版",
			"quantity": "10",
			"price":    "¥999.00",
			"total":    "¥9,990.00",
		},
		map[string]interface{}{
			"name":     "技术支持服务",
			"quantity": "12",
			"price":    "¥500.00",
			"total":    "¥6,000.00",
		},
		map[string]interface{}{
			"name":     "培训课程",
			"quantity": "5",
			"price":    "¥800.00",
			"total":    "¥4,000.00",
		},
	}

	data.SetList("items", salesItems)

	// 渲染表格模板
	resultDoc, err := engine.RenderTemplateToDocument("table_template", data)
	if err != nil {
		log.Fatalf("渲染表格模板失败: %v", err)
	}

	// 保存结果文档
	outputFile := "examples/output/table_result_" + time.Now().Format("20060102_150405") + ".docx"
	err = resultDoc.Save(outputFile)
	if err != nil {
		log.Fatalf("保存表格结果失败: %v", err)
	}

	fmt.Printf("✓ 生成表格报表文档: %s\n", outputFile)
}

// demonstrateComplexDocumentTemplate 演示复杂文档模板
func demonstrateComplexDocumentTemplate() {
	fmt.Println("\n📋 演示3：复杂文档模板")

	// 创建复杂的文档模板
	templateDoc := document.New()

	// 注释：文档属性设置功能需要实现
	// templateDoc.SetProperty("title", "{{projectName}} - 项目报告")
	// templateDoc.SetProperty("author", "{{manager}}")
	// templateDoc.SetProperty("subject", "项目进度报告")

	// 添加封面
	titlePara := templateDoc.AddParagraph("")
	titleRun := &document.Run{
		Properties: &document.RunProperties{
			Bold:     &document.Bold{},
			FontSize: &document.FontSize{Val: "48"}, // 24磅
			Color:    &document.Color{Val: "2F5496"},
		},
		Text: document.Text{Content: "{{projectName}}"},
	}
	titlePara.Runs = []document.Run{*titleRun}
	titlePara.SetAlignment(document.AlignCenter)
	titlePara.SetSpacing(&document.SpacingConfig{
		BeforePara: 72, // 1英寸
		AfterPara:  36, // 0.5英寸
	})

	// 添加项目基本信息
	templateDoc.AddHeadingParagraph("项目基本信息", 2)

	templateDoc.AddParagraph("项目经理：{{manager}}\n报告日期：{{reportDate}}\n项目状态：{{#if isActive}}进行中{{/if}}{{#if isComplete}}已完成{{/if}}\n完成进度：{{progress}}%")

	// 添加团队成员表格
	templateDoc.AddHeadingParagraph("团队成员", 2)

	teamTableConfig := &document.TableConfig{
		Rows:  2, // 表头 + 模板行
		Cols:  3,
		Width: 8000,
	}
	teamTable, _ := templateDoc.CreateTable(teamTableConfig)

	// 设置团队表格表头
	teamTable.SetCellText(0, 0, "姓名")
	teamTable.SetCellText(0, 1, "角色")
	teamTable.SetCellText(0, 2, "工作内容")

	// 设置团队表格模板行
	teamTable.SetCellText(1, 0, "{{#each team}}{{name}}")
	teamTable.SetCellText(1, 1, "{{role}}")
	teamTable.SetCellText(1, 2, "{{work}}{{/each}}")

	templateDoc.Body.AddElement(teamTable)

	// 保存复杂模板
	templateFile := "examples/output/complex_template.docx"
	err := templateDoc.Save(templateFile)
	if err != nil {
		log.Fatalf("保存复杂模板失败: %v", err)
	}
	fmt.Printf("✓ 创建复杂文档模板: %s\n", templateFile)

	// 创建模板引擎并渲染
	engine := document.NewTemplateEngine()
	_, err = engine.LoadTemplateFromDocument("complex_template", templateDoc)
	if err != nil {
		log.Fatalf("加载复杂模板失败: %v", err)
	}

	// 准备复杂数据
	data := document.NewTemplateData()
	data.SetVariable("projectName", "WordZero 企业文档管理系统")
	data.SetVariable("manager", "李项目经理")
	data.SetVariable("reportDate", time.Now().Format("2006年01月02日"))
	data.SetVariable("progress", "88")

	// 设置条件
	data.SetCondition("isActive", true)
	data.SetCondition("isComplete", false)

	// 设置团队成员数据
	teamMembers := []interface{}{
		map[string]interface{}{
			"name": "张开发",
			"role": "技术负责人",
			"work": "架构设计与核心开发",
		},
		map[string]interface{}{
			"name": "王测试",
			"role": "质量保证",
			"work": "功能测试与性能优化",
		},
		map[string]interface{}{
			"name": "刘设计",
			"role": "UI设计师",
			"work": "界面设计与用户体验",
		},
	}
	data.SetList("team", teamMembers)

	// 渲染复杂文档
	resultDoc, err := engine.RenderTemplateToDocument("complex_template", data)
	if err != nil {
		log.Fatalf("渲染复杂模板失败: %v", err)
	}

	// 保存结果
	outputFile := "examples/output/complex_result_" + time.Now().Format("20060102_150405") + ".docx"
	err = resultDoc.Save(outputFile)
	if err != nil {
		log.Fatalf("保存复杂结果失败: %v", err)
	}

	fmt.Printf("✓ 生成复杂项目报告: %s\n", outputFile)
}
