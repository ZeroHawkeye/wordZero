package document

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"
)

// TestImagePersistenceAfterOpenAndSave 测试打开包含图片的文档并重新保存后，图片是否仍然存在
func TestImagePersistenceAfterOpenAndSave(t *testing.T) {
	// 步骤1: 创建一个包含图片的文档
	doc1 := New()
	doc1.AddParagraph("测试文档 - 图片持久性测试")
	
	// 创建测试图片
	imageData := createTestImageForPersistence(100, 75, color.RGBA{255, 100, 100, 255})
	
	// 添加图片（文件名会被自动转换为安全的image0.png）
	imageInfo, err := doc1.AddImageFromData(
		imageData,
		"test_image.png",
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
		t.Fatalf("添加图片失败: %v", err)
	}
	
	doc1.AddParagraph("图片下方的文字")
	
	// 保存第一个文档
	testFile1 := "test_image_persistence_1.docx"
	err = doc1.Save(testFile1)
	if err != nil {
		t.Fatalf("保存第一个文档失败: %v", err)
	}
	defer os.Remove(testFile1)
	
	// 验证第一个文档中包含图片数据（使用安全文件名image0.png）
	if _, exists := doc1.parts["word/media/image0.png"]; !exists {
		t.Fatal("第一个文档中没有找到图片数据")
	}
	
	// 步骤2: 打开刚刚保存的文档
	doc2, err := Open(testFile1)
	if err != nil {
		t.Fatalf("打开文档失败: %v", err)
	}
	
	// 验证打开的文档包含图片数据
	if _, exists := doc2.parts["word/media/image0.png"]; !exists {
		t.Fatal("打开的文档中没有找到图片数据")
	}
	
	// 验证文档关系中包含图片关系
	foundImageRelationship := false
	for _, rel := range doc2.documentRelationships.Relationships {
		if rel.Type == "http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" {
			foundImageRelationship = true
			t.Logf("找到图片关系: ID=%s, Target=%s", rel.ID, rel.Target)
			break
		}
	}
	if !foundImageRelationship {
		t.Fatal("打开的文档中没有找到图片关系")
	}
	
	// 步骤3: 修改文档并保存为新文件
	doc2.AddParagraph("这是新添加的段落")
	
	testFile2 := "test_image_persistence_2.docx"
	err = doc2.Save(testFile2)
	if err != nil {
		t.Fatalf("保存第二个文档失败: %v", err)
	}
	defer os.Remove(testFile2)
	
	// 步骤4: 打开第二个文档，验证图片是否仍然存在
	doc3, err := Open(testFile2)
	if err != nil {
		t.Fatalf("打开第二个文档失败: %v", err)
	}
	
	// 验证图片数据仍然存在
	if _, exists := doc3.parts["word/media/image0.png"]; !exists {
		t.Fatal("【问题】第二个文档中没有找到图片数据 - 图片在保存后丢失！")
	}
	
	// 验证图片关系仍然存在
	foundImageRelationship = false
	for _, rel := range doc3.documentRelationships.Relationships {
		if rel.Type == "http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" {
			foundImageRelationship = true
			t.Logf("第二个文档中找到图片关系: ID=%s, Target=%s", rel.ID, rel.Target)
			break
		}
	}
	if !foundImageRelationship {
		t.Fatal("【问题】第二个文档中没有找到图片关系 - 图片关系在保存后丢失！")
	}
	
	// 验证图片数据完整性
	originalImageData := doc1.parts["word/media/image0.png"]
	finalImageData := doc3.parts["word/media/image0.png"]
	
	if !bytes.Equal(originalImageData, finalImageData) {
		t.Fatal("图片数据在保存和重新打开后发生了变化")
	}
	
	t.Log("✓ 图片持久性测试通过：图片在修改文档并保存后仍然存在")
	t.Logf("✓ 原始图片信息: ID=%s, 格式=%s, 尺寸=%dx%d", 
		imageInfo.ID, imageInfo.Format, imageInfo.Width, imageInfo.Height)
}

// TestAddImageToOpenedDocument 测试向已打开的文档添加新图片
func TestAddImageToOpenedDocument(t *testing.T) {
	// 步骤1: 创建包含一张图片的文档
	doc1 := New()
	doc1.AddParagraph("原始文档")
	
	// 添加第一张图片（红色）- 将被保存为image0.png
	imageData1 := createTestImageForPersistence(100, 75, color.RGBA{255, 0, 0, 255})
	_, err := doc1.AddImageFromData(
		imageData1,
		"image1.png",
		ImageFormatPNG,
		100, 75,
		&ImageConfig{
			Position:  ImagePositionInline,
			Alignment: AlignCenter,
		},
	)
	if err != nil {
		t.Fatalf("添加第一张图片失败: %v", err)
	}
	
	// 保存文档
	testFile1 := "test_add_image_to_opened_1.docx"
	err = doc1.Save(testFile1)
	if err != nil {
		t.Fatalf("保存文档失败: %v", err)
	}
	defer os.Remove(testFile1)
	
	// 步骤2: 打开文档并添加第二张图片
	doc2, err := Open(testFile1)
	if err != nil {
		t.Fatalf("打开文档失败: %v", err)
	}
	
	doc2.AddParagraph("添加第二张图片")
	
	// 添加第二张图片（蓝色）- 将被保存为image1.png
	imageData2 := createTestImageForPersistence(100, 75, color.RGBA{0, 0, 255, 255})
	_, err = doc2.AddImageFromData(
		imageData2,
		"image2.png",
		ImageFormatPNG,
		100, 75,
		&ImageConfig{
			Position:  ImagePositionInline,
			Alignment: AlignCenter,
		},
	)
	if err != nil {
		t.Fatalf("添加第二张图片失败: %v", err)
	}
	
	// 保存文档
	testFile2 := "test_add_image_to_opened_2.docx"
	err = doc2.Save(testFile2)
	if err != nil {
		t.Fatalf("保存包含两张图片的文档失败: %v", err)
	}
	defer os.Remove(testFile2)
	
	// 步骤3: 打开文档，验证两张图片都存在
	doc3, err := Open(testFile2)
	if err != nil {
		t.Fatalf("打开包含两张图片的文档失败: %v", err)
	}
	
	// 验证两张图片数据都存在（现在使用安全文件名image0.png和image1.png）
	if _, exists := doc3.parts["word/media/image0.png"]; !exists {
		t.Fatal("【问题】第一张图片数据丢失")
	}
	
	if _, exists := doc3.parts["word/media/image1.png"]; !exists {
		t.Fatal("【问题】第二张图片数据丢失")
	}
	
	// 验证图片关系数量
	imageRelCount := 0
	for _, rel := range doc3.documentRelationships.Relationships {
		if rel.Type == "http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" {
			imageRelCount++
			t.Logf("找到图片关系: ID=%s, Target=%s", rel.ID, rel.Target)
		}
	}
	
	if imageRelCount != 2 {
		t.Fatalf("期望有2个图片关系，实际有 %d 个", imageRelCount)
	}
	
	t.Log("✓ 向已打开的文档添加图片测试通过：两张图片都正确保存")
}

// TestImageIDCounterAfterOpen 测试打开文档后图片ID计数器是否正确更新
func TestImageIDCounterAfterOpen(t *testing.T) {
	// 步骤1: 创建包含两张图片的文档
	doc1 := New()
	doc1.AddParagraph("测试图片ID计数器")
	
	// 添加两张图片（将被保存为image0.png和image1.png）
	imageData := createTestImageForPersistence(50, 50, color.RGBA{255, 0, 0, 255})
	
	_, err := doc1.AddImageFromData(imageData, "img1.png", ImageFormatPNG, 50, 50, nil)
	if err != nil {
		t.Fatalf("添加第一张图片失败: %v", err)
	}
	
	_, err = doc1.AddImageFromData(imageData, "img2.png", ImageFormatPNG, 50, 50, nil)
	if err != nil {
		t.Fatalf("添加第二张图片失败: %v", err)
	}
	
	// 保存文档
	testFile := "test_image_id_counter.docx"
	err = doc1.Save(testFile)
	if err != nil {
		t.Fatalf("保存文档失败: %v", err)
	}
	defer os.Remove(testFile)
	
	// 步骤2: 打开文档
	doc2, err := Open(testFile)
	if err != nil {
		t.Fatalf("打开文档失败: %v", err)
	}
	
	// 验证nextImageID已正确更新
	// doc1有两张图片，使用了rId2和rId3（rId1是styles.xml）
	// 所以打开后nextImageID应该至少为2（最大rId为3）
	if doc2.nextImageID < 2 {
		t.Fatalf("nextImageID未正确更新：期望 >= 2，实际 = %d", doc2.nextImageID)
	}
	
	t.Logf("✓ 打开文档后nextImageID = %d（符合预期）", doc2.nextImageID)
	
	// 步骤3: 添加第三张图片
	_, err = doc2.AddImageFromData(imageData, "img3.png", ImageFormatPNG, 50, 50, nil)
	if err != nil {
		t.Fatalf("添加第三张图片失败: %v", err)
	}
	
	// 保存并重新打开，验证所有三张图片都存在
	testFile2 := "test_image_id_counter_2.docx"
	err = doc2.Save(testFile2)
	if err != nil {
		t.Fatalf("保存包含三张图片的文档失败: %v", err)
	}
	defer os.Remove(testFile2)
	
	doc3, err := Open(testFile2)
	if err != nil {
		t.Fatalf("打开包含三张图片的文档失败: %v", err)
	}
	
	// 验证三张图片都存在（使用安全文件名image0.png, image1.png, image2.png）
	images := []string{"image0.png", "image1.png", "image2.png"}
	for _, imgName := range images {
		if _, exists := doc3.parts["word/media/"+imgName]; !exists {
			t.Fatalf("【问题】图片 %s 丢失", imgName)
		}
	}
	
	// 验证图片关系数量
	imageRelCount := 0
	for _, rel := range doc3.documentRelationships.Relationships {
		if rel.Type == "http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" {
			imageRelCount++
		}
	}
	
	if imageRelCount != 3 {
		t.Fatalf("期望有3个图片关系，实际有 %d 个", imageRelCount)
	}
	
	t.Log("✓ 图片ID计数器测试通过：所有图片ID都正确且无冲突")
}

// createTestImageForPersistence 创建用于持久性测试的图片
func createTestImageForPersistence(width, height int, bgColor color.RGBA) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	
	// 填充背景色
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, bgColor)
		}
	}
	
	// 添加边框
	borderColor := color.RGBA{0, 0, 0, 255}
	for x := 0; x < width; x++ {
		img.Set(x, 0, borderColor)
		img.Set(x, height-1, borderColor)
	}
	for y := 0; y < height; y++ {
		img.Set(0, y, borderColor)
		img.Set(width-1, y, borderColor)
	}
	
	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}
