// Package document 模板图片保留测试
package document

import (
	"os"
	"testing"
)

// TestTemplateImagePreservation 测试模板替换后图片和格式是否保留
func TestTemplateImagePreservation(t *testing.T) {
	// 创建一个最小的有效PNG图片数据用于测试
	testImageData := []byte{
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

	// 步骤1：创建一个包含图片和格式的模板文档
	t.Log("步骤1: 创建包含图片和格式的模板文档")
	templateDoc := New()

	// 添加带模板变量的标题
	templateDoc.AddHeadingParagraph("文档标题: {{title}}", 1)

	// 添加带格式的段落
	formattedPara := templateDoc.AddFormattedParagraph("作者: {{author}}", &TextFormat{
		Bold:      true,
		FontSize:  14,
		FontColor: "FF0000",
		FontName:  "Arial",
	})
	formattedPara.SetAlignment(AlignCenter)

	// 添加图片
	imageConfig := &ImageConfig{
		Position:  ImagePositionInline,
		Alignment: AlignCenter,
		Size:      &ImageSize{Width: 50, KeepAspectRatio: true},
	}

	imageInfo, err := templateDoc.AddImageFromData(testImageData, "test_image.png", ImageFormatPNG, 1, 1, imageConfig)
	if err != nil {
		t.Fatalf("添加图片失败: %v", err)
	}
	t.Logf("已添加图片，ID: %s", imageInfo.ID)

	// 添加另一个带模板变量的段落
	templateDoc.AddParagraph("日期: {{date}}")

	// 保存模板文档
	templatePath := "test_template_with_image.docx"
	err = templateDoc.Save(templatePath)
	if err != nil {
		t.Fatalf("保存模板文档失败: %v", err)
	}
	defer os.Remove(templatePath)

	// 统计初始文档中的图片元素
	initialImageCount := 0
	for _, elem := range templateDoc.Body.Elements {
		if para, ok := elem.(*Paragraph); ok {
			for _, run := range para.Runs {
				if run.Drawing != nil {
					initialImageCount++
				}
			}
		}
	}
	t.Logf("初始文档包含 %d 个图片元素", initialImageCount)

	// 步骤2：重新打开模板文档
	t.Log("步骤2: 重新打开模板文档")
	openedDoc, err := Open(templatePath)
	if err != nil {
		t.Fatalf("打开模板文档失败: %v", err)
	}

	// 检查打开文档中的媒体文件
	openedHasImageMedia := false
	for partName := range openedDoc.parts {
		if len(partName) > 11 && partName[:11] == "word/media/" {
			openedHasImageMedia = true
			t.Logf("打开的文档包含媒体文件: %s", partName)
		}
	}
	if !openedHasImageMedia {
		t.Logf("警告: 打开的文档中未找到媒体文件")
	}

	// 步骤3：使用模板引擎进行变量替换
	t.Log("步骤3: 使用模板引擎进行变量替换")
	engine := NewTemplateEngine()

	_, err = engine.LoadTemplateFromDocument("test_template", openedDoc)
	if err != nil {
		t.Fatalf("加载模板失败: %v", err)
	}

	// 准备数据
	data := NewTemplateData()
	data.SetVariable("title", "测试报告")
	data.SetVariable("author", "张三")
	data.SetVariable("date", "2025年1月1日")

	// 渲染模板
	resultDoc, err := engine.RenderTemplateToDocument("test_template", data)
	if err != nil {
		t.Fatalf("渲染模板失败: %v", err)
	}

	// 步骤4：检查结果文档
	t.Log("步骤4: 检查结果文档")

	// 检查结果文档中的图片元素
	resultImageCount := 0
	for _, elem := range resultDoc.Body.Elements {
		if para, ok := elem.(*Paragraph); ok {
			for _, run := range para.Runs {
				if run.Drawing != nil {
					resultImageCount++
					t.Logf("结果文档中找到图片元素")
				}
			}
		}
	}

	// 检查结果文档中的媒体文件
	resultHasImageMedia := false
	for partName := range resultDoc.parts {
		if len(partName) > 11 && partName[:11] == "word/media/" {
			resultHasImageMedia = true
			t.Logf("结果文档包含媒体文件: %s", partName)
		}
	}

	// 验证图片元素是否保留
	if resultImageCount == 0 {
		t.Errorf("模板渲染后图片元素丢失。初始: %d, 结果: %d", initialImageCount, resultImageCount)
	} else {
		t.Logf("✓ 图片元素已保留: %d", resultImageCount)
	}

	// 验证媒体文件是否保留
	if !resultHasImageMedia {
		t.Errorf("模板渲染后媒体文件丢失")
	} else {
		t.Logf("✓ 媒体文件已保留")
	}

	// 保存结果文档
	resultPath := "test_template_rendered.docx"
	err = resultDoc.Save(resultPath)
	if err != nil {
		t.Fatalf("保存结果文档失败: %v", err)
	}
	defer os.Remove(resultPath)

	// 步骤5：验证变量替换是否成功
	t.Log("步骤5: 验证变量替换")
	foundTitle := false
	foundAuthor := false
	for _, elem := range resultDoc.Body.Elements {
		if para, ok := elem.(*Paragraph); ok {
			for _, run := range para.Runs {
				if run.Text.Content != "" {
					if containsSubstring(run.Text.Content, "测试报告") {
						foundTitle = true
					}
					if containsSubstring(run.Text.Content, "张三") {
						foundAuthor = true
					}
				}
			}
		}
	}

	if !foundTitle {
		t.Logf("警告: 标题变量未被替换")
	}
	if !foundAuthor {
		t.Logf("警告: 作者变量未被替换")
	}

	t.Log("TestTemplateImagePreservation 测试完成")
}

// containsSubstring 检查字符串中是否包含子字符串
func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && containsSubstringHelper(s, substr)))
}

func containsSubstringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
