package document

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"
)

// createTestImage 创建一个测试用的PNG图片
func createTestImage(width, height int) []byte {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// 创建一个简单的红色矩形
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}

	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}

func TestDetectImageFormat(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected ImageFormat
		hasError bool
	}{
		{
			name:     "PNG格式",
			data:     []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A},
			expected: ImageFormatPNG,
			hasError: false,
		},
		{
			name:     "JPEG格式",
			data:     []byte{0xFF, 0xD8, 0xFF},
			expected: ImageFormatJPEG,
			hasError: false,
		},
		{
			name:     "GIF87a格式",
			data:     []byte("GIF87a"),
			expected: ImageFormatGIF,
			hasError: false,
		},
		{
			name:     "GIF89a格式",
			data:     []byte("GIF89a"),
			expected: ImageFormatGIF,
			hasError: false,
		},
		{
			name:     "数据太短",
			data:     []byte{0x89},
			expected: "",
			hasError: true,
		},
		{
			name:     "不支持的格式",
			data:     []byte("INVALID_FORMAT"),
			expected: "",
			hasError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			format, err := detectImageFormat(tt.data)

			if tt.hasError {
				if err == nil {
					t.Errorf("期望有错误，但没有错误")
				}
			} else {
				if err != nil {
					t.Errorf("不期望有错误，但出现错误: %v", err)
				}
				if format != tt.expected {
					t.Errorf("期望格式 %v，但得到 %v", tt.expected, format)
				}
			}
		})
	}
}

func TestGetImageDimensions(t *testing.T) {
	// 创建一个100x50的测试图片
	testImageData := createTestImage(100, 50)

	width, height, err := getImageDimensions(testImageData, ImageFormatPNG)
	if err != nil {
		t.Fatalf("获取图片尺寸失败: %v", err)
	}

	if width != 100 {
		t.Errorf("期望宽度100，得到 %d", width)
	}

	if height != 50 {
		t.Errorf("期望高度50，得到 %d", height)
	}
}

func TestCalculateDisplaySize(t *testing.T) {
	doc := New()

	tests := []struct {
		name      string
		imageInfo *ImageInfo
		expectedW int64
		expectedH int64
	}{
		{
			name: "默认尺寸",
			imageInfo: &ImageInfo{
				Width:  100,
				Height: 50,
				Config: nil,
			},
			expectedW: 100 * 9525, // 像素转EMU
			expectedH: 50 * 9525,
		},
		{
			name: "指定具体尺寸",
			imageInfo: &ImageInfo{
				Width:  100,
				Height: 50,
				Config: &ImageConfig{
					Size: &ImageSize{
						Width:  50.0, // 50毫米
						Height: 25.0, // 25毫米
					},
				},
			},
			expectedW: int64(50.0 * 36000), // 毫米转EMU
			expectedH: int64(25.0 * 36000),
		},
		{
			name: "只指定宽度，保持长宽比",
			imageInfo: &ImageInfo{
				Width:  100,
				Height: 50,
				Config: &ImageConfig{
					Size: &ImageSize{
						Width:           50.0, // 50毫米
						KeepAspectRatio: true,
					},
				},
			},
			expectedW: int64(50.0 * 36000),
			expectedH: int64(50.0 * 36000 * 0.5), // 保持长宽比2:1
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			width, height := doc.calculateDisplaySize(tt.imageInfo)

			if width != tt.expectedW {
				t.Errorf("期望宽度 %d，得到 %d", tt.expectedW, width)
			}

			if height != tt.expectedH {
				t.Errorf("期望高度 %d，得到 %d", tt.expectedH, height)
			}
		})
	}
}

func TestAddImageFromData(t *testing.T) {
	doc := New()

	// 创建测试图片数据
	imageData := createTestImage(100, 50)

	// 添加图片
	imageInfo, err := doc.AddImageFromData(
		imageData,
		"test.png",
		ImageFormatPNG,
		100,
		50,
		&ImageConfig{
			AltText: "测试图片",
			Title:   "测试标题",
		},
	)

	if err != nil {
		t.Fatalf("添加图片失败: %v", err)
	}

	// 验证图片信息
	if imageInfo.Format != ImageFormatPNG {
		t.Errorf("期望格式PNG，得到 %v", imageInfo.Format)
	}

	if imageInfo.Width != 100 {
		t.Errorf("期望宽度100，得到 %d", imageInfo.Width)
	}

	if imageInfo.Height != 50 {
		t.Errorf("期望高度50，得到 %d", imageInfo.Height)
	}

	if imageInfo.Config.AltText != "测试图片" {
		t.Errorf("期望替代文字'测试图片'，得到 '%s'", imageInfo.Config.AltText)
	}

	// 验证关系是否正确添加
	if len(doc.documentRelationships.Relationships) != 1 {
		t.Errorf("期望1个关系，得到 %d", len(doc.documentRelationships.Relationships))
	}

	// 验证图片数据是否存储（现在使用安全的文件名 image0.png）
	if _, exists := doc.parts["word/media/image0.png"]; !exists {
		t.Error("图片数据未正确存储")
	}

	// 验证内容类型是否添加
	foundPNG := false
	for _, def := range doc.contentTypes.Defaults {
		if def.Extension == "png" && def.ContentType == "image/png" {
			foundPNG = true
			break
		}
	}
	if !foundPNG {
		t.Error("PNG内容类型未正确添加")
	}
}

func TestResizeImage(t *testing.T) {
	doc := New()

	imageInfo := &ImageInfo{
		Config: &ImageConfig{},
	}

	newSize := &ImageSize{
		Width:           100.0,
		Height:          50.0,
		KeepAspectRatio: true,
	}

	err := doc.ResizeImage(imageInfo, newSize)
	if err != nil {
		t.Fatalf("调整图片大小失败: %v", err)
	}

	if imageInfo.Config.Size != newSize {
		t.Error("图片大小未正确设置")
	}
}

func TestSetImagePosition(t *testing.T) {
	doc := New()

	imageInfo := &ImageInfo{
		Config: &ImageConfig{},
	}

	err := doc.SetImagePosition(imageInfo, ImagePositionFloatLeft, 10.0, 20.0)
	if err != nil {
		t.Fatalf("设置图片位置失败: %v", err)
	}

	if imageInfo.Config.Position != ImagePositionFloatLeft {
		t.Error("图片位置未正确设置")
	}

	if imageInfo.Config.OffsetX != 10.0 {
		t.Error("图片X偏移未正确设置")
	}

	if imageInfo.Config.OffsetY != 20.0 {
		t.Error("图片Y偏移未正确设置")
	}
}

func TestSetImageWrapText(t *testing.T) {
	doc := New()

	imageInfo := &ImageInfo{
		Config: &ImageConfig{},
	}

	err := doc.SetImageWrapText(imageInfo, ImageWrapSquare)
	if err != nil {
		t.Fatalf("设置图片文字环绕失败: %v", err)
	}

	if imageInfo.Config.WrapText != ImageWrapSquare {
		t.Error("图片文字环绕未正确设置")
	}
}

// TestFloatingImageXMLStructure 测试浮动图片XML结构修复
func TestFloatingImageXMLStructure(t *testing.T) {
	doc := New()

	// 创建测试图片数据
	imageData := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A} // PNG header

	// 测试左浮动 + 紧密环绕
	config := &ImageConfig{
		Position: ImagePositionFloatLeft,
		WrapText: ImageWrapTight,
		Size: &ImageSize{
			Width:  50,
			Height: 37.5,
		},
		AltText: "测试图片",
		Title:   "测试图片",
	}

	imageInfo, err := doc.AddImageFromData(imageData, "test.png", ImageFormatPNG, 100, 75, config)
	if err != nil {
		t.Fatalf("添加浮动图片失败: %v", err)
	}

	// 验证配置是否正确设置
	if imageInfo.Config.Position != ImagePositionFloatLeft {
		t.Error("图片位置未正确设置为左浮动")
	}

	if imageInfo.Config.WrapText != ImageWrapTight {
		t.Error("图片环绕类型未正确设置为紧密环绕")
	}

	// 保存文档并检查是否成功
	err = doc.Save("test_floating_fix.docx")
	if err != nil {
		t.Fatalf("保存包含修复后浮动图片的文档失败: %v", err)
	}

	// 清理测试文件
	defer func() {
		if err := os.Remove("test_floating_fix.docx"); err != nil {
			t.Logf("清理测试文件失败: %v", err)
		}
	}()

	t.Log("✓ 浮动图片XML结构修复测试通过")
}

// TestCreateDefaultWrapPolygon 测试默认环绕多边形创建
func TestCreateDefaultWrapPolygon(t *testing.T) {
	doc := New()

	polygon := doc.createDefaultWrapPolygon()
	if polygon == nil {
		t.Fatal("创建默认环绕多边形失败")
	}

	if polygon.Start == nil {
		t.Error("环绕多边形缺少起点")
	}

	if len(polygon.LineTo) == 0 {
		t.Error("环绕多边形缺少线段")
	}

	// 验证起点坐标
	if polygon.Start.X != "0" || polygon.Start.Y != "0" {
		t.Error("环绕多边形起点坐标不正确")
	}

	// 验证是否形成闭合路径
	expectedPoints := 4 // 矩形应该有4个点
	if len(polygon.LineTo) != expectedPoints {
		t.Errorf("期望%d个线段，实际%d个", expectedPoints, len(polygon.LineTo))
	}

	t.Log("✓ 默认环绕多边形创建测试通过")
}

func TestSetImageAltText(t *testing.T) {
	doc := New()

	imageInfo := &ImageInfo{
		Config: &ImageConfig{},
	}

	err := doc.SetImageAltText(imageInfo, "新的替代文字")
	if err != nil {
		t.Fatalf("设置图片替代文字失败: %v", err)
	}

	if imageInfo.Config.AltText != "新的替代文字" {
		t.Error("图片替代文字未正确设置")
	}
}

func TestSetImageTitle(t *testing.T) {
	doc := New()

	imageInfo := &ImageInfo{
		Config: &ImageConfig{},
	}

	err := doc.SetImageTitle(imageInfo, "新的标题")
	if err != nil {
		t.Fatalf("设置图片标题失败: %v", err)
	}

	if imageInfo.Config.Title != "新的标题" {
		t.Error("图片标题未正确设置")
	}
}

func TestAddImageContentType(t *testing.T) {
	doc := New()

	// 测试添加PNG内容类型
	doc.addImageContentType(ImageFormatPNG)

	found := false
	for _, def := range doc.contentTypes.Defaults {
		if def.Extension == "png" && def.ContentType == "image/png" {
			found = true
			break
		}
	}

	if !found {
		t.Error("PNG内容类型未正确添加")
	}

	// 测试重复添加同一类型
	originalCount := len(doc.contentTypes.Defaults)
	doc.addImageContentType(ImageFormatPNG)

	if len(doc.contentTypes.Defaults) != originalCount {
		t.Error("重复添加内容类型应该被忽略")
	}
}

func TestSetImageAlignment(t *testing.T) {
	doc := New()

	imageInfo := &ImageInfo{
		Config: &ImageConfig{},
	}

	// 测试各种对齐方式
	alignments := []AlignmentType{
		AlignLeft,
		AlignCenter,
		AlignRight,
		AlignJustify,
	}

	for _, alignment := range alignments {
		err := doc.SetImageAlignment(imageInfo, alignment)
		if err != nil {
			t.Fatalf("设置图片对齐方式失败: %v, 对齐方式: %s", err, alignment)
		}

		if imageInfo.Config.Alignment != alignment {
			t.Errorf("图片对齐方式未正确设置，预期: %s，实际: %s", alignment, imageInfo.Config.Alignment)
		}
	}
}

func TestImageParagraphAlignment(t *testing.T) {
	doc := New()

	// 创建测试图片数据
	imageData := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, // PNG头
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52, // IHDR
		0x00, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x01, // 1x1像素
		0x08, 0x06, 0x00, 0x00, 0x00, 0x1F, 0x15, 0xC4,
		0x89, 0x00, 0x00, 0x00, 0x0B, 0x49, 0x44, 0x41,
		0x54, 0x78, 0x9C, 0x63, 0x00, 0x01, 0x00, 0x00,
		0x05, 0x00, 0x01, 0x0D, 0x0A, 0x2D, 0xB4, 0x00,
		0x00, 0x00, 0x00, 0x49, 0x45, 0x4E, 0x44, 0xAE,
		0x42, 0x60, 0x82,
	}

	// 测试居中对齐的图片
	imageInfo, err := doc.AddImageFromData(
		imageData,
		"test.png",
		ImageFormatPNG,
		1, 1,
		&ImageConfig{
			Position:  ImagePositionInline,
			Alignment: AlignCenter,
		},
	)
	if err != nil {
		t.Fatalf("添加图片失败: %v", err)
	}

	// 验证图片段落的对齐设置
	if len(doc.Body.Elements) == 0 {
		t.Fatal("文档中没有添加段落")
	}

	paragraph, ok := doc.Body.Elements[0].(*Paragraph)
	if !ok {
		t.Fatal("第一个元素不是段落")
	}

	if paragraph.Properties == nil {
		t.Fatal("段落属性为空")
	}

	if paragraph.Properties.Justification == nil {
		t.Fatal("段落对齐属性为空")
	}

	if paragraph.Properties.Justification.Val != string(AlignCenter) {
		t.Errorf("段落对齐方式不正确，预期: %s，实际: %s",
			AlignCenter, paragraph.Properties.Justification.Val)
	}

	// 测试修改对齐方式
	err = doc.SetImageAlignment(imageInfo, AlignRight)
	if err != nil {
		t.Fatalf("修改图片对齐方式失败: %v", err)
	}

	if imageInfo.Config.Alignment != AlignRight {
		t.Error("图片对齐方式修改失败")
	}
}
