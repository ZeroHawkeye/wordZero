// Package main 演示动态创建复杂模板文件并渲染
package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ZeroHawkeye/wordZero/pkg/document"
)

func main() {
	fmt.Println("=== 动态创建复杂模板并渲染演示 ===")

	// 确保输出目录存在
	err := os.MkdirAll("examples/output", 0755)
	if err != nil {
		log.Fatalf("创建输出目录失败: %v", err)
	}

	// 1. 创建复杂的模板文档
	fmt.Println("🎨 创建复杂模板文档...")
	templateDoc := createComplexTemplate()

	// 保存模板文档供参考
	templateFile := "examples/output/generated_complex_template.docx"
	err = templateDoc.Save(templateFile)
	if err != nil {
		log.Fatalf("保存模板文档失败: %v", err)
	}
	fmt.Printf("✓ 模板文档已保存: %s\n", templateFile)

	// 2. 创建模板引擎
	engine := document.NewTemplateEngine()

	// 3. 从文档加载模板
	_, err = engine.LoadTemplateFromDocument("complex_report_template", templateDoc)
	if err != nil {
		log.Fatalf("加载模板失败: %v", err)
	}
	fmt.Println("✓ 模板加载成功")

	// 4. 准备渲染数据
	fmt.Println("📊 准备渲染数据...")
	data := prepareTemplateData()

	fmt.Printf("   - 基础变量: %d 个\n", len(data.Variables))
	fmt.Printf("   - 列表数据: %d 个\n", len(data.Lists))
	fmt.Printf("   - 条件变量: %d 个\n", len(data.Conditions))

	// 5. 渲染模板
	fmt.Println("🔄 开始渲染模板...")
	resultDoc, err := engine.RenderTemplateToDocument("complex_report_template", data)
	if err != nil {
		log.Fatalf("渲染模板失败: %v", err)
	}

	// 6. 保存结果文档
	timestamp := time.Now().Format("20060102_150405")
	outputFile := fmt.Sprintf("examples/output/complex_report_result_%s.docx", timestamp)
	err = resultDoc.Save(outputFile)
	if err != nil {
		log.Fatalf("保存文档失败: %v", err)
	}

	fmt.Printf("✅ 渲染完成！输出文件: %s\n", outputFile)

	// 显示文件信息
	if fileInfo, err := os.Stat(outputFile); err == nil {
		fmt.Printf("📄 文件大小: %.2f KB\n", float64(fileInfo.Size())/1024)
	}

	fmt.Println("\n📋 功能说明:")
	fmt.Println("   ✨ 动态创建复杂模板文档，无需依赖外部模板文件")
	fmt.Println("   🎨 自定义多种字体样式：标题、副标题、正文、强调文本")
	fmt.Println("   📊 包含样式化表格：项目任务表、团队成员表、统计数据表")
	fmt.Println("   🔄 支持条件渲染和循环渲染")
	fmt.Println("   📝 完整的文档结构：封面、目录、正文、附录")
	fmt.Println("   💼 企业级样式：专业配色、统一字体、规范布局")
}

// createComplexTemplate 创建复杂的模板文档
func createComplexTemplate() *document.Document {
	doc := document.New()

	// === 文档封面 ===
	createDocumentCover(doc)

	// === 项目概述部分 ===
	createProjectOverview(doc)

	// === 项目进度部分 ===
	createProjectProgress(doc)

	// === 团队成员部分 ===
	createTeamSection(doc)

	// === 任务列表部分 ===
	createTaskSection(doc)

	// === 里程碑部分 ===
	createMilestoneSection(doc)

	// === 统计数据部分 ===
	createStatisticsSection(doc)

	// === 附加信息部分 ===
	createAdditionalInfo(doc)

	return doc
}

// createDocumentCover 创建文档封面
func createDocumentCover(doc *document.Document) {
	fmt.Println("   📄 创建文档封面...")

	// 主标题 - 大号蓝色粗体
	titlePara := doc.AddParagraph("")
	titleRun := &document.Run{
		Properties: &document.RunProperties{
			Bold:       &document.Bold{},
			FontSize:   &document.FontSize{Val: "56"},  // 28磅
			Color:      &document.Color{Val: "1F4E79"}, // 深蓝色
			FontFamily: &document.FontFamily{ASCII: "Microsoft YaHei"},
		},
		Text: document.Text{Content: "{{title}}"},
	}
	titlePara.Runs = []document.Run{*titleRun}
	titlePara.SetAlignment(document.AlignCenter)
	titlePara.SetSpacing(&document.SpacingConfig{
		BeforePara: 144, // 2英寸
		AfterPara:  72,  // 1英寸
	})

	// 副标题 - 中号灰色斜体
	subtitlePara := doc.AddParagraph("")
	subtitleRun := &document.Run{
		Properties: &document.RunProperties{
			Italic:     &document.Italic{},
			FontSize:   &document.FontSize{Val: "32"},  // 16磅
			Color:      &document.Color{Val: "5B9BD5"}, // 浅蓝色
			FontFamily: &document.FontFamily{ASCII: "Microsoft YaHei"},
		},
		Text: document.Text{Content: "{{subtitle}}"},
	}
	subtitlePara.Runs = []document.Run{*subtitleRun}
	subtitlePara.SetAlignment(document.AlignCenter)
	subtitlePara.SetSpacing(&document.SpacingConfig{
		AfterPara: 36, // 0.5英寸
	})

	// 公司信息 - 标准字体
	companyPara := doc.AddParagraph("")
	companyRun := &document.Run{
		Properties: &document.RunProperties{
			FontSize:   &document.FontSize{Val: "24"},  // 12磅
			Color:      &document.Color{Val: "70AD47"}, // 绿色
			FontFamily: &document.FontFamily{ASCII: "Calibri"},
		},
		Text: document.Text{Content: "{{company}}"},
	}
	companyPara.Runs = []document.Run{*companyRun}
	companyPara.SetAlignment(document.AlignCenter)
	companyPara.SetSpacing(&document.SpacingConfig{
		AfterPara: 72, // 1英寸
	})

	// 作者和日期信息
	authorDatePara := doc.AddParagraph("")
	authorRun := &document.Run{
		Properties: &document.RunProperties{
			FontSize:   &document.FontSize{Val: "20"},  // 10磅
			Color:      &document.Color{Val: "7F7F7F"}, // 灰色
			FontFamily: &document.FontFamily{ASCII: "Calibri"},
		},
		Text: document.Text{Content: "编写：{{author}} | 日期：{{date}} | 版本：{{version}}"},
	}
	authorDatePara.Runs = []document.Run{*authorRun}
	authorDatePara.SetAlignment(document.AlignCenter)

	// 添加分页符
	doc.AddPageBreak()
}

// createProjectOverview 创建项目概述部分
func createProjectOverview(doc *document.Document) {
	fmt.Println("   📋 创建项目概述部分...")

	// 章节标题
	doc.AddHeadingParagraph("项目概述", 1)

	// 项目基本信息段落
	infoPara := doc.AddParagraph("")

	// 项目名称
	nameRun := &document.Run{
		Properties: &document.RunProperties{
			Bold:       &document.Bold{},
			FontSize:   &document.FontSize{Val: "22"}, // 11磅
			Color:      &document.Color{Val: "1F4E79"},
			FontFamily: &document.FontFamily{ASCII: "Microsoft YaHei"},
		},
		Text: document.Text{Content: "项目名称：{{projectName}}\n"},
	}

	// 项目经理
	managerRun := &document.Run{
		Properties: &document.RunProperties{
			FontSize:   &document.FontSize{Val: "22"}, // 11磅
			FontFamily: &document.FontFamily{ASCII: "Calibri"},
		},
		Text: document.Text{Content: "项目经理："},
	}

	managerValueRun := &document.Run{
		Properties: &document.RunProperties{
			Bold:       &document.Bold{},
			FontSize:   &document.FontSize{Val: "22"},  // 11磅
			Color:      &document.Color{Val: "E74C3C"}, // 红色强调
			FontFamily: &document.FontFamily{ASCII: "Calibri"},
		},
		Text: document.Text{Content: "{{manager}}\n"},
	}

	// 报告日期
	dateRun := &document.Run{
		Properties: &document.RunProperties{
			FontSize:   &document.FontSize{Val: "22"}, // 11磅
			FontFamily: &document.FontFamily{ASCII: "Calibri"},
		},
		Text: document.Text{Content: "报告日期：{{reportDate}}\n"},
	}

	// 项目状态（条件渲染）
	statusRun := &document.Run{
		Properties: &document.RunProperties{
			FontSize:   &document.FontSize{Val: "22"}, // 11磅
			FontFamily: &document.FontFamily{ASCII: "Calibri"},
		},
		Text: document.Text{Content: "项目状态："},
	}

	statusValueRun := &document.Run{
		Properties: &document.RunProperties{
			Bold:       &document.Bold{},
			FontSize:   &document.FontSize{Val: "22"},  // 11磅
			Color:      &document.Color{Val: "70AD47"}, // 绿色
			FontFamily: &document.FontFamily{ASCII: "Calibri"},
		},
		Text: document.Text{Content: "{{#if isActive}}进行中{{/if}}{{#if isComplete}}已完成{{/if}}{{#if needsAttention}}需要关注{{/if}}\n"},
	}

	// 完成进度
	progressRun := &document.Run{
		Properties: &document.RunProperties{
			FontSize:   &document.FontSize{Val: "22"}, // 11磅
			FontFamily: &document.FontFamily{ASCII: "Calibri"},
		},
		Text: document.Text{Content: "完成进度："},
	}

	progressValueRun := &document.Run{
		Properties: &document.RunProperties{
			Bold:       &document.Bold{},
			FontSize:   &document.FontSize{Val: "24"},  // 12磅
			Color:      &document.Color{Val: "F39C12"}, // 橙色
			FontFamily: &document.FontFamily{ASCII: "Calibri"},
		},
		Text: document.Text{Content: "{{progress}}%"},
	}

	infoPara.Runs = []document.Run{*nameRun, *managerRun, *managerValueRun, *dateRun, *statusRun, *statusValueRun, *progressRun, *progressValueRun}
}

// createProjectProgress 创建项目进度部分
func createProjectProgress(doc *document.Document) {
	fmt.Println("   📈 创建项目进度部分...")

	doc.AddHeadingParagraph("项目进度分析", 1)

	// 进度描述段落
	progressPara := doc.AddParagraph("")

	introRun := &document.Run{
		Properties: &document.RunProperties{
			FontSize:   &document.FontSize{Val: "22"}, // 11磅
			FontFamily: &document.FontFamily{ASCII: "Calibri"},
		},
		Text: document.Text{Content: "当前项目进度为 "},
	}

	percentRun := &document.Run{
		Properties: &document.RunProperties{
			Bold:       &document.Bold{},
			FontSize:   &document.FontSize{Val: "28"},  // 14磅
			Color:      &document.Color{Val: "E74C3C"}, // 红色
			FontFamily: &document.FontFamily{ASCII: "Arial"},
		},
		Text: document.Text{Content: "{{progress}}%"},
	}

	endRun := &document.Run{
		Properties: &document.RunProperties{
			FontSize:   &document.FontSize{Val: "22"}, // 11磅
			FontFamily: &document.FontFamily{ASCII: "Calibri"},
		},
		Text: document.Text{Content: "，项目整体进展"},
	}

	statusRun := &document.Run{
		Properties: &document.RunProperties{
			Bold:       &document.Bold{},
			FontSize:   &document.FontSize{Val: "22"},  // 11磅
			Color:      &document.Color{Val: "70AD47"}, // 绿色
			FontFamily: &document.FontFamily{ASCII: "Calibri"},
		},
		Text: document.Text{Content: "{{#if isActive}}顺利{{/if}}{{#if needsAttention}}需要关注{{/if}}"},
	}

	finalRun := &document.Run{
		Properties: &document.RunProperties{
			FontSize:   &document.FontSize{Val: "22"}, // 11磅
			FontFamily: &document.FontFamily{ASCII: "Calibri"},
		},
		Text: document.Text{Content: "。"},
	}

	progressPara.Runs = []document.Run{*introRun, *percentRun, *endRun, *statusRun, *finalRun}
}

// createTeamSection 创建团队成员部分
func createTeamSection(doc *document.Document) {
	fmt.Println("   👥 创建团队成员部分...")

	doc.AddHeadingParagraph("团队成员", 1)

	// 创建团队成员表格
	teamTableConfig := &document.TableConfig{
		Rows:  2, // 表头 + 模板行
		Cols:  4, // 姓名、角色、工作内容、是否负责人
		Width: 9000,
	}
	teamTable, _ := doc.CreateTable(teamTableConfig)

	// 设置表头
	headers := []string{"姓名", "角色", "工作内容", "负责人"}
	headerFormat := &document.TextFormat{
		Bold:       true,
		FontSize:   11,
		FontColor:  "FFFFFF",
		FontFamily: "Microsoft YaHei",
	}

	for i, header := range headers {
		teamTable.SetCellFormattedText(0, i, header, headerFormat)
		// 设置表头背景色 - 深蓝色
		teamTable.SetCellShading(0, i, &document.ShadingConfig{
			Pattern:         document.ShadingPatternClear,
			BackgroundColor: "1F4E79",
		})
	}

	// 设置模板行（包含循环语法）
	teamTable.SetCellText(1, 0, "{{#each team}}{{name}}{{/each}}")
	teamTable.SetCellText(1, 1, "{{#each team}}{{role}}{{/each}}")
	teamTable.SetCellText(1, 2, "{{#each team}}{{work}}{{/each}}")
	teamTable.SetCellText(1, 3, "{{#each team}}{{#if isLeader}}是{{else}}否{{/if}}{{/each}}")

	doc.Body.AddElement(teamTable)
}

// createTaskSection 创建任务列表部分
func createTaskSection(doc *document.Document) {
	fmt.Println("   📝 创建任务列表部分...")

	doc.AddHeadingParagraph("项目任务", 1)

	// 创建任务表格
	taskTableConfig := &document.TableConfig{
		Rows:  2, // 表头 + 模板行
		Cols:  5, // 任务名称、状态、进度、负责人、优先级
		Width: 10000,
	}
	taskTable, _ := doc.CreateTable(taskTableConfig)

	// 设置表头
	taskHeaders := []string{"任务名称", "状态", "进度", "负责人", "优先级"}
	taskHeaderFormat := &document.TextFormat{
		Bold:       true,
		FontSize:   11,
		FontColor:  "FFFFFF",
		FontFamily: "Microsoft YaHei",
	}

	for i, header := range taskHeaders {
		taskTable.SetCellFormattedText(0, i, header, taskHeaderFormat)
		// 设置表头背景色 - 深绿色
		taskTable.SetCellShading(0, i, &document.ShadingConfig{
			Pattern:         document.ShadingPatternClear,
			BackgroundColor: "70AD47",
		})
	}

	// 设置模板行
	taskTable.SetCellText(1, 0, "{{#each tasks}}{{name}}{{/each}}")
	taskTable.SetCellText(1, 1, "{{#each tasks}}{{status}}{{/each}}")
	taskTable.SetCellText(1, 2, "{{#each tasks}}{{progress}}%{{/each}}")
	taskTable.SetCellText(1, 3, "{{#each tasks}}{{responsible}}{{/each}}")
	taskTable.SetCellText(1, 4, "{{#each tasks}}{{priority}}{{/each}}")

	doc.Body.AddElement(taskTable)
}

// createMilestoneSection 创建里程碑部分
func createMilestoneSection(doc *document.Document) {
	fmt.Println("   🎯 创建里程碑部分...")

	doc.AddHeadingParagraph("项目里程碑", 1)

	// 创建里程碑表格
	milestoneTableConfig := &document.TableConfig{
		Rows:  2, // 表头 + 模板行
		Cols:  4, // 里程碑名称、日期、状态、是否完成
		Width: 9000,
	}
	milestoneTable, _ := doc.CreateTable(milestoneTableConfig)

	// 设置表头
	milestoneHeaders := []string{"里程碑", "计划日期", "状态", "完成"}
	milestoneHeaderFormat := &document.TextFormat{
		Bold:       true,
		FontSize:   11,
		FontColor:  "FFFFFF",
		FontFamily: "Microsoft YaHei",
	}

	for i, header := range milestoneHeaders {
		milestoneTable.SetCellFormattedText(0, i, header, milestoneHeaderFormat)
		// 设置表头背景色 - 橙色
		milestoneTable.SetCellShading(0, i, &document.ShadingConfig{
			Pattern:         document.ShadingPatternClear,
			BackgroundColor: "F39C12",
		})
	}

	// 设置模板行
	milestoneTable.SetCellText(1, 0, "{{#each milestones}}{{name}}{{/each}}")
	milestoneTable.SetCellText(1, 1, "{{#each milestones}}{{date}}{{/each}}")
	milestoneTable.SetCellText(1, 2, "{{#each milestones}}{{status}}{{/each}}")
	milestoneTable.SetCellText(1, 3, "{{#each milestones}}{{#if completed}}✓{{/if}}{{#if notCompleted}}○{{/if}}{{/each}}")

	doc.Body.AddElement(milestoneTable)
}

// createStatisticsSection 创建统计数据部分
func createStatisticsSection(doc *document.Document) {
	fmt.Println("   📊 创建统计数据部分...")

	doc.AddHeadingParagraph("项目统计", 1)

	// 创建统计表格
	statsTableConfig := &document.TableConfig{
		Rows:  2, // 表头 + 模板行
		Cols:  3, // 指标、数值、单位
		Width: 8000,
	}
	statsTable, _ := doc.CreateTable(statsTableConfig)

	// 设置表头
	statsHeaders := []string{"统计指标", "数值", "单位"}
	statsHeaderFormat := &document.TextFormat{
		Bold:       true,
		FontSize:   11,
		FontColor:  "FFFFFF",
		FontFamily: "Microsoft YaHei",
	}

	for i, header := range statsHeaders {
		statsTable.SetCellFormattedText(0, i, header, statsHeaderFormat)
		// 设置表头背景色 - 紫色
		statsTable.SetCellShading(0, i, &document.ShadingConfig{
			Pattern:         document.ShadingPatternClear,
			BackgroundColor: "8E44AD",
		})
	}

	// 设置模板行
	statsTable.SetCellText(1, 0, "{{#each statistics}}{{metric}}{{/each}}")
	statsTable.SetCellText(1, 1, "{{#each statistics}}{{value}}{{/each}}")
	statsTable.SetCellText(1, 2, "{{#each statistics}}{{unit}}{{/each}}")

	doc.Body.AddElement(statsTable)
}

// createAdditionalInfo 创建附加信息部分
func createAdditionalInfo(doc *document.Document) {
	fmt.Println("   💼 创建附加信息部分...")

	doc.AddHeadingParagraph("附加信息", 1)

	// 创建混合格式的总结段落
	summaryPara := doc.AddParagraph("")

	part1Run := &document.Run{
		Properties: &document.RunProperties{
			FontSize:   &document.FontSize{Val: "22"}, // 11磅
			FontFamily: &document.FontFamily{ASCII: "Calibri"},
		},
		Text: document.Text{Content: "本报告生成于 "},
	}

	dateRun := &document.Run{
		Properties: &document.RunProperties{
			Bold:       &document.Bold{},
			FontSize:   &document.FontSize{Val: "22"},  // 11磅
			Color:      &document.Color{Val: "E74C3C"}, // 红色
			FontFamily: &document.FontFamily{ASCII: "Calibri"},
		},
		Text: document.Text{Content: "{{reportDate}}"},
	}

	part2Run := &document.Run{
		Properties: &document.RunProperties{
			FontSize:   &document.FontSize{Val: "22"}, // 11磅
			FontFamily: &document.FontFamily{ASCII: "Calibri"},
		},
		Text: document.Text{Content: "，由 "},
	}

	managerRun := &document.Run{
		Properties: &document.RunProperties{
			Bold:       &document.Bold{},
			FontSize:   &document.FontSize{Val: "22"},  // 11磅
			Color:      &document.Color{Val: "1F4E79"}, // 蓝色
			FontFamily: &document.FontFamily{ASCII: "Calibri"},
		},
		Text: document.Text{Content: "{{manager}}"},
	}

	part3Run := &document.Run{
		Properties: &document.RunProperties{
			FontSize:   &document.FontSize{Val: "22"}, // 11磅
			FontFamily: &document.FontFamily{ASCII: "Calibri"},
		},
		Text: document.Text{Content: " 负责 "},
	}

	projectRun := &document.Run{
		Properties: &document.RunProperties{
			Italic:     &document.Italic{},
			FontSize:   &document.FontSize{Val: "22"},  // 11磅
			Color:      &document.Color{Val: "70AD47"}, // 绿色
			FontFamily: &document.FontFamily{ASCII: "Calibri"},
		},
		Text: document.Text{Content: "{{projectName}}"},
	}

	part4Run := &document.Run{
		Properties: &document.RunProperties{
			FontSize:   &document.FontSize{Val: "22"}, // 11磅
			FontFamily: &document.FontFamily{ASCII: "Calibri"},
		},
		Text: document.Text{Content: " 项目。感谢团队成员的辛勤工作和贡献！"},
	}

	summaryPara.Runs = []document.Run{*part1Run, *dateRun, *part2Run, *managerRun, *part3Run, *projectRun, *part4Run}

	// 添加版权信息
	doc.AddParagraph("\n")
	copyrightPara := doc.AddParagraph("")
	copyrightRun := &document.Run{
		Properties: &document.RunProperties{
			FontSize:   &document.FontSize{Val: "18"},  // 9磅
			Color:      &document.Color{Val: "7F7F7F"}, // 灰色
			FontFamily: &document.FontFamily{ASCII: "Calibri"},
		},
		Text: document.Text{Content: "© 2025 {{company}} 版权所有 | 生成版本：{{version}}"},
	}
	copyrightPara.Runs = []document.Run{*copyrightRun}
	copyrightPara.SetAlignment(document.AlignCenter)
}

// prepareTemplateData 准备模板渲染数据
func prepareTemplateData() *document.TemplateData {
	data := document.NewTemplateData()

	// 设置基础变量
	data.SetVariable("projectName", "WordZero 企业文档管理系统")
	data.SetVariable("title", "项目进展报告")
	data.SetVariable("subtitle", "月度总结与下阶段规划")
	data.SetVariable("company", "WordZero 科技有限公司")
	data.SetVariable("author", "项目管理部")
	data.SetVariable("manager", "李项目经理")
	data.SetVariable("reportDate", time.Now().Format("2006年01月02日"))
	data.SetVariable("date", time.Now().Format("2006年01月02日"))
	data.SetVariable("progress", "88")
	data.SetVariable("version", "v1.3.5")

	// 设置条件变量
	data.SetCondition("isActive", true)
	data.SetCondition("isComplete", false)
	data.SetCondition("needsAttention", true)

	// 设置团队成员列表
	teamMembers := []interface{}{
		map[string]interface{}{
			"name":     "张开发",
			"role":     "技术负责人",
			"work":     "架构设计与核心开发",
			"isLeader": true,
		},
		map[string]interface{}{
			"name":     "王测试",
			"role":     "质量保证",
			"work":     "功能测试与性能优化",
			"isLeader": false,
		},
		map[string]interface{}{
			"name":     "刘设计",
			"role":     "UI设计师",
			"work":     "界面设计与用户体验",
			"isLeader": false,
		},
		map[string]interface{}{
			"name":     "陈产品",
			"role":     "产品经理",
			"work":     "需求分析与产品规划",
			"isLeader": false,
		},
	}
	data.SetList("team", teamMembers)

	// 设置项目任务列表
	tasks := []interface{}{
		map[string]interface{}{
			"name":        "模板功能开发",
			"status":      "已完成",
			"progress":    "100",
			"responsible": "张开发",
			"priority":    "高",
		},
		map[string]interface{}{
			"name":        "样式保持修复",
			"status":      "已完成",
			"progress":    "100",
			"responsible": "王测试",
			"priority":    "高",
		},
		map[string]interface{}{
			"name":        "用户界面优化",
			"status":      "进行中",
			"progress":    "75",
			"responsible": "刘设计",
			"priority":    "中",
		},
		map[string]interface{}{
			"name":        "文档完善",
			"status":      "进行中",
			"progress":    "60",
			"responsible": "陈产品",
			"priority":    "中",
		},
		map[string]interface{}{
			"name":        "性能优化",
			"status":      "计划中",
			"progress":    "0",
			"responsible": "张开发",
			"priority":    "低",
		},
	}
	data.SetList("tasks", tasks)

	// 设置里程碑列表
	milestones := []interface{}{
		map[string]interface{}{
			"name":         "需求分析",
			"date":         "2025年01月01日",
			"status":       "已完成",
			"completed":    true,
			"notCompleted": false,
		},
		map[string]interface{}{
			"name":         "系统设计",
			"date":         "2025年01月10日",
			"status":       "已完成",
			"completed":    true,
			"notCompleted": false,
		},
		map[string]interface{}{
			"name":         "核心开发",
			"date":         "2025年01月18日",
			"status":       "已完成",
			"completed":    true,
			"notCompleted": false,
		},
		map[string]interface{}{
			"name":         "功能测试",
			"date":         "2025年01月25日",
			"status":       "进行中",
			"completed":    false,
			"notCompleted": true,
		},
		map[string]interface{}{
			"name":         "系统集成",
			"date":         "2025年02月01日",
			"status":       "计划中",
			"completed":    false,
			"notCompleted": true,
		},
		map[string]interface{}{
			"name":         "上线部署",
			"date":         "2025年02月15日",
			"status":       "计划中",
			"completed":    false,
			"notCompleted": true,
		},
	}
	data.SetList("milestones", milestones)

	// 设置项目统计数据
	statistics := []interface{}{
		map[string]interface{}{
			"metric": "代码行数",
			"value":  "15,000+",
			"unit":   "行",
		},
		map[string]interface{}{
			"metric": "测试覆盖率",
			"value":  "95",
			"unit":   "%",
		},
		map[string]interface{}{
			"metric": "文档完成度",
			"value":  "90",
			"unit":   "%",
		},
		map[string]interface{}{
			"metric": "Bug修复率",
			"value":  "98",
			"unit":   "%",
		},
		map[string]interface{}{
			"metric": "团队成员",
			"value":  "4",
			"unit":   "人",
		},
	}
	data.SetList("statistics", statistics)

	return data
}
