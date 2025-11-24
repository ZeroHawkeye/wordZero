package document

import (
	"os"
	"testing"

	"github.com/ZeroHawkeye/wordZero/pkg/style"
)

// assertParagraphContent 验证段落的文本内容（带边界检查）
// assertParagraphContent verifies the text content of a paragraph with bounds checking
func assertParagraphContent(t *testing.T, paragraphs []*Paragraph, index int, expectedContent string) {
	t.Helper()
	if index >= len(paragraphs) {
		t.Errorf("段落索引 %d 超出范围，总共只有 %d 个段落", index, len(paragraphs))
		return
	}

	para := paragraphs[index]
	if len(para.Runs) == 0 {
		t.Errorf("索引 %d 的段落没有任何运行（Runs）", index)
		return
	}

	actualContent := para.Runs[0].Text.Content
	if actualContent != expectedContent {
		t.Errorf("索引 %d 的段落内容应该是'%s'，实际是'%s'", index, expectedContent, actualContent)
	}
}

// TestNewDocument 测试新文档创建
func TestNewDocument(t *testing.T) {
	doc := New()

	// 验证基本结构
	if doc == nil {
		t.Fatal("Failed to create new document")
	}

	if doc.Body == nil {
		t.Fatal("Document body is nil")
	}

	if doc.styleManager == nil {
		t.Fatal("Style manager is nil")
	}

	// 验证初始状态
	if len(doc.Body.GetParagraphs()) != 0 {
		t.Errorf("Expected 0 paragraphs, got %d", len(doc.Body.GetParagraphs()))
	}

	// 验证样式管理器初始化
	styles := doc.styleManager.GetAllStyles()
	if len(styles) == 0 {
		t.Error("Style manager should have predefined styles")
	}
}

// TestAddParagraph 测试添加普通段落
func TestAddParagraph(t *testing.T) {
	doc := New()
	text := "测试段落内容"

	para := doc.AddParagraph(text)

	// 验证段落添加
	if len(doc.Body.GetParagraphs()) != 1 {
		t.Errorf("Expected 1 paragraph, got %d", len(doc.Body.GetParagraphs()))
	}

	// 验证段落内容
	if len(para.Runs) != 1 {
		t.Errorf("Expected 1 run, got %d", len(para.Runs))
	}

	if para.Runs[0].Text.Content != text {
		t.Errorf("Expected %s, got %s", text, para.Runs[0].Text.Content)
	}

	// 验证返回的指针是否正确
	paragraphs := doc.Body.GetParagraphs()
	if paragraphs[0] != para {
		t.Error("Returned paragraph pointer is incorrect")
	}
}

// TestAddHeadingParagraph 测试添加标题段落
func TestAddHeadingParagraph(t *testing.T) {
	doc := New()

	testCases := []struct {
		text    string
		level   int
		styleID string
	}{
		{"第一级标题", 1, "Heading1"},
		{"第二级标题", 2, "Heading2"},
		{"第三级标题", 3, "Heading3"},
		{"第九级标题", 9, "Heading9"},
	}

	for _, tc := range testCases {
		para := doc.AddHeadingParagraph(tc.text, tc.level)

		// 验证段落样式设置
		if para.Properties == nil {
			t.Errorf("Heading paragraph should have properties")
			continue
		}

		if para.Properties.ParagraphStyle == nil {
			t.Errorf("Heading paragraph should have style reference")
			continue
		}

		if para.Properties.ParagraphStyle.Val != tc.styleID {
			t.Errorf("Expected style %s, got %s", tc.styleID, para.Properties.ParagraphStyle.Val)
		}

		// 验证内容
		if len(para.Runs) != 1 {
			t.Errorf("Expected 1 run, got %d", len(para.Runs))
			continue
		}

		if para.Runs[0].Text.Content != tc.text {
			t.Errorf("Expected %s, got %s", tc.text, para.Runs[0].Text.Content)
		}
	}

	// 测试超出范围的级别
	para := doc.AddHeadingParagraph("超出范围", 10)
	if para.Properties.ParagraphStyle.Val != "Heading1" {
		t.Error("Out of range level should default to Heading1")
	}

	para = doc.AddHeadingParagraph("负数级别", -1)
	if para.Properties.ParagraphStyle.Val != "Heading1" {
		t.Error("Negative level should default to Heading1")
	}
}

// TestAddFormattedParagraph 测试添加格式化段落
func TestAddFormattedParagraph(t *testing.T) {
	doc := New()
	text := "格式化文本"

	format := &TextFormat{
		Bold:       true,
		Italic:     true,
		FontSize:   14,
		FontColor:  "FF0000",
		FontFamily: "宋体",
	}

	para := doc.AddFormattedParagraph(text, format)

	// 验证段落添加
	if len(doc.Body.GetParagraphs()) != 1 {
		t.Error("Failed to add formatted paragraph")
	}

	// 验证格式设置
	run := para.Runs[0]
	if run.Properties == nil {
		t.Fatal("Run properties should not be nil")
	}

	if run.Properties.Bold == nil {
		t.Error("Bold property should be set")
	}

	if run.Properties.Italic == nil {
		t.Error("Italic property should be set")
	}

	if run.Properties.FontSize == nil || run.Properties.FontSize.Val != "28" {
		t.Errorf("Expected font size 28, got %v", run.Properties.FontSize)
	}

	if run.Properties.Color == nil || run.Properties.Color.Val != "FF0000" {
		t.Errorf("Expected color FF0000, got %v", run.Properties.Color)
	}

	if run.Properties.FontFamily == nil || run.Properties.FontFamily.ASCII != "宋体" {
		t.Errorf("Expected font family 宋体, got %v", run.Properties.FontFamily)
	}
}

// TestParagraphSetAlignment 测试段落对齐设置
func TestParagraphSetAlignment(t *testing.T) {
	doc := New()
	para := doc.AddParagraph("测试对齐")

	testCases := []AlignmentType{
		AlignLeft,
		AlignCenter,
		AlignRight,
		AlignJustify,
	}

	for _, alignment := range testCases {
		para.SetAlignment(alignment)

		if para.Properties == nil {
			t.Fatal("Properties should not be nil after setting alignment")
		}

		if para.Properties.Justification == nil {
			t.Fatal("Justification should not be nil")
		}

		if para.Properties.Justification.Val != string(alignment) {
			t.Errorf("Expected alignment %s, got %s", alignment, para.Properties.Justification.Val)
		}
	}
}

// TestParagraphSetSpacing 测试段落间距设置
func TestParagraphSetSpacing(t *testing.T) {
	doc := New()
	para := doc.AddParagraph("测试间距")

	config := &SpacingConfig{
		LineSpacing:     1.5,
		BeforePara:      12,
		AfterPara:       6,
		FirstLineIndent: 24,
	}

	para.SetSpacing(config)

	// 验证属性设置
	if para.Properties == nil {
		t.Fatal("Properties should not be nil")
	}

	if para.Properties.Spacing == nil {
		t.Fatal("Spacing should not be nil")
	}

	// 验证间距值（转换为TWIPs）
	spacing := para.Properties.Spacing
	if spacing.Before != "240" { // 12 * 20
		t.Errorf("Expected before spacing 240, got %s", spacing.Before)
	}

	if spacing.After != "120" { // 6 * 20
		t.Errorf("Expected after spacing 120, got %s", spacing.After)
	}

	if spacing.Line != "360" { // 1.5 * 240
		t.Errorf("Expected line spacing 360, got %s", spacing.Line)
	}

	// 验证首行缩进
	if para.Properties.Indentation == nil {
		t.Fatal("Indentation should not be nil")
	}

	if para.Properties.Indentation.FirstLine != "480" { // 24 * 20
		t.Errorf("Expected first line indent 480, got %s", para.Properties.Indentation.FirstLine)
	}
}

// TestParagraphAddFormattedText 测试段落添加格式化文本
func TestParagraphAddFormattedText(t *testing.T) {
	doc := New()
	para := doc.AddParagraph("初始文本")

	// 添加格式化文本
	format := &TextFormat{
		Bold:      true,
		FontColor: "0000FF",
	}

	para.AddFormattedText("格式化文本", format)

	// 验证运行数量
	if len(para.Runs) != 2 {
		t.Errorf("Expected 2 runs, got %d", len(para.Runs))
	}

	// 验证第二个运行的格式
	run := para.Runs[1]
	if run.Properties == nil {
		t.Fatal("Second run should have properties")
	}

	if run.Properties.Bold == nil {
		t.Error("Second run should be bold")
	}

	if run.Properties.Color == nil || run.Properties.Color.Val != "0000FF" {
		t.Error("Second run should be blue")
	}

	if run.Text.Content != "格式化文本" {
		t.Errorf("Expected '格式化文本', got '%s'", run.Text.Content)
	}
}

// TestParagraphSetStyle 测试段落样式设置
func TestParagraphSetStyle(t *testing.T) {
	doc := New()
	para := doc.AddParagraph("测试样式")

	para.SetStyle("Heading1")

	if para.Properties == nil {
		t.Fatal("Properties should not be nil")
	}

	if para.Properties.ParagraphStyle == nil {
		t.Fatal("ParagraphStyle should not be nil")
	}

	if para.Properties.ParagraphStyle.Val != "Heading1" {
		t.Errorf("Expected style Heading1, got %s", para.Properties.ParagraphStyle.Val)
	}
}

// TestParagraphSetIndentation 测试段落缩进设置
func TestParagraphSetIndentation(t *testing.T) {
	doc := New()
	para := doc.AddParagraph("测试缩进")

	// 测试首行缩进
	para.SetIndentation(0.5, 0, 0)

	if para.Properties == nil {
		t.Fatal("Properties should not be nil")
	}

	if para.Properties.Indentation == nil {
		t.Fatal("Indentation should not be nil")
	}

	// 0.5厘米 = 283.5 TWIPs，四舍五入为284
	expectedFirstLine := "283"
	if para.Properties.Indentation.FirstLine != expectedFirstLine {
		t.Errorf("Expected FirstLine %s, got %s", expectedFirstLine, para.Properties.Indentation.FirstLine)
	}

	// 测试左右缩进
	para.SetIndentation(-0.5, 1.0, 0.5)

	expectedFirstLine = "-283" // 悬挂缩进
	expectedLeft := "567"      // 1厘米
	expectedRight := "283"     // 0.5厘米

	if para.Properties.Indentation.FirstLine != expectedFirstLine {
		t.Errorf("Expected FirstLine %s, got %s", expectedFirstLine, para.Properties.Indentation.FirstLine)
	}
	if para.Properties.Indentation.Left != expectedLeft {
		t.Errorf("Expected Left %s, got %s", expectedLeft, para.Properties.Indentation.Left)
	}
	if para.Properties.Indentation.Right != expectedRight {
		t.Errorf("Expected Right %s, got %s", expectedRight, para.Properties.Indentation.Right)
	}
}

// TestParagraphSetKeepWithNext 测试段落与下一段保持在一起
func TestParagraphSetKeepWithNext(t *testing.T) {
	doc := New()
	para := doc.AddParagraph("测试保持与下一段")

	// 测试启用
	para.SetKeepWithNext(true)

	if para.Properties == nil {
		t.Fatal("Properties should not be nil")
	}

	if para.Properties.KeepNext == nil {
		t.Fatal("KeepNext should not be nil")
	}

	if para.Properties.KeepNext.Val != "1" {
		t.Errorf("Expected KeepNext Val to be '1', got '%s'", para.Properties.KeepNext.Val)
	}

	// 测试禁用
	para.SetKeepWithNext(false)

	if para.Properties.KeepNext != nil {
		t.Error("KeepNext should be nil when disabled")
	}
}

// TestParagraphSetKeepLines 测试段落行保持在一起
func TestParagraphSetKeepLines(t *testing.T) {
	doc := New()
	para := doc.AddParagraph("测试行保持")

	// 测试启用
	para.SetKeepLines(true)

	if para.Properties == nil {
		t.Fatal("Properties should not be nil")
	}

	if para.Properties.KeepLines == nil {
		t.Fatal("KeepLines should not be nil")
	}

	if para.Properties.KeepLines.Val != "1" {
		t.Errorf("Expected KeepLines Val to be '1', got '%s'", para.Properties.KeepLines.Val)
	}

	// 测试禁用
	para.SetKeepLines(false)

	if para.Properties.KeepLines != nil {
		t.Error("KeepLines should be nil when disabled")
	}
}

// TestParagraphSetPageBreakBefore 测试段前分页
func TestParagraphSetPageBreakBefore(t *testing.T) {
	doc := New()
	para := doc.AddParagraph("测试段前分页")

	// 测试启用
	para.SetPageBreakBefore(true)

	if para.Properties == nil {
		t.Fatal("Properties should not be nil")
	}

	if para.Properties.PageBreakBefore == nil {
		t.Fatal("PageBreakBefore should not be nil")
	}

	if para.Properties.PageBreakBefore.Val != "1" {
		t.Errorf("Expected PageBreakBefore Val to be '1', got '%s'", para.Properties.PageBreakBefore.Val)
	}

	// 测试禁用
	para.SetPageBreakBefore(false)

	if para.Properties.PageBreakBefore != nil {
		t.Error("PageBreakBefore should be nil when disabled")
	}
}

// TestParagraphSetWidowControl 测试孤行控制
func TestParagraphSetWidowControl(t *testing.T) {
	doc := New()
	para := doc.AddParagraph("测试孤行控制")

	// 测试启用
	para.SetWidowControl(true)

	if para.Properties == nil {
		t.Fatal("Properties should not be nil")
	}

	if para.Properties.WidowControl == nil {
		t.Fatal("WidowControl should not be nil")
	}

	if para.Properties.WidowControl.Val != "1" {
		t.Errorf("Expected WidowControl Val to be '1', got '%s'", para.Properties.WidowControl.Val)
	}

	// 测试禁用
	para.SetWidowControl(false)

	if para.Properties.WidowControl == nil {
		t.Fatal("WidowControl should not be nil when set to false")
	}

	if para.Properties.WidowControl.Val != "0" {
		t.Errorf("Expected WidowControl Val to be '0' when disabled, got '%s'", para.Properties.WidowControl.Val)
	}
}

// TestParagraphSetOutlineLevel 测试大纲级别设置
func TestParagraphSetOutlineLevel(t *testing.T) {
	doc := New()
	para := doc.AddParagraph("测试大纲级别")

	// 测试有效级别
	para.SetOutlineLevel(0)

	if para.Properties == nil {
		t.Fatal("Properties should not be nil")
	}

	if para.Properties.OutlineLevel == nil {
		t.Fatal("OutlineLevel should not be nil")
	}

	if para.Properties.OutlineLevel.Val != "0" {
		t.Errorf("Expected OutlineLevel Val to be '0', got '%s'", para.Properties.OutlineLevel.Val)
	}

	// 测试其他级别
	para.SetOutlineLevel(3)
	if para.Properties.OutlineLevel.Val != "3" {
		t.Errorf("Expected OutlineLevel Val to be '3', got '%s'", para.Properties.OutlineLevel.Val)
	}

	// 测试边界值
	para.SetOutlineLevel(8)
	if para.Properties.OutlineLevel.Val != "8" {
		t.Errorf("Expected OutlineLevel Val to be '8', got '%s'", para.Properties.OutlineLevel.Val)
	}

	// 测试超出范围的值（应该被限制）
	para.SetOutlineLevel(10)
	if para.Properties.OutlineLevel.Val != "8" {
		t.Errorf("Expected OutlineLevel to be capped at '8', got '%s'", para.Properties.OutlineLevel.Val)
	}

	para.SetOutlineLevel(-1)
	if para.Properties.OutlineLevel.Val != "0" {
		t.Errorf("Expected OutlineLevel to be floored at '0', got '%s'", para.Properties.OutlineLevel.Val)
	}
}

// TestParagraphSetParagraphFormat 测试综合段落格式设置
func TestParagraphSetParagraphFormat(t *testing.T) {
	doc := New()
	para := doc.AddParagraph("测试综合格式设置")

	// 测试完整配置
	config := &ParagraphFormatConfig{
		Alignment:       AlignCenter,
		Style:           "Heading1",
		LineSpacing:     1.5,
		BeforePara:      24,
		AfterPara:       12,
		FirstLineCm:     0.5,
		LeftCm:          1.0,
		RightCm:         0.5,
		KeepWithNext:    true,
		KeepLines:       true,
		PageBreakBefore: true,
		WidowControl:    true,
		OutlineLevel:    0,
	}

	para.SetParagraphFormat(config)

	// 验证所有属性
	if para.Properties == nil {
		t.Fatal("Properties should not be nil")
	}

	// 验证对齐
	if para.Properties.Justification == nil || para.Properties.Justification.Val != string(AlignCenter) {
		t.Error("Alignment not set correctly")
	}

	// 验证样式
	if para.Properties.ParagraphStyle == nil || para.Properties.ParagraphStyle.Val != "Heading1" {
		t.Error("Style not set correctly")
	}

	// 验证间距
	if para.Properties.Spacing == nil {
		t.Fatal("Spacing should not be nil")
	}

	// 验证缩进
	if para.Properties.Indentation == nil {
		t.Fatal("Indentation should not be nil")
	}

	// 验证分页控制
	if para.Properties.KeepNext == nil || para.Properties.KeepNext.Val != "1" {
		t.Error("KeepNext not set correctly")
	}

	if para.Properties.KeepLines == nil || para.Properties.KeepLines.Val != "1" {
		t.Error("KeepLines not set correctly")
	}

	if para.Properties.PageBreakBefore == nil || para.Properties.PageBreakBefore.Val != "1" {
		t.Error("PageBreakBefore not set correctly")
	}

	if para.Properties.WidowControl == nil || para.Properties.WidowControl.Val != "1" {
		t.Error("WidowControl not set correctly")
	}

	// 验证大纲级别
	if para.Properties.OutlineLevel == nil || para.Properties.OutlineLevel.Val != "0" {
		t.Error("OutlineLevel not set correctly")
	}
}

// TestParagraphSetParagraphFormatNil 测试nil配置
func TestParagraphSetParagraphFormatNil(t *testing.T) {
	doc := New()
	para := doc.AddParagraph("测试nil配置")

	// nil配置不应该导致panic
	para.SetParagraphFormat(nil)

	// 段落应该保持默认状态
	if para.Properties != nil && para.Properties.Justification != nil {
		t.Error("Properties should remain unchanged with nil config")
	}
}

// TestParagraphSetParagraphFormatPartial 测试部分配置
func TestParagraphSetParagraphFormatPartial(t *testing.T) {
	doc := New()
	para := doc.AddParagraph("测试部分配置")

	// 只设置部分属性
	config := &ParagraphFormatConfig{
		Alignment:    AlignRight,
		KeepWithNext: true,
		LineSpacing:  2.0,
	}

	para.SetParagraphFormat(config)

	// 验证设置的属性
	if para.Properties == nil {
		t.Fatal("Properties should not be nil")
	}

	if para.Properties.Justification == nil || para.Properties.Justification.Val != string(AlignRight) {
		t.Error("Alignment not set correctly")
	}

	if para.Properties.KeepNext == nil || para.Properties.KeepNext.Val != "1" {
		t.Error("KeepNext not set correctly")
	}

	if para.Properties.Spacing == nil {
		t.Error("Spacing should be set")
	}

	// 验证未设置的属性保持默认
	if para.Properties.PageBreakBefore != nil {
		t.Error("PageBreakBefore should remain nil")
	}
}

// TestDocumentSave 测试文档保存
func TestDocumentSave(t *testing.T) {
	doc := New()
	doc.AddParagraph("测试保存功能")

	filename := "test_save.docx"
	defer os.Remove(filename) // 清理测试文件

	err := doc.Save(filename)
	if err != nil {
		t.Fatalf("Failed to save document: %v", err)
	}

	// 验证文件是否存在
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Error("Saved file does not exist")
	}

	// 验证文件大小
	stat, err := os.Stat(filename)
	if err != nil {
		t.Fatalf("Failed to get file stats: %v", err)
	}

	if stat.Size() == 0 {
		t.Error("Saved file is empty")
	}
}

// TestDocumentGetStyleManager 测试获取样式管理器
func TestDocumentGetStyleManager(t *testing.T) {
	doc := New()

	styleManager := doc.GetStyleManager()
	if styleManager == nil {
		t.Fatal("Style manager should not be nil")
	}

	// 验证样式管理器功能
	if !styleManager.StyleExists("Normal") {
		t.Error("Normal style should exist")
	}

	if !styleManager.StyleExists("Heading1") {
		t.Error("Heading1 style should exist")
	}
}

// TestComplexDocument 测试复杂文档创建
func TestComplexDocument(t *testing.T) {
	doc := New()

	// 添加标题
	title := doc.AddFormattedParagraph("文档标题", &TextFormat{
		Bold:     true,
		FontSize: 18,
	})
	title.SetAlignment(AlignCenter)

	// 添加各级标题
	doc.AddHeadingParagraph("第一章", 1)
	doc.AddHeadingParagraph("1.1 概述", 2)
	doc.AddHeadingParagraph("1.1.1 背景", 3)

	// 添加带间距的段落
	para := doc.AddParagraph("这是一个带有特殊间距的段落")
	para.SetSpacing(&SpacingConfig{
		LineSpacing: 1.5,
		BeforePara:  12,
		AfterPara:   6,
	})

	// 添加混合格式段落
	mixed := doc.AddParagraph("这段文字包含")
	mixed.AddFormattedText("粗体", &TextFormat{Bold: true})
	mixed.AddFormattedText("和", nil)
	mixed.AddFormattedText("斜体", &TextFormat{Italic: true})
	mixed.AddFormattedText("文本。", nil)

	// 验证文档结构
	if len(doc.Body.GetParagraphs()) != 6 {
		t.Errorf("Expected 6 paragraphs, got %d", len(doc.Body.GetParagraphs()))
	}

	// 保存并验证
	filename := "test_complex.docx"
	defer os.Remove(filename)

	err := doc.Save(filename)
	if err != nil {
		t.Fatalf("Failed to save complex document: %v", err)
	}
}

// TestDocumentOpen 测试打开文档（需要先创建一个测试文档）
func TestDocumentOpen(t *testing.T) {
	// 先创建一个测试文档
	originalDoc := New()
	originalDoc.AddParagraph("第一段")
	originalDoc.AddParagraph("第二段")
	originalDoc.AddHeadingParagraph("标题", 1)

	filename := "test_open.docx"
	defer os.Remove(filename)

	err := originalDoc.Save(filename)
	if err != nil {
		t.Fatalf("Failed to save test document: %v", err)
	}

	// 打开文档
	loadedDoc, err := Open(filename)
	if err != nil {
		t.Fatalf("Failed to open document: %v", err)
	}

	// 验证文档内容
	if len(loadedDoc.Body.GetParagraphs()) != 3 {
		t.Errorf("Expected 3 paragraphs, got %d", len(loadedDoc.Body.GetParagraphs()))
	}

	// 验证第一段内容
	if len(loadedDoc.Body.GetParagraphs()[0].Runs) > 0 {
		content := loadedDoc.Body.GetParagraphs()[0].Runs[0].Text.Content
		if content != "第一段" {
			t.Errorf("Expected '第一段', got '%s'", content)
		}
	}
}

// TestErrorHandling 测试错误处理
func TestErrorHandling(t *testing.T) {
	// 测试打开不存在的文件
	_, err := Open("nonexistent.docx")
	if err == nil {
		t.Error("Should return error when opening non-existent file")
	}

	// 测试保存到只读目录（如果创建失败则跳过这个测试）
	doc := New()
	doc.AddParagraph("测试")

	// 尝试保存到一个包含空字符的无效文件名
	invalidPath := "test\x00invalid.docx"
	err = doc.Save(invalidPath)
	if err == nil {
		// 如果第一个测试没有失败，尝试另一个策略
		// 尝试保存到一个超长路径
		longPath := string(make([]byte, 300)) + ".docx"
		err = doc.Save(longPath)
		if err == nil {
			t.Log("Warning: Unable to trigger save error - filesystem may be permissive")
		}
	}
}

// TestStyleIntegration 测试样式集成
func TestStyleIntegration(t *testing.T) {
	doc := New()
	styleManager := doc.GetStyleManager()
	quickAPI := style.NewQuickStyleAPI(styleManager)

	// 创建自定义样式
	config := style.QuickStyleConfig{
		ID:      "TestStyle",
		Name:    "测试样式",
		Type:    style.StyleTypeParagraph,
		BasedOn: "Normal",
		RunConfig: &style.QuickRunConfig{
			Bold:      true,
			FontColor: "FF0000",
		},
	}

	_, err := quickAPI.CreateQuickStyle(config)
	if err != nil {
		t.Fatalf("Failed to create custom style: %v", err)
	}

	// 使用自定义样式
	para := doc.AddParagraph("使用自定义样式")
	para.SetStyle("TestStyle")

	// 验证样式应用
	if para.Properties == nil || para.Properties.ParagraphStyle == nil {
		t.Fatal("Style should be applied to paragraph")
	}

	if para.Properties.ParagraphStyle.Val != "TestStyle" {
		t.Errorf("Expected TestStyle, got %s", para.Properties.ParagraphStyle.Val)
	}

	// 验证样式存在
	if !styleManager.StyleExists("TestStyle") {
		t.Error("Custom style should exist in style manager")
	}
}

// BenchmarkAddParagraph 基准测试 - 添加段落性能
func BenchmarkAddParagraph(b *testing.B) {
	doc := New()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doc.AddParagraph("基准测试段落")
	}
}

// BenchmarkDocumentSave 基准测试 - 文档保存性能
func BenchmarkDocumentSave(b *testing.B) {
	doc := New()

	// 创建一个中等大小的文档
	for i := 0; i < 100; i++ {
		doc.AddParagraph("基准测试段落内容")
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		filename := "benchmark_save.docx"
		err := doc.Save(filename)
		if err != nil {
			b.Fatalf("Failed to save: %v", err)
		}
		os.Remove(filename)
	}
}

// TestTextFormatValidation 测试文本格式验证
func TestTextFormatValidation(t *testing.T) {
	doc := New()

	// 测试颜色格式
	testCases := []struct {
		color    string
		expected string
	}{
		{"#FF0000", "FF0000"}, // 带#前缀
		{"FF0000", "FF0000"},  // 不带#前缀
		{"#123456", "123456"},
		{"ABCDEF", "ABCDEF"},
	}

	for _, tc := range testCases {
		format := &TextFormat{
			FontColor: tc.color,
		}

		para := doc.AddFormattedParagraph("测试颜色", format)
		if para.Runs[0].Properties.Color.Val != tc.expected {
			t.Errorf("Color %s should be formatted as %s, got %s",
				tc.color, tc.expected, para.Runs[0].Properties.Color.Val)
		}
	}
}

// TestMemoryUsage 测试内存使用
func TestMemoryUsage(t *testing.T) {
	doc := New()

	// 添加大量段落测试内存使用
	const numParagraphs = 1000
	for i := 0; i < numParagraphs; i++ {
		doc.AddParagraph("内存测试段落")
	}

	if len(doc.Body.GetParagraphs()) != numParagraphs {
		t.Errorf("Expected %d paragraphs, got %d", numParagraphs, len(doc.Body.GetParagraphs()))
	}

	// 测试保存大文档
	filename := "test_memory.docx"
	defer os.Remove(filename)

	err := doc.Save(filename)
	if err != nil {
		t.Fatalf("Failed to save large document: %v", err)
	}
}

func TestDocumentOpenFromMemory(t *testing.T) {
	// 先创建一个测试文档
	originalDoc := New()
	originalDoc.AddParagraph("第一段")
	originalDoc.AddParagraph("第二段")
	originalDoc.AddHeadingParagraph("标题", 1)

	filename := "test_open.docx"
	defer os.Remove(filename)

	err := originalDoc.Save(filename)
	if err != nil {
		t.Fatalf("Failed to save test document: %v", err)
	}

	// 打开文档
	files, err := os.Open(filename)
	if err != nil {
		t.Fatalf("Failed to open test document: %v", err)
	}
	defer files.Close()

	loadedDoc, err := OpenFromMemory(files)
	if err != nil {
		t.Fatalf("Failed to open document: %v", err)
	}

	for _, paragraphs := range loadedDoc.Body.GetParagraphs() {
		for _, run := range paragraphs.Runs {
			t.Log(run.Text.Content)
		}
	}
}

// TestAddPageBreak 测试添加分页符功能
func TestAddPageBreak(t *testing.T) {
	doc := New()

	// 添加第一页内容
	doc.AddParagraph("第一页内容")

	// 添加分页符
	doc.AddPageBreak()

	// 添加第二页内容
	doc.AddParagraph("第二页内容")

	// 验证文档包含3个元素（段落、分页符段落、段落）
	if len(doc.Body.Elements) != 3 {
		t.Errorf("期望文档包含3个元素，实际包含 %d 个", len(doc.Body.Elements))
	}

	// 验证第二个元素是包含分页符的段落
	if p, ok := doc.Body.Elements[1].(*Paragraph); ok {
		if len(p.Runs) == 0 || p.Runs[0].Break == nil {
			t.Error("第二个元素应该是包含分页符的段落")
		} else if p.Runs[0].Break.Type != "page" {
			t.Errorf("分页符类型应该是 'page'，实际是 '%s'", p.Runs[0].Break.Type)
		}
	} else {
		t.Error("第二个元素应该是段落类型")
	}

	// 保存并验证文档可以正常生成
	filename := "test_page_break.docx"
	defer os.Remove(filename)

	err := doc.Save(filename)
	if err != nil {
		t.Fatalf("保存包含分页符的文档失败: %v", err)
	}

	// 验证文件存在
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Error("保存的文档文件不存在")
	}
}

// TestParagraphAddPageBreak 测试段落内添加分页符功能
func TestParagraphAddPageBreak(t *testing.T) {
	doc := New()

	// 创建一个段落并添加分页符
	para := doc.AddParagraph("分页符前的内容")
	para.AddPageBreak()
	para.AddFormattedText("分页符后的内容", nil)

	// 验证段落包含3个运行（文本、分页符、文本）
	if len(para.Runs) != 3 {
		t.Errorf("期望段落包含3个运行，实际包含 %d 个", len(para.Runs))
	}

	// 验证第二个运行是分页符
	if para.Runs[1].Break == nil {
		t.Error("第二个运行应该是分页符")
	} else if para.Runs[1].Break.Type != "page" {
		t.Errorf("分页符类型应该是 'page'，实际是 '%s'", para.Runs[1].Break.Type)
	}

	// 保存并验证文档可以正常生成
	filename := "test_paragraph_page_break.docx"
	defer os.Remove(filename)

	err := doc.Save(filename)
	if err != nil {
		t.Fatalf("保存包含段落内分页符的文档失败: %v", err)
	}

	// 验证文件存在
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Error("保存的文档文件不存在")
	}
}

// TestRemoveParagraph 测试删除段落功能
func TestRemoveParagraph(t *testing.T) {
	doc := New()

	// 添加三个段落
	para1 := doc.AddParagraph("第一段")
	para2 := doc.AddParagraph("第二段")
	para3 := doc.AddParagraph("第三段")

	// 验证初始状态
	if len(doc.Body.Elements) != 3 {
		t.Fatalf("期望文档包含3个段落，实际包含 %d 个", len(doc.Body.Elements))
	}

	// 删除第二个段落
	if !doc.RemoveParagraph(para2) {
		t.Error("删除段落应该成功")
	}

	// 验证删除后的状态
	if len(doc.Body.Elements) != 2 {
		t.Errorf("删除后期望文档包含2个段落，实际包含 %d 个", len(doc.Body.Elements))
	}

	// 验证剩余的段落是正确的
	paragraphs := doc.Body.GetParagraphs()
	if len(paragraphs) != 2 {
		t.Fatalf("期望获取到2个段落，实际获取到 %d 个", len(paragraphs))
	}

	if paragraphs[0] != para1 {
		t.Error("第一个段落应该是 para1")
	}
	if paragraphs[1] != para3 {
		t.Error("第二个段落应该是 para3")
	}

	// 尝试删除已删除的段落（应该返回false）
	if doc.RemoveParagraph(para2) {
		t.Error("删除不存在的段落应该返回false")
	}
}

// TestRemoveParagraphAt 测试按索引删除段落功能
func TestRemoveParagraphAt(t *testing.T) {
	doc := New()

	// 添加三个段落
	doc.AddParagraph("第一段")
	doc.AddParagraph("第二段")
	doc.AddParagraph("第三段")

	// 验证初始状态
	paragraphs := doc.Body.GetParagraphs()
	if len(paragraphs) != 3 {
		t.Fatalf("期望文档包含3个段落，实际包含 %d 个", len(paragraphs))
	}

	// 删除索引为1的段落（第二段）
	if !doc.RemoveParagraphAt(1) {
		t.Error("删除索引1的段落应该成功")
	}

	// 验证删除后的状态
	paragraphs = doc.Body.GetParagraphs()
	if len(paragraphs) != 2 {
		t.Errorf("删除后期望文档包含2个段落，实际包含 %d 个", len(paragraphs))
	}

	// 验证剩余段落的内容
	assertParagraphContent(t, paragraphs, 0, "第一段")
	assertParagraphContent(t, paragraphs, 1, "第三段")

	// 尝试删除超出范围的索引
	if doc.RemoveParagraphAt(10) {
		t.Error("删除超出范围的索引应该返回false")
	}

	if doc.RemoveParagraphAt(-1) {
		t.Error("删除负数索引应该返回false")
	}
}

// TestRemoveElementAt 测试按元素索引删除功能
func TestRemoveElementAt(t *testing.T) {
	doc := New()

	// 添加段落和表格
	doc.AddParagraph("段落1")
	doc.AddTable(&TableConfig{Rows: 2, Cols: 2})
	doc.AddParagraph("段落2")

	// 验证初始状态
	if len(doc.Body.Elements) != 3 {
		t.Fatalf("期望文档包含3个元素，实际包含 %d 个", len(doc.Body.Elements))
	}

	// 删除索引为1的元素（表格）
	if !doc.RemoveElementAt(1) {
		t.Error("删除索引1的元素应该成功")
	}

	// 验证删除后的状态
	if len(doc.Body.Elements) != 2 {
		t.Errorf("删除后期望文档包含2个元素，实际包含 %d 个", len(doc.Body.Elements))
	}

	// 验证剩余的都是段落
	paragraphs := doc.Body.GetParagraphs()
	if len(paragraphs) != 2 {
		t.Errorf("期望获取到2个段落，实际获取到 %d 个", len(paragraphs))
	}

	// 尝试删除超出范围的索引
	if doc.RemoveElementAt(10) {
		t.Error("删除超出范围的索引应该返回false")
	}
}

// TestPageBreakAndDeletion 综合测试分页符和删除功能
func TestPageBreakAndDeletion(t *testing.T) {
	doc := New()

	// 创建一个包含分页符的文档
	doc.AddParagraph("第一页 - 段落1")
	doc.AddParagraph("第一页 - 段落2")
	doc.AddPageBreak()
	doc.AddParagraph("第二页 - 段落1")
	doc.AddPageBreak()
	doc.AddParagraph("第三页 - 段落1")

	// 验证初始状态（2个段落 + 1个分页符 + 1个段落 + 1个分页符 + 1个段落 = 6个元素）
	if len(doc.Body.Elements) != 6 {
		t.Fatalf("期望文档包含6个元素，实际包含 %d 个", len(doc.Body.Elements))
	}

	// 删除第一个分页符（索引2）
	if !doc.RemoveElementAt(2) {
		t.Error("删除分页符应该成功")
	}

	// 验证删除后的状态
	if len(doc.Body.Elements) != 5 {
		t.Errorf("删除后期望文档包含5个元素，实际包含 %d 个", len(doc.Body.Elements))
	}

	// 保存文档并验证
	filename := "test_pagebreak_deletion.docx"
	defer os.Remove(filename)

	err := doc.Save(filename)
	if err != nil {
		t.Fatalf("保存文档失败: %v", err)
	}

	// 验证文件存在
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Error("保存的文档文件不存在")
	}
}

// TestAddFormattedHeader 测试添加格式化页眉
func TestAddFormattedHeader(t *testing.T) {
	doc := New()

	// 测试添加格式化页眉
	config := &HeaderFooterConfig{
		Text: "公司报告",
		Format: &TextFormat{
			FontSize:   10,
			FontColor:  "8e8e8e",
			FontFamily: "Arial",
		},
		Alignment: AlignCenter,
	}

	err := doc.AddFormattedHeader(HeaderFooterTypeDefault, config)
	if err != nil {
		t.Fatalf("添加格式化页眉失败: %v", err)
	}

	// 验证页眉文件被创建
	headerPartName := "word/header1.xml"
	if _, ok := doc.parts[headerPartName]; !ok {
		t.Errorf("页眉文件 %s 未创建", headerPartName)
	}

	// 保存并验证文档
	filename := "test_formatted_header.docx"
	defer os.Remove(filename)

	err = doc.Save(filename)
	if err != nil {
		t.Fatalf("保存文档失败: %v", err)
	}

	// 验证文件存在
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Error("保存的文档文件不存在")
	}
}

// TestAddFormattedFooter 测试添加格式化页脚
func TestAddFormattedFooter(t *testing.T) {
	doc := New()

	// 测试添加格式化页脚
	config := &HeaderFooterConfig{
		Text: "第 1 页",
		Format: &TextFormat{
			FontSize:   9,
			FontColor:  "666666",
			FontFamily: "宋体",
			Bold:       true,
		},
		Alignment: AlignCenter,
	}

	err := doc.AddFormattedFooter(HeaderFooterTypeDefault, config)
	if err != nil {
		t.Fatalf("添加格式化页脚失败: %v", err)
	}

	// 验证页脚文件被创建
	footerPartName := "word/footer1.xml"
	if _, ok := doc.parts[footerPartName]; !ok {
		t.Errorf("页脚文件 %s 未创建", footerPartName)
	}

	// 保存并验证文档
	filename := "test_formatted_footer.docx"
	defer os.Remove(filename)

	err = doc.Save(filename)
	if err != nil {
		t.Fatalf("保存文档失败: %v", err)
	}

	// 验证文件存在
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Error("保存的文档文件不存在")
	}
}

// TestAddFormattedHeaderWithNilConfig 测试使用空配置添加页眉
func TestAddFormattedHeaderWithNilConfig(t *testing.T) {
	doc := New()

	// 测试使用nil配置添加页眉
	err := doc.AddFormattedHeader(HeaderFooterTypeDefault, nil)
	if err != nil {
		t.Fatalf("使用nil配置添加页眉失败: %v", err)
	}

	// 验证页眉文件被创建
	headerPartName := "word/header1.xml"
	if _, ok := doc.parts[headerPartName]; !ok {
		t.Errorf("页眉文件 %s 未创建", headerPartName)
	}
}

// TestAddFormattedHeaderWithAllFormats 测试所有格式选项
func TestAddFormattedHeaderWithAllFormats(t *testing.T) {
	doc := New()

	// 测试所有格式选项
	config := &HeaderFooterConfig{
		Text: "格式化测试",
		Format: &TextFormat{
			Bold:       true,
			Italic:     true,
			FontSize:   12,
			FontColor:  "FF0000",
			FontFamily: "Times New Roman",
			Underline:  true,
			Strike:     true,
			Highlight:  "yellow",
		},
		Alignment: AlignRight,
	}

	err := doc.AddFormattedHeader(HeaderFooterTypeDefault, config)
	if err != nil {
		t.Fatalf("添加格式化页眉失败: %v", err)
	}

	// 保存并验证文档
	filename := "test_all_formats_header.docx"
	defer os.Remove(filename)

	err = doc.Save(filename)
	if err != nil {
		t.Fatalf("保存文档失败: %v", err)
	}
}

// TestCreateFormattedParagraph 测试createFormattedParagraph函数
func TestCreateFormattedParagraph(t *testing.T) {
	// 测试基本段落创建
	para := createFormattedParagraph("测试文本", nil, "")
	if len(para.Runs) != 1 {
		t.Errorf("期望1个Run，实际得到%d个", len(para.Runs))
	}
	if para.Runs[0].Text.Content != "测试文本" {
		t.Errorf("期望文本'测试文本'，实际得到'%s'", para.Runs[0].Text.Content)
	}

	// 测试带对齐方式的段落
	para2 := createFormattedParagraph("居中文本", nil, AlignCenter)
	if para2.Properties == nil {
		t.Fatal("段落属性不应为nil")
	}
	if para2.Properties.Justification == nil {
		t.Fatal("对齐方式属性不应为nil")
	}
	if para2.Properties.Justification.Val != string(AlignCenter) {
		t.Errorf("期望对齐方式'center'，实际得到'%s'", para2.Properties.Justification.Val)
	}

	// 测试带格式的段落
	format := &TextFormat{
		Bold:       true,
		FontSize:   14,
		FontColor:  "0000FF",
		FontFamily: "Arial",
	}
	para3 := createFormattedParagraph("格式化文本", format, AlignLeft)
	if para3.Runs[0].Properties == nil {
		t.Fatal("Run属性不应为nil")
	}
	if para3.Runs[0].Properties.Bold == nil {
		t.Error("粗体属性不应为nil")
	}
	if para3.Runs[0].Properties.FontSize == nil {
		t.Error("字体大小属性不应为nil")
	}
	if para3.Runs[0].Properties.FontSize.Val != "28" { // 14 * 2 = 28
		t.Errorf("期望字体大小'28'，实际得到'%s'", para3.Runs[0].Properties.FontSize.Val)
	}
	if para3.Runs[0].Properties.Color == nil {
		t.Error("颜色属性不应为nil")
	}
	if para3.Runs[0].Properties.Color.Val != "0000FF" {
		t.Errorf("期望颜色'0000FF'，实际得到'%s'", para3.Runs[0].Properties.Color.Val)
	}
	if para3.Runs[0].Properties.FontFamily == nil {
		t.Error("字体属性不应为nil")
	}
	if para3.Runs[0].Properties.FontFamily.ASCII != "Arial" {
		t.Errorf("期望字体'Arial'，实际得到'%s'", para3.Runs[0].Properties.FontFamily.ASCII)
	}

	// 测试空文本
	para4 := createFormattedParagraph("", nil, AlignCenter)
	if len(para4.Runs) != 0 {
		t.Errorf("空文本应该不添加Run，实际得到%d个", len(para4.Runs))
	}
}
