// Package test 测试TOC更新功能
package test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ZeroHawkeye/wordZero/pkg/document"
)

// TestTOCUpdate 测试UpdateTOC功能
func TestTOCUpdate(t *testing.T) {
	// 创建新文档
	doc := document.New()

	// 配置目录
	tocConfig := &document.TOCConfig{
		Title:       "目录", // 目录标题
		MaxLevel:    3,      // 包含到哪个标题级别
		ShowPageNum: true,   // 是否显示页码
		DotLeader:   true,   // 是否使用点状引导线
	}

	// 添加封面
	doc.AddParagraph("封面示例")

	// 生成目录
	err := doc.GenerateTOC(tocConfig)
	if err != nil {
		t.Fatalf("GenerateTOC failed: %v", err)
	}

	// 添加标题
	doc.AddHeadingParagraph("第一章", 1)
	doc.AddHeadingParagraph("1.1", 2)
	doc.AddHeadingParagraph("第二章", 1)

	// 更新目录
	err = doc.UpdateTOC()
	if err != nil {
		t.Fatalf("UpdateTOC failed: %v", err)
	}

	// 保存文档用于检查
	outputDir := filepath.Join("test", "output")
	os.MkdirAll(outputDir, 0755)
	outputPath := filepath.Join(outputDir, "toc_update_test.docx")

	err = doc.Save(outputPath)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	t.Logf("测试文档已保存到: %s", outputPath)

	// 验证标题已被收集
	headings := doc.ListHeadings()
	if len(headings) == 0 {
		t.Error("Expected headings to be collected, but got none")
	}

	expectedHeadings := []struct {
		text  string
		level int
	}{
		{"第一章", 1},
		{"1.1", 2},
		{"第二章", 1},
	}

	if len(headings) != len(expectedHeadings) {
		t.Errorf("Expected %d headings, got %d", len(expectedHeadings), len(headings))
	}

	for i, expected := range expectedHeadings {
		if i < len(headings) {
			if headings[i].Text != expected.text {
				t.Errorf("Heading %d: expected text '%s', got '%s'", i, expected.text, headings[i].Text)
			}
			if headings[i].Level != expected.level {
				t.Errorf("Heading %d: expected level %d, got %d", i, expected.level, headings[i].Level)
			}
		}
	}
}
