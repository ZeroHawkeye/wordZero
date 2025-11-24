package document

import (
	"os"
	"strconv"
	"strings"
	"testing"
)

// TestChineseFilename 测试中文文件名的图片是否能正常保存和打开
func TestChineseFilename(t *testing.T) {
	doc := New()
	doc.AddParagraph("测试中文文件名")

	// 创建测试图片
	imageData := createTestImage(100, 75)

	// 添加使用中文文件名的图片
	_, err := doc.AddImageFromData(
		imageData,
		"测试图片.png", // 中文文件名
		ImageFormatPNG,
		100, 75,
		&ImageConfig{
			Position:  ImagePositionInline,
			Alignment: AlignCenter,
			AltText:   "测试图片",
			Title:     "测试图片标题",
		},
	)
	if err != nil {
		t.Fatalf("添加中文文件名图片失败: %v", err)
	}

	doc.AddParagraph("图片下方的文字")

	// 保存文档
	testFile := "test_chinese_filename.docx"
	err = doc.Save(testFile)
	if err != nil {
		t.Fatalf("保存文档失败: %v", err)
	}
	defer os.Remove(testFile)

	// 验证图片使用安全的文件名存储（image0.png而不是测试图片.png）
	foundSafeFilename := false
	foundChineseFilename := false
	for partName := range doc.parts {
		if strings.Contains(partName, "word/media/") {
			if strings.Contains(partName, "image0.png") {
				foundSafeFilename = true
			}
			if strings.Contains(partName, "测试") {
				foundChineseFilename = true
			}
			t.Logf("找到图片: %s", partName)
		}
	}

	if !foundSafeFilename {
		t.Error("未找到安全文件名(image0.png)，中文文件名转换失败")
	}

	if foundChineseFilename {
		t.Error("找到中文文件名，应该已经转换为安全的ASCII文件名")
	}

	// 验证关系中也使用安全的文件名
	foundImageRelationship := false
	for _, rel := range doc.documentRelationships.Relationships {
		if rel.Type == "http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" {
			foundImageRelationship = true
			if !strings.Contains(rel.Target, "image0.png") {
				t.Errorf("图片关系未使用安全文件名，Target=%s", rel.Target)
			}
			if strings.Contains(rel.Target, "测试") {
				t.Errorf("图片关系包含中文字符，Target=%s", rel.Target)
			}
			t.Logf("图片关系: ID=%s, Target=%s", rel.ID, rel.Target)
			break
		}
	}

	if !foundImageRelationship {
		t.Error("未找到图片关系")
	}

	// 打开文档验证
	doc2, err := Open(testFile)
	if err != nil {
		t.Fatalf("打开文档失败: %v", err)
	}

	// 验证图片数据存在
	if _, exists := doc2.parts["word/media/image0.png"]; !exists {
		t.Error("打开的文档中未找到图片数据")
	}

	t.Log("✓ 中文文件名测试通过：自动转换为安全的ASCII文件名")
}

// TestMultipleNonASCIIFilenames 测试多个非ASCII文件名的图片
func TestMultipleNonASCIIFilenames(t *testing.T) {
	doc := New()
	doc.AddParagraph("测试多个非ASCII文件名")

	imageData := createTestImage(50, 50)

	// 添加多个使用不同语言文件名的图片
	testFilenames := []string{
		"中文图片.png",    // 中文
		"日本語.png",     // 日文
		"한국어.png",     // 韩文
		"Русский.png", // 俄文
		"العربية.png", // 阿拉伯文
	}

	for i, filename := range testFilenames {
		_, err := doc.AddImageFromData(imageData, filename, ImageFormatPNG, 50, 50, nil)
		if err != nil {
			t.Fatalf("添加图片 %s 失败: %v", filename, err)
		}

		// 验证每个图片都使用了安全的文件名
		expectedSafeFilename := "image" + strconv.Itoa(i) + ".png"
		if _, exists := doc.parts["word/media/"+expectedSafeFilename]; !exists {
			t.Errorf("图片 %s 未使用安全文件名 %s", filename, expectedSafeFilename)
		}
	}

	// 保存并重新打开
	testFile := "test_multiple_nonascii_filenames.docx"
	err := doc.Save(testFile)
	if err != nil {
		t.Fatalf("保存文档失败: %v", err)
	}
	defer os.Remove(testFile)

	doc2, err := Open(testFile)
	if err != nil {
		t.Fatalf("打开文档失败: %v", err)
	}

	// 验证所有图片都存在
	for i := 0; i < len(testFilenames); i++ {
		expectedSafeFilename := "image" + strconv.Itoa(i) + ".png"
		if _, exists := doc2.parts["word/media/"+expectedSafeFilename]; !exists {
			t.Errorf("打开的文档中未找到图片 %s", expectedSafeFilename)
		}
	}

	t.Logf("✓ 多语言文件名测试通过：所有 %d 个图片都正确转换", len(testFilenames))
}
