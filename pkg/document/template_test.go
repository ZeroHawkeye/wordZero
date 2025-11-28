// Package document 模板功能测试
package document

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNewTemplateEngine 测试创建模板引擎
func TestNewTemplateEngine(t *testing.T) {
	engine := NewTemplateEngine()
	if engine == nil {
		t.Fatal("Expected template engine to be created")
	}

	if engine.cache == nil {
		t.Fatal("Expected cache to be initialized")
	}
}

// TestTemplateVariableReplacement 测试变量替换功能
func TestTemplateVariableReplacement(t *testing.T) {
	engine := NewTemplateEngine()

	// 创建包含变量的模板
	templateContent := "Hello {{name}}, welcome to {{company}}!"
	template, err := engine.LoadTemplate("test_template", templateContent)
	if err != nil {
		t.Fatalf("Failed to load template: %v", err)
	}

	// 验证模板变量解析
	if len(template.Variables) != 2 {
		t.Errorf("Expected 2 variables, got %d", len(template.Variables))
	}

	if _, exists := template.Variables["name"]; !exists {
		t.Error("Expected 'name' variable to be found")
	}

	if _, exists := template.Variables["company"]; !exists {
		t.Error("Expected 'company' variable to be found")
	}

	// 创建模板数据
	data := NewTemplateData()
	data.SetVariable("name", "张三")
	data.SetVariable("company", "WordZero公司")

	// 渲染模板
	doc, err := engine.RenderToDocument("test_template", data)
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	if doc == nil {
		t.Fatal("Expected document to be created")
	}

	// 检查文档内容
	if len(doc.Body.Elements) == 0 {
		t.Error("Expected document to have content")
	}
}

// TestTemplateConditionalStatements 测试条件语句功能
func TestTemplateConditionalStatements(t *testing.T) {
	engine := NewTemplateEngine()

	// 创建包含条件语句的模板
	templateContent := `{{#if showWelcome}}欢迎使用WordZero！{{/if}}
{{#if showDescription}}这是一个强大的Word文档操作库。{{/if}}`

	template, err := engine.LoadTemplate("conditional_template", templateContent)
	if err != nil {
		t.Fatalf("Failed to load template: %v", err)
	}

	// 验证条件块解析
	if len(template.Blocks) < 2 {
		t.Errorf("Expected at least 2 blocks, got %d", len(template.Blocks))
	}

	// 测试条件为真的情况
	data := NewTemplateData()
	data.SetCondition("showWelcome", true)
	data.SetCondition("showDescription", false)

	doc, err := engine.RenderToDocument("conditional_template", data)
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	if doc == nil {
		t.Fatal("Expected document to be created")
	}
}

// TestTemplateLoopStatements 测试循环语句功能
func TestTemplateLoopStatements(t *testing.T) {
	engine := NewTemplateEngine()

	// 创建包含循环语句的模板
	templateContent := `产品列表：
{{#each products}}
- 产品名称：{{name}}，价格：{{price}}元
{{/each}}`

	template, err := engine.LoadTemplate("loop_template", templateContent)
	if err != nil {
		t.Fatalf("Failed to load template: %v", err)
	}

	// 验证循环块解析
	foundEachBlock := false
	for _, block := range template.Blocks {
		if block.Type == "each" && block.Variable == "products" {
			foundEachBlock = true
			break
		}
	}

	if !foundEachBlock {
		t.Error("Expected to find 'each products' block")
	}

	// 创建列表数据
	data := NewTemplateData()
	products := []interface{}{
		map[string]interface{}{
			"name":  "iPhone",
			"price": 8999,
		},
		map[string]interface{}{
			"name":  "iPad",
			"price": 5999,
		},
	}
	data.SetList("products", products)

	doc, err := engine.RenderToDocument("loop_template", data)
	if err != nil {
		t.Fatalf("Failed to render template: %v", err)
	}

	if doc == nil {
		t.Fatal("Expected document to be created")
	}
}

// TestTemplateInheritance 测试模板继承功能
func TestTemplateInheritance(t *testing.T) {
	engine := NewTemplateEngine()

	// 创建基础模板
	baseTemplateContent := `文档标题：{{title}}
基础内容：这是基础模板的内容。`

	_, err := engine.LoadTemplate("base_template", baseTemplateContent)
	if err != nil {
		t.Fatalf("Failed to load base template: %v", err)
	}

	// 创建继承模板
	childTemplateContent := `{{extends "base_template"}}
扩展内容：这是子模板的内容。`

	childTemplate, err := engine.LoadTemplate("child_template", childTemplateContent)
	if err != nil {
		t.Fatalf("Failed to load child template: %v", err)
	}

	// 验证继承关系
	if childTemplate.Parent == nil {
		t.Error("Expected child template to have parent")
	}

	if childTemplate.Parent.Name != "base_template" {
		t.Errorf("Expected parent template name to be 'base_template', got '%s'", childTemplate.Parent.Name)
	}
}

// TestTemplateValidation 测试模板验证功能
func TestTemplateValidation(t *testing.T) {
	engine := NewTemplateEngine()

	// 测试有效模板
	validTemplate := `Hello {{name}}!
{{#if showMessage}}This is a message.{{/if}}
{{#each items}}Item: {{this}}{{/each}}`

	template, err := engine.LoadTemplate("valid_template", validTemplate)
	if err != nil {
		t.Fatalf("Failed to load valid template: %v", err)
	}

	err = engine.ValidateTemplate(template)
	if err != nil {
		t.Errorf("Expected valid template to pass validation, got error: %v", err)
	}

	// 测试无效模板 - 括号不匹配
	invalidTemplate1 := `Hello {{name}!`
	template1, err := engine.LoadTemplate("invalid_template1", invalidTemplate1)
	if err != nil {
		t.Fatalf("Failed to load invalid template: %v", err)
	}

	err = engine.ValidateTemplate(template1)
	if err == nil {
		t.Error("Expected invalid template (mismatched brackets) to fail validation")
	}

	// 测试无效模板 - if语句不匹配
	invalidTemplate2 := `{{#if condition}}Hello`
	template2, err := engine.LoadTemplate("invalid_template2", invalidTemplate2)
	if err != nil {
		t.Fatalf("Failed to load invalid template: %v", err)
	}

	err = engine.ValidateTemplate(template2)
	if err == nil {
		t.Error("Expected invalid template (mismatched if statements) to fail validation")
	}
}

// TestTemplateData 测试模板数据功能
func TestTemplateData(t *testing.T) {
	data := NewTemplateData()

	// 测试设置和获取变量
	data.SetVariable("name", "测试")
	value, exists := data.GetVariable("name")
	if !exists {
		t.Error("Expected variable 'name' to exist")
	}
	if value != "测试" {
		t.Errorf("Expected variable value to be '测试', got '%v'", value)
	}

	// 测试设置和获取列表
	items := []interface{}{"item1", "item2", "item3"}
	data.SetList("items", items)
	list, exists := data.GetList("items")
	if !exists {
		t.Error("Expected list 'items' to exist")
	}
	if len(list) != 3 {
		t.Errorf("Expected list length to be 3, got %d", len(list))
	}

	// 测试设置和获取条件
	data.SetCondition("enabled", true)
	condition, exists := data.GetCondition("enabled")
	if !exists {
		t.Error("Expected condition 'enabled' to exist")
	}
	if !condition {
		t.Error("Expected condition value to be true")
	}

	// 测试批量设置变量
	variables := map[string]interface{}{
		"title":   "测试标题",
		"content": "测试内容",
	}
	data.SetVariables(variables)

	title, exists := data.GetVariable("title")
	if !exists || title != "测试标题" {
		t.Error("Expected batch set variables to work")
	}
}

// TestTemplateDataFromStruct 测试从结构体创建模板数据
func TestTemplateDataFromStruct(t *testing.T) {
	type TestStruct struct {
		Name    string
		Age     int
		Enabled bool
	}

	testData := TestStruct{
		Name:    "张三",
		Age:     30,
		Enabled: true,
	}

	templateData := NewTemplateData()
	err := templateData.FromStruct(testData)
	if err != nil {
		t.Fatalf("Failed to create template data from struct: %v", err)
	}

	// 验证变量是否正确设置
	name, exists := templateData.GetVariable("name")
	if !exists || name != "张三" {
		t.Error("Expected 'name' variable to be set correctly")
	}

	age, exists := templateData.GetVariable("age")
	if !exists || age != 30 {
		t.Error("Expected 'age' variable to be set correctly")
	}

	enabled, exists := templateData.GetVariable("enabled")
	if !exists || enabled != true {
		t.Error("Expected 'enabled' variable to be set correctly")
	}
}

// TestTemplateMerge 测试模板数据合并
func TestTemplateMerge(t *testing.T) {
	data1 := NewTemplateData()
	data1.SetVariable("name", "张三")
	data1.SetCondition("enabled", true)

	data2 := NewTemplateData()
	data2.SetVariable("age", 30)
	data2.SetList("items", []interface{}{"item1", "item2"})

	// 合并数据
	data1.Merge(data2)

	// 验证合并结果
	name, exists := data1.GetVariable("name")
	if !exists || name != "张三" {
		t.Error("Expected original variable to remain")
	}

	age, exists := data1.GetVariable("age")
	if !exists || age != 30 {
		t.Error("Expected merged variable to be present")
	}

	enabled, exists := data1.GetCondition("enabled")
	if !exists || !enabled {
		t.Error("Expected original condition to remain")
	}

	items, exists := data1.GetList("items")
	if !exists || len(items) != 2 {
		t.Error("Expected merged list to be present")
	}
}

// TestTemplateCache 测试模板缓存功能
func TestTemplateCache(t *testing.T) {
	engine := NewTemplateEngine()

	// 加载模板
	templateContent := "Hello {{name}}!"
	template1, err := engine.LoadTemplate("cached_template", templateContent)
	if err != nil {
		t.Fatalf("Failed to load template: %v", err)
	}

	// 从缓存获取模板
	template2, err := engine.GetTemplate("cached_template")
	if err != nil {
		t.Fatalf("Failed to get template from cache: %v", err)
	}

	// 验证是同一个模板实例
	if template1 != template2 {
		t.Error("Expected to get same template instance from cache")
	}

	// 清空缓存
	engine.ClearCache()

	// 尝试获取已清空的模板
	_, err = engine.GetTemplate("cached_template")
	if err == nil {
		t.Error("Expected error when getting template after cache clear")
	}
}

// TestComplexTemplateRendering 测试复杂模板渲染
func TestComplexTemplateRendering(t *testing.T) {
	engine := NewTemplateEngine()

	// 创建复杂模板
	complexTemplate := `报告标题：{{title}}
作者：{{author}}

{{#if showSummary}}
概要：{{summary}}
{{/if}}

详细内容：
{{#each sections}}
章节 {{@index}}: {{title}}
内容：{{content}}

{{/each}}

{{#if showFooter}}
报告完毕。
{{/if}}`

	_, err := engine.LoadTemplate("complex_template", complexTemplate)
	if err != nil {
		t.Fatalf("Failed to load complex template: %v", err)
	}

	// 创建复杂数据
	data := NewTemplateData()
	data.SetVariable("title", "WordZero功能测试报告")
	data.SetVariable("author", "开发团队")
	data.SetVariable("summary", "本报告测试了WordZero的模板功能。")

	data.SetCondition("showSummary", true)
	data.SetCondition("showFooter", true)

	sections := []interface{}{
		map[string]interface{}{
			"title":   "基础功能",
			"content": "测试了基础的文档操作功能。",
		},
		map[string]interface{}{
			"title":   "模板功能",
			"content": "测试了模板引擎的各种功能。",
		},
	}
	data.SetList("sections", sections)

	// 渲染复杂模板
	doc, err := engine.RenderToDocument("complex_template", data)
	if err != nil {
		t.Fatalf("Failed to render complex template: %v", err)
	}

	if doc == nil {
		t.Fatal("Expected document to be created")
	}

	// 验证文档有内容
	if len(doc.Body.Elements) == 0 {
		t.Error("Expected document to have content")
	}
}

// TestImagePlaceholder 测试图片占位符功能
func TestImagePlaceholder(t *testing.T) {
	engine := NewTemplateEngine()

	// 测试基础图片占位符解析
	t.Run("解析图片占位符", func(t *testing.T) {
		templateContent := `文档标题

这里有一个图片：
{{#image testImage}}

更多内容...`

		template, err := engine.LoadTemplate("image_test", templateContent)
		if err != nil {
			t.Fatalf("加载模板失败: %v", err)
		}

		// 检查是否正确解析了图片占位符
		hasImageBlock := false
		for _, block := range template.Blocks {
			if block.Type == "image" && block.Name == "testImage" {
				hasImageBlock = true
				break
			}
		}

		if !hasImageBlock {
			t.Error("模板应该包含图片块")
		}
	})

	// 测试图片占位符渲染（字符串模板）
	t.Run("渲染图片占位符到字符串", func(t *testing.T) {
		templateContent := `产品介绍：{{productName}}

产品图片：
{{#image productImage}}

描述：{{description}}`

		_, err := engine.LoadTemplate("product", templateContent)
		if err != nil {
			t.Fatalf("加载模板失败: %v", err)
		}

		data := NewTemplateData()
		data.SetVariable("productName", "测试产品")
		data.SetVariable("description", "这是一个测试产品")

		// 创建图片配置
		imageConfig := &ImageConfig{
			Position:  ImagePositionInline,
			Alignment: AlignCenter,
			Size: &ImageSize{
				Width:           100,
				KeepAspectRatio: true,
			},
			AltText: "测试图片",
			Title:   "测试产品图片",
		}

		// 设置图片数据（使用示例二进制数据）
		imageData := createTestImageData()
		data.SetImageFromData("productImage", imageData, imageConfig)

		// 渲染模板
		doc, err := engine.RenderToDocument("product", data)
		if err != nil {
			t.Fatalf("渲染模板失败: %v", err)
		}

		if doc == nil {
			t.Error("渲染结果不应为空")
		}
	})

	// 测试从文档模板渲染图片占位符
	t.Run("从文档模板渲染图片占位符", func(t *testing.T) {
		// 创建基础文档
		baseDoc := New()
		baseDoc.AddParagraph("报告标题：{{title}}")
		baseDoc.AddParagraph("{{#image reportChart}}")
		baseDoc.AddParagraph("总结：{{summary}}")

		// 从文档创建模板
		template, err := engine.LoadTemplateFromDocument("report_template", baseDoc)
		if err != nil {
			t.Fatalf("从文档创建模板失败: %v", err)
		}

		if len(template.Variables) == 0 {
			t.Error("模板应该包含变量")
		}

		// 准备数据
		data := NewTemplateData()
		data.SetVariable("title", "月度报告")
		data.SetVariable("summary", "数据显示增长趋势良好")

		chartConfig := &ImageConfig{
			Position:  ImagePositionInline,
			Alignment: AlignCenter,
			Size: &ImageSize{
				Width: 120,
			},
		}

		imageData := createTestImageData()
		data.SetImageFromData("reportChart", imageData, chartConfig)

		// 使用RenderTemplateToDocument方法（推荐用于文档模板）
		doc, err := engine.RenderTemplateToDocument("report_template", data)
		if err != nil {
			t.Fatalf("渲染文档模板失败: %v", err)
		}

		if doc == nil {
			t.Fatal("渲染结果不应为空")
		}

		// 检查文档中是否有元素
		if len(doc.Body.Elements) == 0 {
			t.Error("文档应该包含元素")
		}
	})

	// 测试图片数据管理方法
	t.Run("测试图片数据管理", func(t *testing.T) {
		data := NewTemplateData()

		// 测试SetImage方法
		config := &ImageConfig{
			Position: ImagePositionInline,
			Size:     &ImageSize{Width: 50},
		}
		data.SetImage("test1", "path/to/image.jpg", config)

		// 测试SetImageFromData方法
		imageData := createTestImageData()
		data.SetImageFromData("test2", imageData, config)

		// 测试SetImageWithDetails方法
		data.SetImageWithDetails("test3", "path/to/image2.jpg", imageData, config, "alt text", "title")

		// 测试GetImage方法
		img1, exists1 := data.GetImage("test1")
		if !exists1 || img1.FilePath != "path/to/image.jpg" {
			t.Error("图片1数据不正确")
		}

		img2, exists2 := data.GetImage("test2")
		if !exists2 || len(img2.Data) == 0 {
			t.Error("图片2数据不正确")
		}

		img3, exists3 := data.GetImage("test3")
		if !exists3 || img3.AltText != "alt text" || img3.Title != "title" {
			t.Error("图片3数据不正确")
		}

		// 测试不存在的图片
		_, exists4 := data.GetImage("nonexistent")
		if exists4 {
			t.Error("不存在的图片不应该返回true")
		}
	})

	// 测试图片占位符与其他模板语法的兼容性
	t.Run("图片占位符与其他语法兼容性", func(t *testing.T) {
		templateContent := `{{#if showImage}}
图片标题：{{imageTitle}}
{{#image dynamicImage}}
{{/if}}

{{#each items}}
项目：{{name}}
{{#image itemImage}}
描述：{{description}}
{{/each}}`

		_, err := engine.LoadTemplate("complex", templateContent)
		if err != nil {
			t.Fatalf("加载复杂模板失败: %v", err)
		}

		data := NewTemplateData()
		data.SetCondition("showImage", true)
		data.SetVariable("imageTitle", "主要图片")

		items := []interface{}{
			map[string]interface{}{
				"name":        "项目1",
				"description": "项目1描述",
			},
		}
		data.SetList("items", items)

		config := &ImageConfig{Position: ImagePositionInline}
		imageData := createTestImageData()
		data.SetImageFromData("dynamicImage", imageData, config)
		data.SetImageFromData("itemImage", imageData, config)

		// 渲染不应该出错
		doc, err := engine.RenderToDocument("complex", data)
		if err != nil {
			t.Fatalf("渲染复杂模板失败: %v", err)
		}

		if doc == nil {
			t.Error("渲染结果不应为空")
		}
	})
}

// createTestImageData 创建测试用的图片数据
func createTestImageData() []byte {
	// 创建一个最小的PNG图片数据用于测试
	return []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01,
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4,
		0x89, 0x00, 0x00, 0x00, 0x0A, 0x49, 0x44, 0x41,
		0x54, 0x78, 0x9C, 0x63, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0D, 0x0A, 0x2D, 0xB4, 0x00,
		0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE,
		0x42, 0x60, 0x82,
	}
}

// TestNestedLoops 测试嵌套循环功能
func TestNestedLoops(t *testing.T) {
	engine := NewTemplateEngine()

	// 创建包含嵌套循环的模板
	templateContent := `会议纪要

日期：{{date}}

参会人员：
{{#each attendees}}
- {{name}} ({{role}})
  任务清单：
  {{#each tasks}}
  * {{taskName}} - 状态: {{status}}
  {{/each}}
{{/each}}

会议总结：{{summary}}`

	template, err := engine.LoadTemplate("meeting_minutes", templateContent)
	if err != nil {
		t.Fatalf("Failed to load template with nested loops: %v", err)
	}

	if len(template.Blocks) < 1 {
		t.Error("Expected at least 1 block in template")
	}

	// 创建嵌套数据结构
	data := NewTemplateData()
	data.SetVariable("date", "2024-12-01")
	data.SetVariable("summary", "会议圆满结束")

	attendees := []interface{}{
		map[string]interface{}{
			"name": "张三",
			"role": "项目经理",
			"tasks": []interface{}{
				map[string]interface{}{
					"taskName": "制定项目计划",
					"status":   "进行中",
				},
				map[string]interface{}{
					"taskName": "分配资源",
					"status":   "已完成",
				},
			},
		},
		map[string]interface{}{
			"name": "李四",
			"role": "开发工程师",
			"tasks": []interface{}{
				map[string]interface{}{
					"taskName": "实现核心功能",
					"status":   "进行中",
				},
				map[string]interface{}{
					"taskName": "编写单元测试",
					"status":   "待开始",
				},
			},
		},
	}
	data.SetList("attendees", attendees)

	// 渲染模板
	doc, err := engine.RenderToDocument("meeting_minutes", data)
	if err != nil {
		t.Fatalf("Failed to render template with nested loops: %v", err)
	}

	if doc == nil {
		t.Fatal("Expected document to be created")
	}

	// 验证文档内容
	if len(doc.Body.Elements) == 0 {
		t.Error("Expected document to have content")
	}

	// 检查生成的内容是否包含预期的嵌套数据
	foundNestedContent := false
	for _, element := range doc.Body.Elements {
		if para, ok := element.(*Paragraph); ok {
			fullText := ""
			for _, run := range para.Runs {
				fullText += run.Text.Content
			}

			// 检查是否包含嵌套循环生成的内容（任务名称）
			if fullText == "  * 制定项目计划 - 状态: 进行中" ||
				fullText == "  * 实现核心功能 - 状态: 进行中" {
				foundNestedContent = true
			}

			// 确保没有未处理的模板语法
			if fullText == "{{#each tasks}}" || fullText == "  * {{taskName}} - 状态: {{status}}" {
				t.Errorf("Found unprocessed template syntax in output: %s", fullText)
			}
		}
	}

	if !foundNestedContent {
		t.Error("Expected to find nested loop content in rendered document")
	}
}

// TestDeepNestedLoops 测试深度嵌套循环（三层）
func TestDeepNestedLoops(t *testing.T) {
	engine := NewTemplateEngine()

	// 创建三层嵌套循环的模板
	templateContent := `组织架构：
{{#each departments}}
部门：{{name}}
{{#each teams}}
  团队：{{teamName}}
  {{#each members}}
    成员：{{memberName}} - {{position}}
  {{/each}}
{{/each}}
{{/each}}`

	_, err := engine.LoadTemplate("org_structure", templateContent)
	if err != nil {
		t.Fatalf("Failed to load template with deep nested loops: %v", err)
	}

	// 创建三层嵌套数据
	data := NewTemplateData()

	departments := []interface{}{
		map[string]interface{}{
			"name": "技术部",
			"teams": []interface{}{
				map[string]interface{}{
					"teamName": "前端团队",
					"members": []interface{}{
						map[string]interface{}{
							"memberName": "王五",
							"position":   "前端工程师",
						},
						map[string]interface{}{
							"memberName": "赵六",
							"position":   "UI设计师",
						},
					},
				},
				map[string]interface{}{
					"teamName": "后端团队",
					"members": []interface{}{
						map[string]interface{}{
							"memberName": "孙七",
							"position":   "后端工程师",
						},
					},
				},
			},
		},
	}
	data.SetList("departments", departments)

	// 渲染模板
	doc, err := engine.RenderToDocument("org_structure", data)
	if err != nil {
		t.Fatalf("Failed to render template with deep nested loops: %v", err)
	}

	if doc == nil {
		t.Fatal("Expected document to be created")
	}

	// 验证第三层嵌套内容是否正确渲染
	foundDeepContent := false
	for _, element := range doc.Body.Elements {
		if para, ok := element.(*Paragraph); ok {
			fullText := ""
			for _, run := range para.Runs {
				fullText += run.Text.Content
			}

			// 检查第三层嵌套内容
			if fullText == "    成员：王五 - 前端工程师" ||
				fullText == "    成员：孙七 - 后端工程师" {
				foundDeepContent = true
			}

			// 确保没有未处理的模板语法
			if fullText == "{{#each members}}" || fullText == "    成员：{{memberName}} - {{position}}" {
				t.Errorf("Found unprocessed template syntax in deep nested output: %s", fullText)
			}
		}
	}

	if !foundDeepContent {
		t.Error("Expected to find deep nested loop content in rendered document")
	}
}

// TestHeaderFooterTemplateVariables 测试页眉页脚中的模板变量识别和替换
func TestHeaderFooterTemplateVariables(t *testing.T) {
	// 创建包含页眉页脚的文档
	doc := New()

	// 添加主体内容
	doc.AddParagraph("{{title}}")
	doc.AddParagraph("文档内容")

	// 添加带有模板变量的页眉
	err := doc.AddHeader(HeaderFooterTypeDefault, "{{headerTitle}} - {{headerID}}")
	if err != nil {
		t.Fatalf("添加页眉失败: %v", err)
	}

	// 添加带有模板变量的页脚
	err = doc.AddFooter(HeaderFooterTypeDefault, "{{footerText}} - 第 {{pageNum}} 页")
	if err != nil {
		t.Fatalf("添加页脚失败: %v", err)
	}

	// 创建模板引擎并加载文档作为模板
	engine := NewTemplateEngine()
	template, err := engine.LoadTemplateFromDocument("header_footer_test", doc)
	if err != nil {
		t.Fatalf("从文档加载模板失败: %v", err)
	}

	// 验证模板变量被正确识别
	expectedVars := []string{"title", "headerTitle", "headerID", "footerText", "pageNum"}
	for _, varName := range expectedVars {
		if _, exists := template.Variables[varName]; !exists {
			t.Errorf("模板变量 '%s' 应该被识别但未找到", varName)
		}
	}

	// 测试使用TemplateRenderer分析包含页眉页脚的模板
	// 创建一个新的带页眉页脚的文档用于测试分析功能
	doc2 := New()
	doc2.AddParagraph("{{mainContent}}")
	err = doc2.AddHeader(HeaderFooterTypeDefault, "{{documentTitle}}")
	if err != nil {
		t.Fatalf("添加页眉失败: %v", err)
	}

	// 通过engine加载
	engine2 := NewTemplateEngine()
	_, err = engine2.LoadTemplateFromDocument("analyze_test", doc2)
	if err != nil {
		t.Fatalf("从文档加载模板失败: %v", err)
	}

	// 创建renderer并使用已加载的模板
	renderer := &TemplateRenderer{
		engine: engine2,
		logger: &TemplateLogger{enabled: false},
	}

	// 分析模板
	analysis, err := renderer.AnalyzeTemplate("analyze_test")
	if err != nil {
		t.Fatalf("分析模板失败: %v", err)
	}

	// 验证分析结果包含页眉中的变量
	if _, exists := analysis.Variables["documentTitle"]; !exists {
		t.Error("分析结果应该包含页眉变量 'documentTitle'")
	}
	if _, exists := analysis.Variables["mainContent"]; !exists {
		t.Error("分析结果应该包含主体变量 'mainContent'")
	}

	t.Logf("分析到的变量: %v", analysis.Variables)
}

// TestHeaderFooterVariableReplacement 测试页眉页脚中的变量替换功能
func TestHeaderFooterVariableReplacement(t *testing.T) {
	// 创建包含页眉页脚的文档
	doc := New()

	// 添加主体内容
	doc.AddParagraph("{{title}}")
	doc.AddParagraph("正文内容")

	// 添加带有模板变量的页眉
	err := doc.AddHeader(HeaderFooterTypeDefault, "报告编号: {{reportID}}")
	if err != nil {
		t.Fatalf("添加页眉失败: %v", err)
	}

	// 添加带有模板变量的页脚
	err = doc.AddFooter(HeaderFooterTypeDefault, "作者: {{author}}")
	if err != nil {
		t.Fatalf("添加页脚失败: %v", err)
	}

	// 创建模板引擎并加载文档作为模板
	engine := NewTemplateEngine()
	_, err = engine.LoadTemplateFromDocument("replacement_test", doc)
	if err != nil {
		t.Fatalf("从文档加载模板失败: %v", err)
	}

	// 准备模板数据
	data := NewTemplateData()
	data.SetVariable("title", "测试报告标题")
	data.SetVariable("reportID", "RPT-2024-001")
	data.SetVariable("author", "测试作者")

	// 渲染模板
	resultDoc, err := engine.RenderTemplateToDocument("replacement_test", data)
	if err != nil {
		t.Fatalf("渲染模板失败: %v", err)
	}

	// 验证页眉中的变量被替换
	headerReplaced := false
	footerReplaced := false

	for partName, partData := range resultDoc.parts {
		content := string(partData)

		if partName == "word/header1.xml" {
			if !strings.Contains(content, "{{reportID}}") && strings.Contains(content, "RPT-2024-001") {
				headerReplaced = true
			}
			t.Logf("页眉内容: %s", content)
		}

		if partName == "word/footer1.xml" {
			if !strings.Contains(content, "{{author}}") && strings.Contains(content, "测试作者") {
				footerReplaced = true
			}
			t.Logf("页脚内容: %s", content)
		}
	}

	if !headerReplaced {
		t.Error("页眉中的变量应该被替换")
	}

	if !footerReplaced {
		t.Error("页脚中的变量应该被替换")
	}
}

// TestTemplateFromFileWithParagraphSectionProperties 确保段落内的节属性仍能保留页眉页脚
func TestTemplateFromFileWithParagraphSectionProperties(t *testing.T) {
	doc := New()
	doc.AddParagraph("{{title}}")

	if err := doc.AddHeader(HeaderFooterTypeDefault, "报告编号: {{reportID}}"); err != nil {
		t.Fatalf("添加页眉失败: %v", err)
	}
	if err := doc.AddFooter(HeaderFooterTypeDefault, "撰写人: {{author}}"); err != nil {
		t.Fatalf("添加页脚失败: %v", err)
	}

	sectionMarker := "__SECTION_BREAK__"
	doc.AddParagraph(sectionMarker)

	tmpDir := t.TempDir()
	basePath := filepath.Join(tmpDir, "base_paragraph_section.docx")
	if err := doc.Save(basePath); err != nil {
		t.Fatalf("保存基础文档失败: %v", err)
	}

	modifiedPath := filepath.Join(tmpDir, "paragraph_section_template.docx")
	if err := moveSectPrIntoParagraph(basePath, modifiedPath, sectionMarker); err != nil {
		t.Fatalf("调整节属性位置失败: %v", err)
	}

	loadedDoc, err := Open(modifiedPath)
	if err != nil {
		t.Fatalf("打开修改后的文档失败: %v", err)
	}

	engine := NewTemplateEngine()
	_, err = engine.LoadTemplateFromDocument("paragraph_section_template", loadedDoc)
	if err != nil {
		t.Fatalf("加载模板失败: %v", err)
	}

	data := NewTemplateData()
	data.SetVariable("title", "段落节属性测试")
	data.SetVariable("reportID", "RPT-2024-009")
	data.SetVariable("author", "测试作者")

	renderedDoc, err := engine.RenderTemplateToDocument("paragraph_section_template", data)
	if err != nil {
		t.Fatalf("渲染模板失败: %v", err)
	}

	headerContent := string(renderedDoc.parts["word/header1.xml"])
	if strings.Contains(headerContent, "{{reportID}}") {
		t.Error("页眉中的变量应该被替换，即使节属性位于段落内")
	}
	if !strings.Contains(headerContent, "RPT-2024-009") {
		t.Error("页眉中缺少替换后的值")
	}

	footerContent := string(renderedDoc.parts["word/footer1.xml"])
	if strings.Contains(footerContent, "{{author}}") {
		t.Error("页脚中的变量应该被替换")
	}
	if !strings.Contains(footerContent, "测试作者") {
		t.Error("页脚中缺少替换后的值")
	}
}

func moveSectPrIntoParagraph(srcPath, dstPath, marker string) error {
	reader, err := zip.OpenReader(srcPath)
	if err != nil {
		return err
	}
	defer reader.Close()

	output, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer output.Close()

	zipWriter := zip.NewWriter(output)
	defer zipWriter.Close()

	for _, file := range reader.File {
		rc, err := file.Open()
		if err != nil {
			return err
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return err
		}

		if file.Name == "word/document.xml" {
			data, err = rewriteSectPrIntoParagraph(data, marker)
			if err != nil {
				return err
			}
		}

		writer, err := zipWriter.Create(file.Name)
		if err != nil {
			return err
		}
		if _, err := writer.Write(data); err != nil {
			return err
		}
	}

	return nil
}

func rewriteSectPrIntoParagraph(xmlData []byte, marker string) ([]byte, error) {
	content := string(xmlData)
	sectStart := strings.Index(content, "<w:sectPr")
	if sectStart == -1 {
		return nil, fmt.Errorf("未找到sectPr")
	}

	sectEndRel := strings.Index(content[sectStart:], "</w:sectPr>")
	if sectEndRel == -1 {
		return nil, fmt.Errorf("sectPr缺少结束标签")
	}
	sectEnd := sectStart + sectEndRel + len("</w:sectPr>")
	sectBlock := content[sectStart:sectEnd]

	sanitized := removeHeaderFooterReferences(sectBlock)
	content = content[:sectStart] + sanitized + content[sectEnd:]

	markerIndex := strings.Index(content, marker)
	if markerIndex == -1 {
		return nil, fmt.Errorf("未找到标记段落")
	}

	pStart := strings.LastIndex(content[:markerIndex], "<w:p")
	if pStart == -1 {
		return nil, fmt.Errorf("未找到段落起始标签")
	}
	openEnd := strings.Index(content[pStart:], ">")
	if openEnd == -1 {
		return nil, fmt.Errorf("段落标签未闭合")
	}
	insertPos := pStart + openEnd + 1

	insert := "<w:pPr>" + sectBlock + "</w:pPr>"
	modified := content[:insertPos] + insert + content[insertPos:]

	return []byte(modified), nil
}

func removeHeaderFooterReferences(block string) string {
	block = stripReferenceTag(block, "<w:headerReference")
	block = stripReferenceTag(block, "<w:footerReference")
	return block
}

func stripReferenceTag(block, tag string) string {
	for {
		start := strings.Index(block, tag)
		if start == -1 {
			break
		}
		end := strings.Index(block[start:], "/>")
		if end == -1 {
			break
		}
		block = block[:start] + block[start+end+2:]
	}
	return block
}

// TestTemplateDocumentPartsPreservation 测试模板渲染时文档部件的完整保留
func TestTemplateDocumentPartsPreservation(t *testing.T) {
	// 创建包含多种文档部件的源文档
	doc := New()

	// 添加页眉和页脚
	err := doc.AddHeader(HeaderFooterTypeDefault, "Template Header - {{headerVar}}")
	if err != nil {
		t.Fatalf("添加页眉失败: %v", err)
	}

	err = doc.AddFooter(HeaderFooterTypeDefault, "Template Footer - {{footerVar}}")
	if err != nil {
		t.Fatalf("添加页脚失败: %v", err)
	}

	// 设置页面设置
	settings := DefaultPageSettings()
	settings.Size = PageSizeA4
	settings.Orientation = OrientationPortrait
	err = doc.SetPageSettings(settings)
	if err != nil {
		t.Fatalf("设置页面设置失败: %v", err)
	}

	// 添加标题和内容
	doc.AddHeadingParagraph("{{docTitle}}", 1)
	doc.AddParagraph("Content with {{variable1}} and more text.")

	// 保存原文档
	originalPath := "test_parts_preservation_original.docx"
	err = doc.Save(originalPath)
	if err != nil {
		t.Fatalf("保存原文档失败: %v", err)
	}
	defer func() {
		if err := os.Remove(originalPath); err != nil {
			t.Logf("清理原文档失败: %v", err)
		}
	}()

	// 打开原文档作为模板
	templateDoc, err := Open(originalPath)
	if err != nil {
		t.Fatalf("打开模板文档失败: %v", err)
	}

	// 记录原文档的parts
	originalParts := make(map[string]bool)
	for partName := range templateDoc.parts {
		originalParts[partName] = true
	}

	// 创建模板引擎并加载模板
	engine := NewTemplateEngine()
	_, err = engine.LoadTemplateFromDocument("parts_test", templateDoc)
	if err != nil {
		t.Fatalf("加载模板失败: %v", err)
	}

	// 渲染模板
	data := NewTemplateData()
	data.SetVariable("headerVar", "Header Value")
	data.SetVariable("footerVar", "Footer Value")
	data.SetVariable("docTitle", "Document Title")
	data.SetVariable("variable1", "Variable 1 Value")

	renderedDoc, err := engine.RenderTemplateToDocument("parts_test", data)
	if err != nil {
		t.Fatalf("渲染模板失败: %v", err)
	}

	// 保存渲染后的文档
	renderedPath := "test_parts_preservation_rendered.docx"
	err = renderedDoc.Save(renderedPath)
	if err != nil {
		t.Fatalf("保存渲染后的文档失败: %v", err)
	}
	defer func() {
		if err := os.Remove(renderedPath); err != nil {
			t.Logf("清理渲染文档失败: %v", err)
		}
	}()

	// 检查渲染后文档的parts
	renderedParts := make(map[string]bool)
	for partName := range renderedDoc.parts {
		renderedParts[partName] = true
	}

	// 验证关键部件被保留
	criticalParts := []string{
		"word/styles.xml",
		"word/header1.xml",
		"word/footer1.xml",
	}

	for _, part := range criticalParts {
		if originalParts[part] && !renderedParts[part] {
			t.Errorf("关键部件 %s 在原文档中存在但在渲染后的文档中丢失", part)
		}
	}

	// 验证页眉页脚变量被替换
	headerContent := string(renderedDoc.parts["word/header1.xml"])
	if strings.Contains(headerContent, "{{headerVar}}") {
		t.Error("页眉中的变量应该被替换")
	}
	if !strings.Contains(headerContent, "Header Value") {
		t.Error("页眉中应该包含替换后的值")
	}

	footerContent := string(renderedDoc.parts["word/footer1.xml"])
	if strings.Contains(footerContent, "{{footerVar}}") {
		t.Error("页脚中的变量应该被替换")
	}
	if !strings.Contains(footerContent, "Footer Value") {
		t.Error("页脚中应该包含替换后的值")
	}

	t.Log("文档部件保留测试通过")
}

// TestTemplateSectionPropertiesPreservation 测试节属性在模板渲染时的保留
func TestTemplateSectionPropertiesPreservation(t *testing.T) {
	// 创建包含节属性的源文档
	doc := New()

	// 设置页面设置（这会创建SectionProperties）
	settings := DefaultPageSettings()
	settings.Size = PageSizeA4
	settings.MarginTop = 30.0
	settings.MarginBottom = 25.0
	settings.MarginLeft = 20.0
	settings.MarginRight = 20.0
	err := doc.SetPageSettings(settings)
	if err != nil {
		t.Fatalf("设置页面设置失败: %v", err)
	}

	// 添加内容
	doc.AddParagraph("Content with {{variable}}")

	// 保存原文档
	originalPath := "test_section_props_original.docx"
	err = doc.Save(originalPath)
	if err != nil {
		t.Fatalf("保存原文档失败: %v", err)
	}
	defer func() {
		if err := os.Remove(originalPath); err != nil {
			t.Logf("清理原文档失败: %v", err)
		}
	}()

	// 打开原文档作为模板
	templateDoc, err := Open(originalPath)
	if err != nil {
		t.Fatalf("打开模板文档失败: %v", err)
	}

	// 获取原文档的页面设置
	originalSettings := templateDoc.GetPageSettings()

	// 创建模板引擎并加载模板
	engine := NewTemplateEngine()
	_, err = engine.LoadTemplateFromDocument("section_test", templateDoc)
	if err != nil {
		t.Fatalf("加载模板失败: %v", err)
	}

	// 渲染模板
	data := NewTemplateData()
	data.SetVariable("variable", "Value")

	renderedDoc, err := engine.RenderTemplateToDocument("section_test", data)
	if err != nil {
		t.Fatalf("渲染模板失败: %v", err)
	}

	// 获取渲染后文档的页面设置
	renderedSettings := renderedDoc.GetPageSettings()

	// 验证页面设置被保留
	if renderedSettings.Size != originalSettings.Size {
		t.Errorf("页面大小不匹配: 期望 %v, 实际 %v", originalSettings.Size, renderedSettings.Size)
	}

	// 允许1mm的误差
	tolerance := 1.0
	if abs(renderedSettings.MarginTop-originalSettings.MarginTop) > tolerance {
		t.Errorf("上边距不匹配: 期望 %.1f, 实际 %.1f", originalSettings.MarginTop, renderedSettings.MarginTop)
	}
	if abs(renderedSettings.MarginBottom-originalSettings.MarginBottom) > tolerance {
		t.Errorf("下边距不匹配: 期望 %.1f, 实际 %.1f", originalSettings.MarginBottom, renderedSettings.MarginBottom)
	}
	if abs(renderedSettings.MarginLeft-originalSettings.MarginLeft) > tolerance {
		t.Errorf("左边距不匹配: 期望 %.1f, 实际 %.1f", originalSettings.MarginLeft, renderedSettings.MarginLeft)
	}
	if abs(renderedSettings.MarginRight-originalSettings.MarginRight) > tolerance {
		t.Errorf("右边距不匹配: 期望 %.1f, 实际 %.1f", originalSettings.MarginRight, renderedSettings.MarginRight)
	}

	t.Log("节属性保留测试通过")
}

// TestTemplateNumberingPropertiesPreservation 测试模板渲染时编号属性的保留
func TestTemplateNumberingPropertiesPreservation(t *testing.T) {
	// 创建包含编号段落的文档
	doc := New()

	// 添加带有编号的列表项
	config := &ListConfig{
		Type:        ListTypeNumber,
		IndentLevel: 0,
		StartNumber: 1,
	}
	doc.AddListItem("第一条 {{itemTitle}}", config)
	doc.AddListItem("第二条 {{itemContent}}", config)

	// 保存原文档
	originalPath := "test_numbering_preservation_original.docx"
	err := doc.Save(originalPath)
	if err != nil {
		t.Fatalf("保存原文档失败: %v", err)
	}
	defer func() {
		if err := os.Remove(originalPath); err != nil {
			t.Logf("清理原文档失败: %v", err)
		}
	}()

	// 打开原文档作为模板
	templateDoc, err := Open(originalPath)
	if err != nil {
		t.Fatalf("打开模板文档失败: %v", err)
	}

	// 验证原文档的编号属性被正确解析
	paragraphs := templateDoc.Body.GetParagraphs()
	if len(paragraphs) < 2 {
		t.Fatalf("期望至少2个段落，实际 %d 个", len(paragraphs))
	}

	// 检查第一个段落的编号属性
	if paragraphs[0].Properties == nil || paragraphs[0].Properties.NumberingProperties == nil {
		t.Error("第一个段落的编号属性应该被解析")
	}

	// 创建模板引擎并加载模板
	engine := NewTemplateEngine()
	_, err = engine.LoadTemplateFromDocument("numbering_test", templateDoc)
	if err != nil {
		t.Fatalf("加载模板失败: %v", err)
	}

	// 渲染模板
	data := NewTemplateData()
	data.SetVariable("itemTitle", "合作项目情况")
	data.SetVariable("itemContent", "合作项目背景")

	renderedDoc, err := engine.RenderTemplateToDocument("numbering_test", data)
	if err != nil {
		t.Fatalf("渲染模板失败: %v", err)
	}

	// 保存渲染后的文档
	renderedPath := "test_numbering_preservation_rendered.docx"
	err = renderedDoc.Save(renderedPath)
	if err != nil {
		t.Fatalf("保存渲染后的文档失败: %v", err)
	}
	defer func() {
		if err := os.Remove(renderedPath); err != nil {
			t.Logf("清理渲染文档失败: %v", err)
		}
	}()

	// 验证渲染后文档的编号属性被保留
	renderedParagraphs := renderedDoc.Body.GetParagraphs()
	if len(renderedParagraphs) < 2 {
		t.Fatalf("渲染后期望至少2个段落，实际 %d 个", len(renderedParagraphs))
	}

	// 检查渲染后段落的编号属性是否被保留
	for i, para := range renderedParagraphs[:2] {
		if para.Properties == nil {
			t.Errorf("段落 %d 的属性不应为空", i+1)
			continue
		}
		if para.Properties.NumberingProperties == nil {
			t.Errorf("段落 %d 的编号属性不应为空", i+1)
			continue
		}
		if para.Properties.NumberingProperties.NumID == nil {
			t.Errorf("段落 %d 的编号ID不应为空", i+1)
		}
		if para.Properties.NumberingProperties.ILevel == nil {
			t.Errorf("段落 %d 的编号级别不应为空", i+1)
		}
	}

	// 验证变量已被替换
	firstParaText := ""
	for _, run := range renderedParagraphs[0].Runs {
		firstParaText += run.Text.Content
	}
	if !strings.Contains(firstParaText, "合作项目情况") {
		t.Errorf("第一个段落应该包含替换后的变量值，实际内容: %s", firstParaText)
	}

	t.Log("编号属性保留测试通过")
}
