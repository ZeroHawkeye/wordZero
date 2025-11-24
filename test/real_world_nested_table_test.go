// Package test 真实场景嵌套表格模板测试
package test

import (
	"os"
	"testing"

	"github.com/ZeroHawkeye/wordZero/pkg/document"
)

// TestRealWorldNestedTableScenario 测试真实场景：简历中的家庭成员表格
// 这个测试模拟了issue中描述的场景
func TestRealWorldNestedTableScenario(t *testing.T) {
	// 确保输出目录存在
	outputDir := "output"
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		err = os.Mkdir(outputDir, 0755)
		if err != nil {
			t.Fatalf("创建输出目录失败: %v", err)
		}
	}

	// 创建模拟真实简历的文档
	doc := document.New()
	doc.AddParagraph("个人简历模板")
	doc.AddParagraph("")

	// 创建主表格（简历表格）
	mainTable := doc.CreateTable(&document.TableConfig{
		Rows:  2,
		Cols:  2,
		Width: 9000,
	})
	doc.Body.Elements = append(doc.Body.Elements, mainTable)

	// 第一行：简历标题
	mainTable.Rows[0].Cells[0].Paragraphs[0].Runs[0].Text.Content = "简历"
	mainTable.Rows[0].Cells[1].Paragraphs[0].Runs[0].Text.Content = "{{resume}}"

	// 第二行：家庭成员信息
	mainTable.Rows[1].Cells[0].Paragraphs[0].Runs[0].Text.Content = "家庭主要成员及重要社会关系"

	// 创建嵌套的家庭成员表格（这是关键部分）
	familyTable := &document.Table{
		Properties: &document.TableProperties{
			TableW: &document.TableWidth{
				W:    "4000",
				Type: "dxa",
			},
			TableBorders: &document.TableBorders{
				Top: &document.TableBorder{
					Val:   "single",
					Sz:    "4",
					Color: "000000",
				},
				Left: &document.TableBorder{
					Val:   "single",
					Sz:    "4",
					Color: "000000",
				},
				Bottom: &document.TableBorder{
					Val:   "single",
					Sz:    "4",
					Color: "000000",
				},
				Right: &document.TableBorder{
					Val:   "single",
					Sz:    "4",
					Color: "000000",
				},
			},
		},
		Rows: []document.TableRow{
			// 标题行
			{
				Cells: []document.TableCell{
					{Paragraphs: []document.Paragraph{{Runs: []document.Run{{Text: document.Text{Content: "姓名"}}}}}},
					{Paragraphs: []document.Paragraph{{Runs: []document.Run{{Text: document.Text{Content: "年龄"}}}}}},
					{Paragraphs: []document.Paragraph{{Runs: []document.Run{{Text: document.Text{Content: "性别"}}}}}},
					{Paragraphs: []document.Paragraph{{Runs: []document.Run{{Text: document.Text{Content: "关系"}}}}}},
				},
			},
			// 数据行（使用模板语法）
			{
				Cells: []document.TableCell{
					{Paragraphs: []document.Paragraph{{Runs: []document.Run{{Text: document.Text{Content: "{{#each family_members}}{{name}}"}}}}}},
					{Paragraphs: []document.Paragraph{{Runs: []document.Run{{Text: document.Text{Content: "{{age}}"}}}}}},
					{Paragraphs: []document.Paragraph{{Runs: []document.Run{{Text: document.Text{Content: "{{gender}}"}}}}}},
					{Paragraphs: []document.Paragraph{{Runs: []document.Run{{Text: document.Text{Content: "{{relationship}}{{/each}}"}}}}}},
				},
			},
		},
	}

	// 将家庭成员表格嵌套到主表格的单元格中
	mainTable.Rows[1].Cells[1].Tables = []document.Table{*familyTable}
	
	// 保存模板文档
	templatePath := "output/resume_template.docx"
	err := doc.Save(templatePath)
	if err != nil {
		t.Fatalf("保存模板文档失败: %v", err)
	}
	t.Logf("模板文档已保存: %s", templatePath)

	// 创建模板引擎并渲染
	engine := document.NewTemplateEngine()
	_, err = engine.LoadTemplateFromDocument("resume", doc)
	if err != nil {
		t.Fatalf("加载模板失败: %v", err)
	}

	// 准备数据
	data := document.NewTemplateData()
	data.SetVariable("resume", "张小明的个人简历")
	data.SetList("family_members", []interface{}{
		map[string]interface{}{
			"name":         "张大明",
			"age":          "50",
			"gender":       "男",
			"relationship": "父亲",
		},
		map[string]interface{}{
			"name":         "李红",
			"age":          "48",
			"gender":       "女",
			"relationship": "母亲",
		},
	})

	// 渲染文档
	renderedDoc, err := engine.RenderTemplateToDocument("resume", data)
	if err != nil {
		t.Fatalf("渲染模板失败: %v", err)
	}

	// 保存渲染后的文档
	outputPath := "output/resume_rendered.docx"
	err = renderedDoc.Save(outputPath)
	if err != nil {
		t.Fatalf("保存渲染文档失败: %v", err)
	}
	t.Logf("渲染文档已保存: %s", outputPath)

	// 验证嵌套表格存在
	tables := renderedDoc.Body.GetTables()
	if len(tables) < 1 {
		t.Fatalf("应该有至少1个主表格，实际有 %d 个", len(tables))
	}

	mainTableResult := tables[0]
	if len(mainTableResult.Rows) < 2 {
		t.Fatalf("主表格应该有2行，实际有 %d 行", len(mainTableResult.Rows))
	}

	// 关键验证：嵌套表格必须存在
	nestedTables := mainTableResult.Rows[1].Cells[1].Tables
	if len(nestedTables) == 0 {
		t.Fatal("❌ BUG REPRODUCED: 嵌套表格在渲染后消失了！这就是issue中描述的问题。")
	}

	t.Log("✅ BUG FIXED: 嵌套表格在渲染后成功保留！")

	// 验证嵌套表格的数据
	nestedTable := nestedTables[0]
	// 应该有3行：1个标题行 + 2个数据行
	expectedRows := 3
	if len(nestedTable.Rows) != expectedRows {
		t.Errorf("嵌套表格应该有 %d 行，实际有 %d 行", expectedRows, len(nestedTable.Rows))
	}

	// 验证第一行数据
	if len(nestedTable.Rows) >= 2 {
		firstRow := nestedTable.Rows[1]
		if len(firstRow.Cells) >= 4 {
			name := firstRow.Cells[0].Paragraphs[0].Runs[0].Text.Content
			age := firstRow.Cells[1].Paragraphs[0].Runs[0].Text.Content
			gender := firstRow.Cells[2].Paragraphs[0].Runs[0].Text.Content
			relation := firstRow.Cells[3].Paragraphs[0].Runs[0].Text.Content

			t.Logf("家庭成员数据 - 姓名: %s, 年龄: %s, 性别: %s, 关系: %s",
				name, age, gender, relation)

			if name != "张大明" || age != "50" || gender != "男" || relation != "父亲" {
				t.Errorf("数据渲染不正确")
			}
		}
	}

	t.Log("✅ 真实场景测试通过：简历中的家庭成员嵌套表格正确渲染")
}
