// Package document 提供Word文档的核心操作功能
package document

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ZeroHawkeye/wordZero/pkg/style"
)

// Document 表示一个Word文档
type Document struct {
	// 文档的主要内容
	Body *Body
	// 文档关系
	relationships *Relationships
	// 文档级关系（用于页眉页脚等）
	documentRelationships *Relationships
	// 内容类型
	contentTypes *ContentTypes
	// 样式管理器
	styleManager *style.StyleManager
	// 临时存储文档部件
	parts map[string][]byte
	// 图片ID计数器，确保每个图片都有唯一的ID
	nextImageID int
}

// Body 表示文档主体
type Body struct {
	XMLName  xml.Name      `xml:"w:body"`
	Elements []interface{} `xml:"-"` // 不序列化此字段，使用自定义方法
}

// MarshalXML 自定义XML序列化，按照元素顺序输出
func (b *Body) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	// 开始元素
	if err := e.EncodeToken(start); err != nil {
		return err
	}

	// 分离SectionProperties和其他元素
	var sectPr *SectionProperties
	var otherElements []interface{}

	for _, element := range b.Elements {
		if sp, ok := element.(*SectionProperties); ok {
			sectPr = sp // 保存最后一个SectionProperties
		} else {
			otherElements = append(otherElements, element)
		}
	}

	// 先序列化其他元素（段落、表格等）
	for _, element := range otherElements {
		if err := e.Encode(element); err != nil {
			return err
		}
	}

	// 最后序列化SectionProperties（如果存在）
	if sectPr != nil {
		if err := e.Encode(sectPr); err != nil {
			return err
		}
	}

	// 结束元素
	return e.EncodeToken(start.End())
}

// BodyElement 文档主体元素接口
type BodyElement interface {
	ElementType() string
}

// ElementType 返回段落元素类型
func (p *Paragraph) ElementType() string {
	return "paragraph"
}

// ElementType 返回表格元素类型
func (t *Table) ElementType() string {
	return "table"
}

// Paragraph 表示一个段落
type Paragraph struct {
	XMLName    xml.Name             `xml:"w:p"`
	Properties *ParagraphProperties `xml:"w:pPr,omitempty"`
	Runs       []Run                `xml:"w:r"`
}

// ParagraphProperties 段落属性
type ParagraphProperties struct {
	XMLName             xml.Name             `xml:"w:pPr"`
	ParagraphStyle      *ParagraphStyle      `xml:"w:pStyle,omitempty"`
	NumberingProperties *NumberingProperties `xml:"w:numPr,omitempty"`
	ParagraphBorder     *ParagraphBorder     `xml:"w:pBdr,omitempty"`
	Tabs                *Tabs                `xml:"w:tabs,omitempty"`
	Spacing             *Spacing             `xml:"w:spacing,omitempty"`
	Indentation         *Indentation         `xml:"w:ind,omitempty"`
	Justification       *Justification       `xml:"w:jc,omitempty"`
	KeepNext            *KeepNext            `xml:"w:keepNext,omitempty"`        // 与下一段落保持在一起
	KeepLines           *KeepLines           `xml:"w:keepLines,omitempty"`       // 段落中的行保持在一起
	PageBreakBefore     *PageBreakBefore     `xml:"w:pageBreakBefore,omitempty"` // 段前分页
	WidowControl        *WidowControl        `xml:"w:widowControl,omitempty"`    // 孤行控制
	OutlineLevel        *OutlineLevel        `xml:"w:outlineLvl,omitempty"`      // 大纲级别
}

// ParagraphBorder 段落边框
type ParagraphBorder struct {
	XMLName xml.Name             `xml:"w:pBdr"`
	Top     *ParagraphBorderLine `xml:"w:top,omitempty"`
	Left    *ParagraphBorderLine `xml:"w:left,omitempty"`
	Bottom  *ParagraphBorderLine `xml:"w:bottom,omitempty"`
	Right   *ParagraphBorderLine `xml:"w:right,omitempty"`
}

// ParagraphBorderLine 段落边框线
type ParagraphBorderLine struct {
	Val   string `xml:"w:val,attr"`
	Color string `xml:"w:color,attr"`
	Sz    string `xml:"w:sz,attr"`
	Space string `xml:"w:space,attr"`
}

// Spacing 间距设置
type Spacing struct {
	XMLName  xml.Name `xml:"w:spacing"`
	Before   string   `xml:"w:before,attr,omitempty"`
	After    string   `xml:"w:after,attr,omitempty"`
	Line     string   `xml:"w:line,attr,omitempty"`
	LineRule string   `xml:"w:lineRule,attr,omitempty"`
}

// Justification 对齐方式
type Justification struct {
	XMLName xml.Name `xml:"w:jc"`
	Val     string   `xml:"w:val,attr,omitempty"`
}

// KeepNext 与下一段落保持在一起
type KeepNext struct {
	XMLName xml.Name `xml:"w:keepNext"`
	Val     string   `xml:"w:val,attr,omitempty"`
}

// KeepLines 段落中的行保持在一起
type KeepLines struct {
	XMLName xml.Name `xml:"w:keepLines"`
	Val     string   `xml:"w:val,attr,omitempty"`
}

// PageBreakBefore 段前分页
type PageBreakBefore struct {
	XMLName xml.Name `xml:"w:pageBreakBefore"`
	Val     string   `xml:"w:val,attr,omitempty"`
}

// WidowControl 孤行控制
type WidowControl struct {
	XMLName xml.Name `xml:"w:widowControl"`
	Val     string   `xml:"w:val,attr,omitempty"`
}

// OutlineLevel 大纲级别
type OutlineLevel struct {
	XMLName xml.Name `xml:"w:outlineLvl"`
	Val     string   `xml:"w:val,attr"`
}

// Run 表示一段文本
type Run struct {
	XMLName    xml.Name        `xml:"w:r"`
	Properties *RunProperties  `xml:"w:rPr,omitempty"`
	Text       Text            `xml:"w:t,omitempty"`
	Break      *Break          `xml:"w:br,omitempty"` // 分页符 / Page break
	Drawing    *DrawingElement `xml:"w:drawing,omitempty"`
	FieldChar  *FieldChar      `xml:"w:fldChar,omitempty"`
	InstrText  *InstrText      `xml:"w:instrText,omitempty"`
}

// RunProperties 文本属性
// 注意：字段顺序必须符合OpenXML标准，w:rFonts必须在w:color之前
type RunProperties struct {
	XMLName    xml.Name    `xml:"w:rPr"`
	FontFamily *FontFamily `xml:"w:rFonts,omitempty"`
	Bold       *Bold       `xml:"w:b,omitempty"`
	BoldCs     *BoldCs     `xml:"w:bCs,omitempty"`
	Italic     *Italic     `xml:"w:i,omitempty"`
	ItalicCs   *ItalicCs   `xml:"w:iCs,omitempty"`
	Underline  *Underline  `xml:"w:u,omitempty"`
	Strike     *Strike     `xml:"w:strike,omitempty"`
	Color      *Color      `xml:"w:color,omitempty"`
	FontSize   *FontSize   `xml:"w:sz,omitempty"`
	FontSizeCs *FontSizeCs `xml:"w:szCs,omitempty"`
	Highlight  *Highlight  `xml:"w:highlight,omitempty"`
}

// Bold 粗体
type Bold struct {
	XMLName xml.Name `xml:"w:b"`
}

// BoldCs 复杂脚本粗体
type BoldCs struct {
	XMLName xml.Name `xml:"w:bCs"`
}

// Italic 斜体
type Italic struct {
	XMLName xml.Name `xml:"w:i"`
}

// ItalicCs 复杂脚本斜体
type ItalicCs struct {
	XMLName xml.Name `xml:"w:iCs"`
}

// Underline 下划线
type Underline struct {
	XMLName xml.Name `xml:"w:u"`
	Val     string   `xml:"w:val,attr,omitempty"`
}

// Strike 删除线
type Strike struct {
	XMLName xml.Name `xml:"w:strike"`
}

// FontSize 字体大小
type FontSize struct {
	XMLName xml.Name `xml:"w:sz"`
	Val     string   `xml:"w:val,attr"`
}

// FontSizeCs 复杂脚本字体大小
type FontSizeCs struct {
	XMLName xml.Name `xml:"w:szCs"`
	Val     string   `xml:"w:val,attr"`
}

// Color 颜色
type Color struct {
	XMLName xml.Name `xml:"w:color"`
	Val     string   `xml:"w:val,attr"`
}

// Highlight 背景色
type Highlight struct {
	XMLName xml.Name `xml:"w:highlight"`
	Val     string   `xml:"w:val,attr"`
}

// Text 文本内容
type Text struct {
	XMLName xml.Name `xml:"w:t"`
	Space   string   `xml:"xml:space,attr,omitempty"`
	Content string   `xml:",chardata"`
}

// Break 分页符
// Break represents page breaks in Word documents
type Break struct {
	XMLName xml.Name `xml:"w:br"`
	Type    string   `xml:"w:type,attr,omitempty"` // "page" 表示分页符 / "page" indicates a page break
}

// Relationships 文档关系
type Relationships struct {
	XMLName       xml.Name       `xml:"Relationships"`
	Xmlns         string         `xml:"xmlns,attr"`
	Relationships []Relationship `xml:"Relationship"`
}

// Relationship 单个关系
type Relationship struct {
	ID     string `xml:"Id,attr"`
	Type   string `xml:"Type,attr"`
	Target string `xml:"Target,attr"`
}

// ContentTypes 内容类型
type ContentTypes struct {
	XMLName   xml.Name   `xml:"Types"`
	Xmlns     string     `xml:"xmlns,attr"`
	Defaults  []Default  `xml:"Default"`
	Overrides []Override `xml:"Override"`
}

// Default 默认内容类型
type Default struct {
	Extension   string `xml:"Extension,attr"`
	ContentType string `xml:"ContentType,attr"`
}

// Override 覆盖内容类型
type Override struct {
	PartName    string `xml:"PartName,attr"`
	ContentType string `xml:"ContentType,attr"`
}

// FontFamily 字体族
type FontFamily struct {
	XMLName  xml.Name `xml:"w:rFonts"`
	ASCII    string   `xml:"w:ascii,attr,omitempty"`
	HAnsi    string   `xml:"w:hAnsi,attr,omitempty"`
	EastAsia string   `xml:"w:eastAsia,attr,omitempty"`
	CS       string   `xml:"w:cs,attr,omitempty"`
	Hint     string   `xml:"w:hint,attr,omitempty"`
}

// TextFormat 文本格式配置
type TextFormat struct {
	Bold       bool   // 是否粗体
	Italic     bool   // 是否斜体
	FontSize   int    // 字体大小（磅）
	FontColor  string // 字体颜色（十六进制，如 "FF0000" 表示红色）
	FontFamily string // 字体名称（首选字段）
	FontName   string // 字体名称别名（为兼容早期文档示例/README 中使用的 FontName）
	Underline  bool   // 是否下划线
	Strike     bool   // 删除线
	Highlight  string //高亮颜色
}

// AlignmentType 对齐类型
type AlignmentType string

const (
	// AlignLeft 左对齐
	AlignLeft AlignmentType = "left"
	// AlignCenter 居中对齐
	AlignCenter AlignmentType = "center"
	// AlignRight 右对齐
	AlignRight AlignmentType = "right"
	// AlignJustify 两端对齐
	AlignJustify AlignmentType = "both"
)

// SpacingConfig 间距配置
type SpacingConfig struct {
	LineSpacing     float64 // 行间距（倍数，如1.5表示1.5倍行距）
	BeforePara      int     // 段前间距（磅）
	AfterPara       int     // 段后间距（磅）
	FirstLineIndent int     // 首行缩进（磅）
}

// Indentation 缩进设置
type Indentation struct {
	XMLName   xml.Name `xml:"w:ind"`
	FirstLine string   `xml:"w:firstLine,attr,omitempty"`
	Left      string   `xml:"w:left,attr,omitempty"`
	Right     string   `xml:"w:right,attr,omitempty"`
}

// Tabs 制表符设置
type Tabs struct {
	XMLName xml.Name `xml:"w:tabs"`
	Tabs    []TabDef `xml:"w:tab"`
}

// TabDef 制表符定义
type TabDef struct {
	XMLName xml.Name `xml:"w:tab"`
	Val     string   `xml:"w:val,attr"`
	Leader  string   `xml:"w:leader,attr,omitempty"`
	Pos     string   `xml:"w:pos,attr"`
}

// ParagraphStyle 段落样式引用
type ParagraphStyle struct {
	XMLName xml.Name `xml:"w:pStyle"`
	Val     string   `xml:"w:val,attr"`
}

// NumberingProperties 段落编号属性
type NumberingProperties struct {
	XMLName xml.Name `xml:"w:numPr"`
	ILevel  *ILevel  `xml:"w:ilvl,omitempty"`
	NumID   *NumID   `xml:"w:numId,omitempty"`
}

// ILevel 编号级别
type ILevel struct {
	XMLName xml.Name `xml:"w:ilvl"`
	Val     string   `xml:"w:val,attr"`
}

// NumID 编号ID
type NumID struct {
	XMLName xml.Name `xml:"w:numId"`
	Val     string   `xml:"w:val,attr"`
}

// New 创建一个新的Word文档
func New() *Document {
	Debugf("创建新文档")

	doc := &Document{
		Body: &Body{
			Elements: make([]interface{}, 0),
		},
		styleManager: style.NewStyleManager(),
		parts:        make(map[string][]byte),
		nextImageID:  0, // 初始化图片ID计数器，从0开始
		documentRelationships: &Relationships{
			Xmlns:         "http://schemas.openxmlformats.org/package/2006/relationships",
			Relationships: []Relationship{},
		},
	}

	// 初始化文档结构
	doc.initializeStructure()

	return doc
}

// Open 打开一个现有的Word文档。
//
// 参数 filename 是要打开的 .docx 文件路径。
// 该函数会解析整个文档结构，包括文本内容、格式和属性。
//
// 如果文件不存在、格式错误或解析失败，会返回相应的错误。
//
// 示例:
//
//	doc, err := document.Open("existing.docx")
//	if err != nil {
//		log.Fatal(err)
//	}
//
//	// 打印所有段落内容
//	for i, para := range doc.Body.Paragraphs {
//		fmt.Printf("段落 %d: ", i+1)
//		for _, run := range para.Runs {
//			fmt.Print(run.Text.Content)
//		}
//		fmt.Println()
//	}
func Open(filename string) (*Document, error) {
	Infof("正在打开文档: %s", filename)

	reader, err := zip.OpenReader(filename)
	if err != nil {
		Errorf("无法打开文件: %s", filename)
		return nil, WrapErrorWithContext("open_file", err, filename)
	}
	defer reader.Close()

	doc := &Document{
		parts: make(map[string][]byte),
		documentRelationships: &Relationships{
			Xmlns:         "http://schemas.openxmlformats.org/package/2006/relationships",
			Relationships: []Relationship{},
		},
		nextImageID: 0, // 初始化图片ID计数器，从0开始
	}

	// 读取所有文件部件
	for _, file := range reader.File {
		rc, err := file.Open()
		if err != nil {
			Errorf("无法打开文件部件: %s", file.Name)
			return nil, WrapErrorWithContext("open_part", err, file.Name)
		}

		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			Errorf("无法读取文件部件: %s", file.Name)
			return nil, WrapErrorWithContext("read_part", err, file.Name)
		}

		doc.parts[file.Name] = data
		Debugf("已读取文件部件: %s (%d 字节)", file.Name, len(data))
	}

	// 初始化样式管理器
	doc.styleManager = style.NewStyleManager()

	// 解析内容类型
	if err := doc.parseContentTypes(); err != nil {
		Debugf("解析内容类型失败，使用默认值: %v", err)
		// 如果解析失败，使用默认值
		doc.contentTypes = &ContentTypes{
			Xmlns: "http://schemas.openxmlformats.org/package/2006/content-types",
			Defaults: []Default{
				{Extension: "rels", ContentType: "application/vnd.openxmlformats-package.relationships+xml"},
				{Extension: "xml", ContentType: "application/xml"},
			},
			Overrides: []Override{
				{PartName: "/word/document.xml", ContentType: "application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"},
				{PartName: "/word/styles.xml", ContentType: "application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"},
			},
		}
	}

	// 解析关系
	if err := doc.parseRelationships(); err != nil {
		Debugf("解析关系失败，使用默认值: %v", err)
		// 如果解析失败，使用默认值
		doc.relationships = &Relationships{
			Xmlns: "http://schemas.openxmlformats.org/package/2006/relationships",
			Relationships: []Relationship{
				{
					ID:     "rId1",
					Type:   "http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument",
					Target: "word/document.xml",
				},
			},
		}
	}

	// 解析主文档
	if err := doc.parseDocument(); err != nil {
		Errorf("解析文档失败: %s", filename)
		return nil, WrapErrorWithContext("parse_document", err, filename)
	}

	// 解析样式文件
	if err := doc.parseStyles(); err != nil {
		Debugf("解析样式失败，使用默认样式: %v", err)
		// 如果样式解析失败，重新初始化为默认样式
		doc.styleManager = style.NewStyleManager()
	}

	// 解析文档关系（包括图片等资源的关系）
	if err := doc.parseDocumentRelationships(); err != nil {
		Debugf("解析文档关系失败，使用默认值: %v", err)
		// 如果解析失败，保持初始化的空关系列表
	}

	// 根据已有的图片关系更新nextImageID计数器
	doc.updateNextImageID()

	Infof("成功打开文档: %s", filename)
	return doc, nil
}

func OpenFromMemory(readCloser io.ReadCloser) (*Document, error) {
	defer readCloser.Close()
	Infof("正在打开文档")

	fileData, err := io.ReadAll(readCloser)
	if err != nil {
		return nil, fmt.Errorf("读取文件内容失败: %w", err)
	}
	defer readCloser.Close()

	readerAt := bytes.NewReader(fileData)

	reader, err := zip.NewReader(readerAt, readerAt.Size())
	if err != nil {
		Errorf("无法打开文件")
		return nil, WrapErrorWithContext("open_file", err, "")
	}

	doc := &Document{
		parts: make(map[string][]byte),
		documentRelationships: &Relationships{
			Xmlns:         "http://schemas.openxmlformats.org/package/2006/relationships",
			Relationships: []Relationship{},
		},
		nextImageID: 0, // 初始化图片ID计数器，从0开始
	}

	// 读取所有文件部件
	for _, file := range reader.File {
		rc, err := file.Open()
		if err != nil {
			Errorf("无法打开文件部件: %s", file.Name)
			return nil, WrapErrorWithContext("open_part", err, file.Name)
		}

		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			Errorf("无法读取文件部件: %s", file.Name)
			return nil, WrapErrorWithContext("read_part", err, file.Name)
		}

		doc.parts[file.Name] = data
		Debugf("已读取文件部件: %s (%d 字节)", file.Name, len(data))
	}

	// 初始化样式管理器
	doc.styleManager = style.NewStyleManager()

	// 解析主文档
	if err := doc.parseDocument(); err != nil {
		Errorf("解析文档失败")
		return nil, WrapErrorWithContext("parse_document", err, "")
	}

	// 解析样式文件
	if err := doc.parseStyles(); err != nil {
		Debugf("解析样式失败，使用默认样式: %v", err)
		// 如果样式解析失败，重新初始化为默认样式
		doc.styleManager = style.NewStyleManager()
	}

	// 解析文档关系（包括图片等资源的关系）
	if err := doc.parseDocumentRelationships(); err != nil {
		Debugf("解析文档关系失败，使用默认值: %v", err)
		// 如果解析失败，保持初始化的空关系列表
	}

	// 根据已有的图片关系更新nextImageID计数器
	doc.updateNextImageID()

	Infof("成功打开文档")
	return doc, nil
}

// Save 将文档保存到指定的文件路径。
//
// 参数 filename 是保存文件的路径，包含文件名和扩展名。
// 如果目录不存在，会自动创建所需的目录结构。
//
// 保存过程包括序列化所有文档内容、压缩为ZIP格式，
// 并写入到文件系统。
//
// 示例:
//
//	doc := document.New()
//	doc.AddParagraph("示例内容")
//
//	// 保存到当前目录
//	err := doc.Save("example.docx")
//
//	// 保存到子目录（会自动创建目录）
//	err = doc.Save("output/documents/example.docx")
//
//	if err != nil {
//		log.Fatal(err)
//	}
func (d *Document) Save(filename string) error {
	Infof("正在保存文档: %s", filename)

	// 确保目录存在
	dir := filepath.Dir(filename)
	if err := os.MkdirAll(dir, 0755); err != nil {
		Errorf("无法创建目录: %s", dir)
		return WrapErrorWithContext("create_dir", err, dir)
	}

	// 创建文件
	file, err := os.Create(filename)
	if err != nil {
		Errorf("无法创建文件: %s", filename)
		return WrapErrorWithContext("create_file", err, filename)
	}
	defer file.Close()

	// 创建ZIP写入器
	zipWriter := zip.NewWriter(file)
	defer zipWriter.Close()

	// 序列化主文档
	if err := d.serializeDocument(); err != nil {
		Errorf("序列化文档失败")
		return WrapError("serialize_document", err)
	}

	// 序列化样式
	if err := d.serializeStyles(); err != nil {
		Errorf("序列化样式失败")
		return WrapError("serialize_styles", err)
	}

	// 序列化内容类型
	d.serializeContentTypes()

	// 序列化关系
	d.serializeRelationships()

	// 序列化文档关系
	d.serializeDocumentRelationships()

	// 写入所有部件
	for name, data := range d.parts {
		writer, err := zipWriter.Create(name)
		if err != nil {
			Errorf("无法创建ZIP条目: %s", name)
			return WrapErrorWithContext("create_zip_entry", err, name)
		}

		if _, err := writer.Write(data); err != nil {
			Errorf("无法写入ZIP条目: %s", name)
			return WrapErrorWithContext("write_zip_entry", err, name)
		}

		Debugf("已写入ZIP条目: %s (%d 字节)", name, len(data))
	}

	Infof("成功保存文档: %s", filename)
	return nil
}

// AddParagraph 向文档添加一个普通段落。
//
// 参数 text 是段落的文本内容。段落会使用默认格式，
// 可以后续通过返回的 Paragraph 指针设置格式和属性。
//
// 返回新创建段落的指针，可用于进一步格式化。
//
// 示例:
//
//	doc := document.New()
//
//	// 添加普通段落
//	para := doc.AddParagraph("这是一个段落")
//
//	// 设置段落属性
//	para.SetAlignment(document.AlignCenter)
//	para.SetSpacing(&document.SpacingConfig{
//		LineSpacing: 1.5,
//		BeforePara:  12,
//	})
func (d *Document) AddParagraph(text string) *Paragraph {
	Debugf("添加段落: %s", text)
	p := &Paragraph{
		Runs: []Run{
			{
				Text: Text{
					Content: text,
					Space:   "preserve",
				},
			},
		},
	}

	d.Body.Elements = append(d.Body.Elements, p)
	return p
}

// AddFormattedParagraph 向文档添加一个格式化段落。
//
// 参数 text 是段落的文本内容。
// 参数 format 指定文本格式，如果为 nil 则使用默认格式。
//
// 返回新创建段落的指针，可用于进一步设置段落属性。
//
// 示例:
//
//	doc := document.New()
//
//	// 创建格式配置
//	titleFormat := &document.TextFormat{
//		Bold:      true,
//		FontSize:  18,
//		FontColor: "FF0000", // 红色
//		FontName:  "微软雅黑",
//	}
//
//	// 添加格式化标题
//	title := doc.AddFormattedParagraph("文档标题", titleFormat)
//	title.SetAlignment(document.AlignCenter)
func (d *Document) AddFormattedParagraph(text string, format *TextFormat) *Paragraph {
	Debugf("添加格式化段落: %s", text)

	// 创建运行属性
	runProps := &RunProperties{}

	if format != nil {
		// 兼容 FontFamily 与 FontName 两个字段
		fontName := ""
		if format.FontFamily != "" {
			fontName = format.FontFamily
		} else if format.FontName != "" { // 向后兼容示例代码
			fontName = format.FontName
		}
		if fontName != "" {
			runProps.FontFamily = &FontFamily{ // 设置所有相关字段，保证测试与渲染一致
				ASCII:    fontName,
				HAnsi:    fontName,
				EastAsia: fontName,
				CS:       fontName,
			}
		}

		if format.Bold {
			runProps.Bold = &Bold{}
		}

		if format.Italic {
			runProps.Italic = &Italic{}
		}

		if format.FontColor != "" {
			// 确保颜色格式正确（移除#前缀）
			color := strings.TrimPrefix(format.FontColor, "#")
			runProps.Color = &Color{Val: color}
		}

		if format.FontSize > 0 {
			// Word中字体大小是半磅为单位，所以需要乘以2
			runProps.FontSize = &FontSize{Val: strconv.Itoa(format.FontSize * 2)}
		}
		if format.Underline {
			runProps.Underline = &Underline{Val: "single"} // 默认单线下划线
		}

		if format.Strike {
			runProps.Strike = &Strike{} // 添加删除线
		}

		if format.Highlight != "" {
			runProps.Highlight = &Highlight{Val: format.Highlight}
		}
	}

	p := &Paragraph{
		Runs: []Run{
			{
				Properties: runProps,
				Text: Text{
					Content: text,
					Space:   "preserve",
				},
			},
		},
	}

	d.Body.Elements = append(d.Body.Elements, p)
	return p
}

// SetAlignment 设置段落的对齐方式。
//
// 参数 alignment 指定对齐类型，支持以下值：
//   - AlignLeft: 左对齐（默认）
//   - AlignCenter: 居中对齐
//   - AlignRight: 右对齐
//   - AlignJustify: 两端对齐
//
// 示例:
//
//	para := doc.AddParagraph("居中标题")
//	para.SetAlignment(document.AlignCenter)
//
//	para2 := doc.AddParagraph("右对齐文本")
//	para2.SetAlignment(document.AlignRight)
func (p *Paragraph) SetAlignment(alignment AlignmentType) {
	if p.Properties == nil {
		p.Properties = &ParagraphProperties{}
	}

	p.Properties.Justification = &Justification{Val: string(alignment)}
	Debugf("设置段落对齐方式: %s", alignment)
}

// SetSpacing 设置段落的间距配置。
//
// 参数 config 包含各种间距设置，如果为 nil 则不进行任何设置。
// 配置选项包括：
//   - LineSpacing: 行间距倍数（如 1.5 表示1.5倍行距）
//   - BeforePara: 段前间距（磅）
//   - AfterPara: 段后间距（磅）
//   - FirstLineIndent: 首行缩进（磅）
//
// 注意：间距值会自动转换为 Word 内部使用的 TWIPs 单位（1磅=20TWIPs）。
//
// 示例:
//
//	para := doc.AddParagraph("带间距的段落")
//
//	// 设置复杂间距
//	para.SetSpacing(&document.SpacingConfig{
//		LineSpacing:     1.5, // 1.5倍行距
//		BeforePara:      12,  // 段前12磅
//		AfterPara:       6,   // 段后6磅
//		FirstLineIndent: 24,  // 首行缩进24磅
//	})
//
//	// 只设置行间距
//	para2 := doc.AddParagraph("双倍行距")
//	para2.SetSpacing(&document.SpacingConfig{
//		LineSpacing: 2.0,
//	})
func (p *Paragraph) SetSpacing(config *SpacingConfig) {
	if p.Properties == nil {
		p.Properties = &ParagraphProperties{}
	}

	if config != nil {
		spacing := &Spacing{}

		if config.BeforePara > 0 {
			// 转换为TWIPs (1/20磅)
			spacing.Before = strconv.Itoa(config.BeforePara * 20)
		}

		if config.AfterPara > 0 {
			// 转换为TWIPs (1/20磅)
			spacing.After = strconv.Itoa(config.AfterPara * 20)
		}

		if config.LineSpacing > 0 {
			// 行间距，240表示单倍行距
			spacing.Line = strconv.Itoa(int(config.LineSpacing * 240))
		}

		p.Properties.Spacing = spacing

		if config.FirstLineIndent > 0 {
			if p.Properties.Indentation == nil {
				p.Properties.Indentation = &Indentation{}
			}
			// 转换为TWIPs (1/20磅)
			p.Properties.Indentation.FirstLine = strconv.Itoa(config.FirstLineIndent * 20)
		}

		Debugf("设置段落间距: 段前=%d, 段后=%d, 行距=%.1f, 首行缩进=%d",
			config.BeforePara, config.AfterPara, config.LineSpacing, config.FirstLineIndent)
	}
}

// AddFormattedText 向段落添加格式化的文本内容。
//
// 此方法允许在一个段落中混合使用不同格式的文本。
// 新的文本会作为一个新的 Run 添加到段落中。
//
// 参数 text 是要添加的文本内容。
// 参数 format 指定文本格式，如果为 nil 则使用默认格式。
//
// 示例:
//
//	para := doc.AddParagraph("这个段落包含")
//
//	// 添加粗体红色文本
//	para.AddFormattedText("粗体红色", &document.TextFormat{
//		Bold: true,
//		FontColor: "FF0000",
//	})
//
//	// 添加普通文本
//	para.AddFormattedText("和普通文本", nil)
//
//	// 添加斜体蓝色文本
//	para.AddFormattedText("以及斜体蓝色", &document.TextFormat{
//		Italic: true,
//		FontColor: "0000FF",
//		FontSize: 14,
//	})
func (p *Paragraph) AddFormattedText(text string, format *TextFormat) {
	// 创建运行属性
	runProps := &RunProperties{}

	if format != nil {
		fontName := ""
		if format.FontFamily != "" {
			fontName = format.FontFamily
		} else if format.FontName != "" { // 兼容旧示例
			fontName = format.FontName
		}
		if fontName != "" {
			runProps.FontFamily = &FontFamily{
				ASCII:    fontName,
				HAnsi:    fontName,
				EastAsia: fontName,
				CS:       fontName,
			}
		}

		if format.Bold {
			runProps.Bold = &Bold{}
		}

		if format.Italic {
			runProps.Italic = &Italic{}
		}

		if format.FontColor != "" {
			color := strings.TrimPrefix(format.FontColor, "#")
			runProps.Color = &Color{Val: color}
		}

		if format.FontSize > 0 {
			runProps.FontSize = &FontSize{Val: strconv.Itoa(format.FontSize * 2)}
		}
		if format.Underline {
			runProps.Underline = &Underline{Val: "single"} // 默认单线下划线
		}

		if format.Strike {
			runProps.Strike = &Strike{} // 添加删除线
		}

		if format.Highlight != "" {
			runProps.Highlight = &Highlight{Val: format.Highlight}
		}
	}

	run := Run{
		Properties: runProps,
		Text: Text{
			Content: text,
			Space:   "preserve",
		},
	}

	p.Runs = append(p.Runs, run)
	Debugf("向段落添加格式化文本: %s", text)
}

// AddHeadingParagraph 向文档添加一个标题段落。
//
// 参数 text 是标题的文本内容。
// 参数 level 是标题级别（1-9），对应 Heading1 到 Heading9。
//
// 返回新创建段落的指针，可用于进一步设置段落属性。
// 此方法会自动设置正确的样式引用，确保标题能被 Word 导航窗格识别。
//
// 示例:
//
//	doc := document.New()
//
//	// 添加一级标题
//	h1 := doc.AddHeadingParagraph("第一章：概述", 1)
//
//	// 添加二级标题
//	h2 := doc.AddHeadingParagraph("1.1 背景", 2)
//
//	// 添加三级标题
//	h3 := doc.AddHeadingParagraph("1.1.1 研究目标", 3)
func (d *Document) AddHeadingParagraph(text string, level int) *Paragraph {
	return d.AddHeadingParagraphWithBookmark(text, level, "")
}

// AddHeadingParagraphWithBookmark 向文档添加一个带书签的标题段落。
//
// 参数 text 是标题的文本内容。
// 参数 level 是标题级别（1-9），对应 Heading1 到 Heading9。
// 参数 bookmarkName 是书签名称，如果为空字符串则不添加书签。
//
// 返回新创建段落的指针，可用于进一步设置段落属性。
// 此方法会自动设置正确的样式引用，确保标题能被 Word 导航窗格识别，
// 并在需要时添加书签以支持目录导航和超链接。
//
// 示例:
//
//	doc := document.New()
//
//	// 添加带书签的一级标题
//	h1 := doc.AddHeadingParagraphWithBookmark("第一章：概述", 1, "chapter1")
//
//	// 添加不带书签的二级标题
//	h2 := doc.AddHeadingParagraphWithBookmark("1.1 背景", 2, "")
//
//	// 添加自动生成书签名的三级标题
//	h3 := doc.AddHeadingParagraphWithBookmark("1.1.1 研究目标", 3, "auto_bookmark")
func (d *Document) AddHeadingParagraphWithBookmark(text string, level int, bookmarkName string) *Paragraph {
	if level < 1 || level > 9 {
		Debugf("标题级别 %d 超出范围，使用默认级别 1", level)
		level = 1
	}

	styleID := fmt.Sprintf("Heading%d", level)
	Debugf("添加标题段落: %s (级别: %d, 样式: %s, 书签: %s)", text, level, styleID, bookmarkName)

	// 获取样式管理器中的样式
	headingStyle := d.styleManager.GetStyle(styleID)
	if headingStyle == nil {
		Debugf("警告：找不到样式 %s，使用默认样式", styleID)
		return d.AddParagraph(text)
	}

	// 创建运行属性，应用样式中的字符格式
	runProps := &RunProperties{}
	if headingStyle.RunPr != nil {
		if headingStyle.RunPr.Bold != nil {
			runProps.Bold = &Bold{}
		}
		if headingStyle.RunPr.Italic != nil {
			runProps.Italic = &Italic{}
		}
		if headingStyle.RunPr.FontSize != nil {
			runProps.FontSize = &FontSize{Val: headingStyle.RunPr.FontSize.Val}
		}
		if headingStyle.RunPr.Color != nil {
			runProps.Color = &Color{Val: headingStyle.RunPr.Color.Val}
		}
		if headingStyle.RunPr.FontFamily != nil {
			runProps.FontFamily = &FontFamily{ASCII: headingStyle.RunPr.FontFamily.ASCII}
		}
	}

	// 创建段落属性，应用样式中的段落格式
	paraProps := &ParagraphProperties{
		ParagraphStyle: &ParagraphStyle{Val: styleID},
	}

	// 应用样式中的段落格式
	if headingStyle.ParagraphPr != nil {
		if headingStyle.ParagraphPr.Spacing != nil {
			paraProps.Spacing = &Spacing{
				Before: headingStyle.ParagraphPr.Spacing.Before,
				After:  headingStyle.ParagraphPr.Spacing.After,
				Line:   headingStyle.ParagraphPr.Spacing.Line,
			}
		}
		if headingStyle.ParagraphPr.Justification != nil {
			paraProps.Justification = &Justification{
				Val: headingStyle.ParagraphPr.Justification.Val,
			}
		}
		if headingStyle.ParagraphPr.Indentation != nil {
			paraProps.Indentation = &Indentation{
				FirstLine: headingStyle.ParagraphPr.Indentation.FirstLine,
				Left:      headingStyle.ParagraphPr.Indentation.Left,
				Right:     headingStyle.ParagraphPr.Indentation.Right,
			}
		}
	}

	// 创建段落的Run列表
	runs := make([]Run, 0)

	// 如果需要添加书签，在段落开始处添加书签开始标记
	if bookmarkName != "" {
		// 生成唯一的书签ID
		bookmarkID := fmt.Sprintf("bookmark_%d_%s", len(d.Body.Elements), bookmarkName)

		// 添加书签开始标记作为单独的元素到文档主体中
		d.Body.Elements = append(d.Body.Elements, &BookmarkStart{
			ID:   bookmarkID,
			Name: bookmarkName,
		})

		Debugf("添加书签开始: ID=%s, Name=%s", bookmarkID, bookmarkName)
	}

	// 添加文本内容
	runs = append(runs, Run{
		Properties: runProps,
		Text: Text{
			Content: text,
			Space:   "preserve",
		},
	})

	// 创建段落
	p := &Paragraph{
		Properties: paraProps,
		Runs:       runs,
	}

	d.Body.Elements = append(d.Body.Elements, p)

	// 如果需要添加书签，在段落结束后添加书签结束标记
	if bookmarkName != "" {
		bookmarkID := fmt.Sprintf("bookmark_%d_%s", len(d.Body.Elements)-2, bookmarkName) // -2 因为段落已经添加了

		// 添加书签结束标记
		d.Body.Elements = append(d.Body.Elements, &BookmarkEnd{
			ID: bookmarkID,
		})

		Debugf("添加书签结束: ID=%s", bookmarkID)
	}

	return p
}

// AddPageBreak 向文档添加一个分页符。
//
// 分页符会强制在当前位置开始一个新页面。
// 此方法会创建一个包含分页符的段落。
//
// 示例:
//
//	doc := document.New()
//	doc.AddParagraph("第一页内容")
//	doc.AddPageBreak()
//	doc.AddParagraph("第二页内容")
func (d *Document) AddPageBreak() {
	Debugf("添加分页符")

	// 创建一个包含分页符的段落
	p := &Paragraph{
		Runs: []Run{
			{
				Break: &Break{
					Type: "page",
				},
			},
		},
	}

	d.Body.Elements = append(d.Body.Elements, p)
}

// SetStyle 设置段落的样式。
//
// 参数 styleID 是要应用的样式ID，如 "Heading1"、"Normal" 等。
// 此方法会设置段落的样式引用，确保段落使用指定的样式。
//
// 示例:
//
//	para := doc.AddParagraph("这是一个段落")
//	para.SetStyle("Heading2")  // 设置为二级标题样式
func (p *Paragraph) SetStyle(styleID string) {
	if p.Properties == nil {
		p.Properties = &ParagraphProperties{}
	}

	p.Properties.ParagraphStyle = &ParagraphStyle{Val: styleID}
	Debugf("设置段落样式: %s", styleID)
}

// SetIndentation 设置段落的缩进属性。
//
// 参数：
//   - firstLineCm: 首行缩进，单位为厘米（可以为负数表示悬挂缩进）
//   - leftCm: 左缩进，单位为厘米
//   - rightCm: 右缩进，单位为厘米
//
// 示例：
//
//	para := doc.AddParagraph("这是一个有缩进的段落")
//	para.SetIndentation(0.5, 0, 0)    // 首行缩进0.5厘米
//	para.SetIndentation(-0.5, 1, 0)  // 悬挂缩进0.5厘米，左缩进1厘米
func (p *Paragraph) SetIndentation(firstLineCm, leftCm, rightCm float64) {
	if p.Properties == nil {
		p.Properties = &ParagraphProperties{}
	}

	if p.Properties.Indentation == nil {
		p.Properties.Indentation = &Indentation{}
	}

	// 转换厘米为TWIPs (1厘米 = 567 TWIPs)
	if firstLineCm != 0 {
		p.Properties.Indentation.FirstLine = strconv.Itoa(int(firstLineCm * 567))
	}

	if leftCm != 0 {
		p.Properties.Indentation.Left = strconv.Itoa(int(leftCm * 567))
	}

	if rightCm != 0 {
		p.Properties.Indentation.Right = strconv.Itoa(int(rightCm * 567))
	}

	Debugf("设置段落缩进: 首行=%.2fcm, 左=%.2fcm, 右=%.2fcm", firstLineCm, leftCm, rightCm)
}

// SetKeepWithNext 设置段落与下一段落保持在同一页。
//
// 此方法用于确保当前段落和下一段落不会被分页符分隔，
// 常用于标题和正文的组合，或需要保持连续性的内容。
//
// 参数：
//   - keep: true表示启用该属性，false表示禁用
//
// 示例：
//
//	// 标题与下一段保持在一起
//	title := doc.AddParagraph("第一章 概述")
//	title.SetKeepWithNext(true)
//	doc.AddParagraph("本章介绍...")  // 这段内容会与标题保持在同一页
func (p *Paragraph) SetKeepWithNext(keep bool) {
	if p.Properties == nil {
		p.Properties = &ParagraphProperties{}
	}

	if keep {
		p.Properties.KeepNext = &KeepNext{Val: "1"}
		Debugf("设置段落与下一段保持在一起")
	} else {
		p.Properties.KeepNext = nil
		Debugf("取消段落与下一段保持在一起")
	}
}

// SetKeepLines 设置段落中的所有行保持在同一页。
//
// 此方法用于防止段落在分页时被拆分到多个页面，
// 确保段落的所有行都显示在同一页上。
//
// 参数：
//   - keep: true表示启用该属性，false表示禁用
//
// 示例：
//
//	// 确保整个段落不被分页
//	para := doc.AddParagraph("这是一个重要的段落，需要保持完整显示。")
//	para.SetKeepLines(true)
func (p *Paragraph) SetKeepLines(keep bool) {
	if p.Properties == nil {
		p.Properties = &ParagraphProperties{}
	}

	if keep {
		p.Properties.KeepLines = &KeepLines{Val: "1"}
		Debugf("设置段落行保持在一起")
	} else {
		p.Properties.KeepLines = nil
		Debugf("取消段落行保持在一起")
	}
}

// SetPageBreakBefore 设置段落前插入分页符。
//
// 此方法用于在段落之前强制插入分页符，使段落从新页开始显示。
// 常用于章节标题或需要单独成页的内容。
//
// 参数：
//   - pageBreak: true表示启用段前分页，false表示禁用
//
// 示例：
//
//	// 章节标题从新页开始
//	chapter := doc.AddParagraph("第二章 详细说明")
//	chapter.SetPageBreakBefore(true)
func (p *Paragraph) SetPageBreakBefore(pageBreak bool) {
	if p.Properties == nil {
		p.Properties = &ParagraphProperties{}
	}

	if pageBreak {
		p.Properties.PageBreakBefore = &PageBreakBefore{Val: "1"}
		Debugf("设置段前分页")
	} else {
		p.Properties.PageBreakBefore = nil
		Debugf("取消段前分页")
	}
}

// SetWidowControl 设置段落的孤行控制。
//
// 孤行控制用于防止段落的第一行或最后一行单独出现在页面底部或顶部，
// 提高文档的排版质量。
//
// 参数：
//   - control: true表示启用孤行控制（默认），false表示禁用
//
// 示例：
//
//	para := doc.AddParagraph("这是一个长段落...")
//	para.SetWidowControl(true)  // 启用孤行控制
func (p *Paragraph) SetWidowControl(control bool) {
	if p.Properties == nil {
		p.Properties = &ParagraphProperties{}
	}

	if control {
		p.Properties.WidowControl = &WidowControl{Val: "1"}
		Debugf("启用段落孤行控制")
	} else {
		p.Properties.WidowControl = &WidowControl{Val: "0"}
		Debugf("禁用段落孤行控制")
	}
}

// SetOutlineLevel 设置段落的大纲级别。
//
// 大纲级别用于在文档导航窗格中显示文档结构，级别范围为0-8。
// 通常用于标题段落，配合目录功能使用。
//
// 参数：
//   - level: 大纲级别，0-8之间的整数（0表示正文，1-8对应标题1-8）
//
// 示例：
//
//	// 设置为一级标题的大纲级别
//	title := doc.AddParagraph("第一章")
//	title.SetOutlineLevel(0)  // 对应Heading1
//
//	// 设置为二级标题的大纲级别
//	subtitle := doc.AddParagraph("1.1 概述")
//	subtitle.SetOutlineLevel(1)  // 对应Heading2
func (p *Paragraph) SetOutlineLevel(level int) {
	if p.Properties == nil {
		p.Properties = &ParagraphProperties{}
	}

	if level < 0 || level > 8 {
		Warnf("大纲级别应在0-8之间，已调整为有效范围")
		if level < 0 {
			level = 0
		} else {
			level = 8
		}
	}

	p.Properties.OutlineLevel = &OutlineLevel{Val: strconv.Itoa(level)}
	Debugf("设置段落大纲级别: %d", level)
}

// ParagraphFormatConfig 段落格式配置
//
// 此结构体提供了段落所有格式属性的统一配置接口，
// 允许一次性设置多个段落属性，提高代码的可读性和易用性。
type ParagraphFormatConfig struct {
	// 基础格式
	Alignment AlignmentType // 对齐方式（AlignLeft, AlignCenter, AlignRight, AlignJustify）
	Style     string        // 段落样式ID（如"Heading1", "Normal"等）

	// 间距设置
	LineSpacing     float64 // 行间距（倍数，如1.5表示1.5倍行距）
	BeforePara      int     // 段前间距（磅）
	AfterPara       int     // 段后间距（磅）
	FirstLineIndent int     // 首行缩进（磅）

	// 缩进设置
	FirstLineCm float64 // 首行缩进（厘米，可以为负数表示悬挂缩进）
	LeftCm      float64 // 左缩进（厘米）
	RightCm     float64 // 右缩进（厘米）

	// 分页与控制
	KeepWithNext    bool // 与下一段落保持在同一页
	KeepLines       bool // 段落中的所有行保持在同一页
	PageBreakBefore bool // 段前分页
	WidowControl    bool // 孤行控制

	// 大纲级别
	OutlineLevel int // 大纲级别（0-8，0表示正文，1-8对应标题1-8）
}

// SetParagraphFormat 使用配置一次性设置段落的所有格式属性。
//
// 此方法提供了一种便捷的方式来设置段落的所有格式属性，
// 而不需要调用多个单独的设置方法。只有非零值的属性会被应用。
//
// 参数：
//   - config: 段落格式配置，包含所有格式属性
//
// 示例：
//
//	// 创建一个带完整格式的段落
//	para := doc.AddParagraph("重要章节标题")
//	para.SetParagraphFormat(&document.ParagraphFormatConfig{
//		Alignment:       document.AlignCenter,
//		Style:           "Heading1",
//		LineSpacing:     1.5,
//		BeforePara:      24,
//		AfterPara:       12,
//		KeepWithNext:    true,
//		PageBreakBefore: true,
//		OutlineLevel:    0,
//	})
//
//	// 设置带缩进的正文段落
//	para2 := doc.AddParagraph("正文内容...")
//	para2.SetParagraphFormat(&document.ParagraphFormatConfig{
//		Alignment:       document.AlignJustify,
//		FirstLineCm:     0.5,
//		LineSpacing:     1.5,
//		BeforePara:      6,
//		AfterPara:       6,
//		WidowControl:    true,
//	})
func (p *Paragraph) SetParagraphFormat(config *ParagraphFormatConfig) {
	if config == nil {
		return
	}

	// 设置对齐方式
	if config.Alignment != "" {
		p.SetAlignment(config.Alignment)
	}

	// 设置样式
	if config.Style != "" {
		p.SetStyle(config.Style)
	}

	// 设置间距（如果有任何间距设置）
	if config.LineSpacing > 0 || config.BeforePara > 0 || config.AfterPara > 0 || config.FirstLineIndent > 0 {
		p.SetSpacing(&SpacingConfig{
			LineSpacing:     config.LineSpacing,
			BeforePara:      config.BeforePara,
			AfterPara:       config.AfterPara,
			FirstLineIndent: config.FirstLineIndent,
		})
	}

	// 设置缩进（如果有任何缩进设置）
	if config.FirstLineCm != 0 || config.LeftCm != 0 || config.RightCm != 0 {
		p.SetIndentation(config.FirstLineCm, config.LeftCm, config.RightCm)
	}

	// 设置分页和控制属性
	p.SetKeepWithNext(config.KeepWithNext)
	p.SetKeepLines(config.KeepLines)
	p.SetPageBreakBefore(config.PageBreakBefore)
	p.SetWidowControl(config.WidowControl)

	// 设置大纲级别
	if config.OutlineLevel >= 0 && config.OutlineLevel <= 8 {
		p.SetOutlineLevel(config.OutlineLevel)
	}

	Debugf("应用段落格式配置: 对齐=%s, 样式=%s, 行距=%.1f, 段前=%d, 段后=%d",
		config.Alignment, config.Style, config.LineSpacing, config.BeforePara, config.AfterPara)
}

// ParagraphBorderConfig 段落边框配置（区别于表格边框配置）
type ParagraphBorderConfig struct {
	Style BorderStyle // 边框样式
	Size  int         // 边框粗细（1/8磅为单位，默认值建议12，即1.5磅）
	Color string      // 边框颜色（十六进制，如"000000"表示黑色）
	Space int         // 边框与文本的间距（磅，默认值建议1）
}

// SetBorder 设置段落的边框。
//
// 此方法用于为段落添加边框装饰，特别适用于实现Markdown分割线(---)的转换。
//
// 参数：
//   - top: 上边框配置，传入nil表示不设置上边框
//   - left: 左边框配置，传入nil表示不设置左边框
//   - bottom: 下边框配置，传入nil表示不设置下边框
//   - right: 右边框配置，传入nil表示不设置右边框
//
// 边框配置包含样式、粗细、颜色和间距等属性。
//
// 示例：
//
//	// 设置分割线效果（仅底边框）
//	para := doc.AddParagraph("")
//	para.SetBorder(nil, nil, &document.ParagraphBorderConfig{
//		Style: document.BorderStyleSingle,
//		Size:  12,   // 1.5磅粗细
//		Color: "000000", // 黑色
//		Space: 1,    // 1磅间距
//	}, nil)
//
//	// 设置完整边框
//	para := doc.AddParagraph("带边框的段落")
//	borderConfig := &document.ParagraphBorderConfig{
//		Style: document.BorderStyleDouble,
//		Size:  8,
//		Color: "0000FF", // 蓝色
//		Space: 2,
//	}
//	para.SetBorder(borderConfig, borderConfig, borderConfig, borderConfig)
func (p *Paragraph) SetBorder(top, left, bottom, right *ParagraphBorderConfig) {
	if p.Properties == nil {
		p.Properties = &ParagraphProperties{}
	}

	// 如果没有任何边框配置，清除边框
	if top == nil && left == nil && bottom == nil && right == nil {
		p.Properties.ParagraphBorder = nil
		return
	}

	// 创建段落边框
	if p.Properties.ParagraphBorder == nil {
		p.Properties.ParagraphBorder = &ParagraphBorder{}
	}

	// 设置上边框
	if top != nil {
		p.Properties.ParagraphBorder.Top = &ParagraphBorderLine{
			Val:   string(top.Style),
			Sz:    strconv.Itoa(top.Size),
			Color: top.Color,
			Space: strconv.Itoa(top.Space),
		}
	} else {
		p.Properties.ParagraphBorder.Top = nil
	}

	// 设置左边框
	if left != nil {
		p.Properties.ParagraphBorder.Left = &ParagraphBorderLine{
			Val:   string(left.Style),
			Sz:    strconv.Itoa(left.Size),
			Color: left.Color,
			Space: strconv.Itoa(left.Space),
		}
	} else {
		p.Properties.ParagraphBorder.Left = nil
	}

	// 设置下边框
	if bottom != nil {
		p.Properties.ParagraphBorder.Bottom = &ParagraphBorderLine{
			Val:   string(bottom.Style),
			Sz:    strconv.Itoa(bottom.Size),
			Color: bottom.Color,
			Space: strconv.Itoa(bottom.Space),
		}
	} else {
		p.Properties.ParagraphBorder.Bottom = nil
	}

	// 设置右边框
	if right != nil {
		p.Properties.ParagraphBorder.Right = &ParagraphBorderLine{
			Val:   string(right.Style),
			Sz:    strconv.Itoa(right.Size),
			Color: right.Color,
			Space: strconv.Itoa(right.Space),
		}
	} else {
		p.Properties.ParagraphBorder.Right = nil
	}

	Debugf("设置段落边框: 上=%v, 左=%v, 下=%v, 右=%v", top != nil, left != nil, bottom != nil, right != nil)
}

// SetHorizontalRule 设置水平分割线。
//
// 此方法是SetBorder的简化版本，专门用于快速创建Markdown风格的分割线效果。
// 只在段落底部添加一条水平线，适用于Markdown中的 --- 或 *** 语法。
//
// 参数：
//   - style: 边框样式，如BorderStyleSingle、BorderStyleDouble等
//   - size: 边框粗细（1/8磅为单位，建议值12-18）
//   - color: 边框颜色（十六进制，如"000000"）
//
// 示例：
//
//	// 创建简单分割线
//	para := doc.AddParagraph("")
//	para.SetHorizontalRule(document.BorderStyleSingle, 12, "000000")
//
//	// 创建粗双线分割线
//	para := doc.AddParagraph("")
//	para.SetHorizontalRule(document.BorderStyleDouble, 18, "808080")
func (p *Paragraph) SetHorizontalRule(style BorderStyle, size int, color string) {
	borderConfig := &ParagraphBorderConfig{
		Style: style,
		Size:  size,
		Color: color,
		Space: 1, // 默认1磅间距
	}

	p.SetBorder(nil, nil, borderConfig, nil)

	Debugf("设置水平分割线: 样式=%s, 粗细=%d, 颜色=%s", style, size, color)
}

// GetStyleManager 获取文档的样式管理器。
//
// 返回文档的样式管理器，可用于访问和管理样式。
//
// 示例:
//
//	doc := document.New()
//	styleManager := doc.GetStyleManager()
//	headingStyle := styleManager.GetStyle("Heading1")
func (d *Document) GetStyleManager() *style.StyleManager {
	return d.styleManager
}

// GetParts 获取文档部件映射
//
// 返回包含文档所有部件的映射，主要用于测试和调试。
// 键是部件名称，值是部件内容的字节数组。
//
// 示例:
//
//	parts := doc.GetParts()
//	settingsXML := parts["word/settings.xml"]
func (d *Document) GetParts() map[string][]byte {
	return d.parts
}

// initializeStructure 初始化文档基础结构
func (d *Document) initializeStructure() {
	// 初始化 content types
	d.contentTypes = &ContentTypes{
		Xmlns: "http://schemas.openxmlformats.org/package/2006/content-types",
		Defaults: []Default{
			{Extension: "rels", ContentType: "application/vnd.openxmlformats-package.relationships+xml"},
			{Extension: "xml", ContentType: "application/xml"},
		},
		Overrides: []Override{
			{PartName: "/word/document.xml", ContentType: "application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"},
			{PartName: "/word/styles.xml", ContentType: "application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"},
		},
	}

	// 初始化主关系
	d.relationships = &Relationships{
		Xmlns: "http://schemas.openxmlformats.org/package/2006/relationships",
		Relationships: []Relationship{
			{
				ID:     "rId1",
				Type:   "http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument",
				Target: "word/document.xml",
			},
		},
	}

	// 添加基础部件
	d.serializeContentTypes()
	d.serializeRelationships()
	d.serializeDocumentRelationships()
}

// parseDocument 解析文档内容
func (d *Document) parseDocument() error {
	Debugf("开始解析文档内容")

	// 解析主文档
	docData, ok := d.parts["word/document.xml"]
	if !ok {
		return WrapError("parse_document", ErrDocumentNotFound)
	}

	// 首先解析基本结构
	decoder := xml.NewDecoder(bytes.NewReader(docData))
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return WrapError("parse_document", err)
		}

		switch t := token.(type) {
		case xml.StartElement:
			if t.Name.Local == "document" && t.Name.Space == "http://schemas.openxmlformats.org/wordprocessingml/2006/main" {
				// 开始解析文档
				if err := d.parseDocumentElement(decoder); err != nil {
					return err
				}
				goto done
			}
		}
	}

done:
	Infof("解析完成，共 %d 个元素", len(d.Body.Elements))
	return nil
}

// parseDocumentElement 解析文档元素
func (d *Document) parseDocumentElement(decoder *xml.Decoder) error {
	// 初始化Body
	d.Body = &Body{
		Elements: make([]interface{}, 0),
	}

	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return WrapError("parse_document_element", err)
		}

		switch t := token.(type) {
		case xml.StartElement:
			switch {
			case t.Name.Local == "body":
				// 解析文档主体
				if err := d.parseBodyElement(decoder); err != nil {
					return err
				}
			}
		case xml.EndElement:
			if t.Name.Local == "document" {
				return nil
			}
		}
	}

	return nil
}

// parseBodyElement 解析文档主体元素
func (d *Document) parseBodyElement(decoder *xml.Decoder) error {
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return WrapError("parse_body_element", err)
		}

		switch t := token.(type) {
		case xml.StartElement:
			element, err := d.parseBodySubElement(decoder, t)
			if err != nil {
				return err
			}
			if element != nil {
				d.Body.Elements = append(d.Body.Elements, element)
			}
		case xml.EndElement:
			if t.Name.Local == "body" {
				return nil
			}
		}
	}

	return nil
}

// parseBodySubElement 解析文档主体的子元素
func (d *Document) parseBodySubElement(decoder *xml.Decoder, startElement xml.StartElement) (interface{}, error) {
	switch startElement.Name.Local {
	case "p":
		// 解析段落
		return d.parseParagraph(decoder, startElement)
	case "tbl":
		// 解析表格
		return d.parseTable(decoder, startElement)
	case "sectPr":
		// 解析节属性
		return d.parseSectionProperties(decoder, startElement)
	default:
		// 跳过未知元素
		Debugf("跳过未知元素: %s", startElement.Name.Local)
		return nil, d.skipElement(decoder, startElement.Name.Local)
	}
}

// parseParagraph 解析段落
func (d *Document) parseParagraph(decoder *xml.Decoder, startElement xml.StartElement) (*Paragraph, error) {
	paragraph := &Paragraph{
		Runs: make([]Run, 0),
	}

	for {
		token, err := decoder.Token()
		if err != nil {
			return nil, WrapError("parse_paragraph", err)
		}

		switch t := token.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "pPr":
				// 解析段落属性
				if err := d.parseParagraphProperties(decoder, paragraph); err != nil {
					return nil, err
				}
			case "r":
				// 解析运行
				run, err := d.parseRun(decoder, t)
				if err != nil {
					return nil, err
				}
				if run != nil {
					paragraph.Runs = append(paragraph.Runs, *run)
				}
			default:
				// 跳过其他元素
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return nil, err
				}
			}
		case xml.EndElement:
			if t.Name.Local == "p" {
				return paragraph, nil
			}
		}
	}
}

// parseParagraphProperties 解析段落属性
func (d *Document) parseParagraphProperties(decoder *xml.Decoder, paragraph *Paragraph) error {
	paragraph.Properties = &ParagraphProperties{}

	for {
		token, err := decoder.Token()
		if err != nil {
			return WrapError("parse_paragraph_properties", err)
		}

		switch t := token.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "pStyle":
				// 段落样式
				val := getAttributeValue(t.Attr, "val")
				if val != "" {
					paragraph.Properties.ParagraphStyle = &ParagraphStyle{Val: val}
				}
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return err
				}
			case "spacing":
				// 间距
				spacing := &Spacing{}
				spacing.Before = getAttributeValue(t.Attr, "before")
				spacing.After = getAttributeValue(t.Attr, "after")
				spacing.Line = getAttributeValue(t.Attr, "line")
				spacing.LineRule = getAttributeValue(t.Attr, "lineRule")
				paragraph.Properties.Spacing = spacing
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return err
				}
			case "jc":
				// 对齐
				val := getAttributeValue(t.Attr, "val")
				if val != "" {
					paragraph.Properties.Justification = &Justification{Val: val}
				}
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return err
				}
			case "ind":
				// 缩进
				indentation := &Indentation{}
				indentation.FirstLine = getAttributeValue(t.Attr, "firstLine")
				indentation.Left = getAttributeValue(t.Attr, "left")
				indentation.Right = getAttributeValue(t.Attr, "right")
				paragraph.Properties.Indentation = indentation
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return err
				}
			default:
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return err
				}
			}
		case xml.EndElement:
			if t.Name.Local == "pPr" {
				return nil
			}
		}
	}
}

// parseRun 解析运行
func (d *Document) parseRun(decoder *xml.Decoder, startElement xml.StartElement) (*Run, error) {
	run := &Run{
		Text: Text{},
	}

	for {
		token, err := decoder.Token()
		if err != nil {
			return nil, WrapError("parse_run", err)
		}

		switch t := token.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "rPr":
				// 解析运行属性
				if err := d.parseRunProperties(decoder, run); err != nil {
					return nil, err
				}
			case "t":
				// 解析文本
				space := getAttributeValue(t.Attr, "space")
				run.Text.Space = space

				// 读取文本内容
				content, err := d.readElementText(decoder, "t")
				if err != nil {
					return nil, err
				}
				run.Text.Content = content
			default:
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return nil, err
				}
			}
		case xml.EndElement:
			if t.Name.Local == "r" {
				return run, nil
			}
		}
	}
}

// parseRunProperties 解析运行属性
func (d *Document) parseRunProperties(decoder *xml.Decoder, run *Run) error {
	run.Properties = &RunProperties{}

	for {
		token, err := decoder.Token()
		if err != nil {
			return WrapError("parse_run_properties", err)
		}

		switch t := token.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "b":
				run.Properties.Bold = &Bold{}
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return err
				}
			case "bCs":
				run.Properties.BoldCs = &BoldCs{}
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return err
				}
			case "i":
				run.Properties.Italic = &Italic{}
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return err
				}
			case "iCs":
				run.Properties.ItalicCs = &ItalicCs{}
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return err
				}
			case "u":
				val := getAttributeValue(t.Attr, "val")
				run.Properties.Underline = &Underline{Val: val}
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return err
				}
			case "strike":
				run.Properties.Strike = &Strike{}
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return err
				}
			case "sz":
				val := getAttributeValue(t.Attr, "val")
				if val != "" {
					run.Properties.FontSize = &FontSize{Val: val}
				}
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return err
				}
			case "szCs":
				val := getAttributeValue(t.Attr, "val")
				if val != "" {
					run.Properties.FontSizeCs = &FontSizeCs{Val: val}
				}
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return err
				}
			case "color":
				val := getAttributeValue(t.Attr, "val")
				if val != "" {
					run.Properties.Color = &Color{Val: val}
				}
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return err
				}
			case "highlight":
				val := getAttributeValue(t.Attr, "val")
				if val != "" {
					run.Properties.Highlight = &Highlight{Val: val}
				}
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return err
				}
			case "rFonts":
				ascii := getAttributeValue(t.Attr, "ascii")
				hAnsi := getAttributeValue(t.Attr, "hAnsi")
				eastAsia := getAttributeValue(t.Attr, "eastAsia")
				cs := getAttributeValue(t.Attr, "cs")
				hint := getAttributeValue(t.Attr, "hint")

				run.Properties.FontFamily = &FontFamily{
					ASCII:    ascii,
					HAnsi:    hAnsi,
					EastAsia: eastAsia,
					CS:       cs,
					Hint:     hint,
				}
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return err
				}
			default:
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return err
				}
			}
		case xml.EndElement:
			if t.Name.Local == "rPr" {
				return nil
			}
		}
	}
}

// parseTable 解析表格
func (d *Document) parseTable(decoder *xml.Decoder, startElement xml.StartElement) (*Table, error) {
	table := &Table{
		Rows: make([]TableRow, 0),
	}

	for {
		token, err := decoder.Token()
		if err != nil {
			return nil, WrapError("parse_table", err)
		}

		switch t := token.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "tblPr":
				// 解析表格属性
				if err := d.parseTableProperties(decoder, table); err != nil {
					return nil, err
				}
			case "tblGrid":
				// 解析表格网格
				if err := d.parseTableGrid(decoder, table); err != nil {
					return nil, err
				}
			case "tr":
				// 解析表格行
				row, err := d.parseTableRow(decoder, t)
				if err != nil {
					return nil, err
				}
				if row != nil {
					table.Rows = append(table.Rows, *row)
				}
			default:
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return nil, err
				}
			}
		case xml.EndElement:
			if t.Name.Local == "tbl" {
				return table, nil
			}
		}
	}
}

// parseTableProperties 解析表格属性
func (d *Document) parseTableProperties(decoder *xml.Decoder, table *Table) error {
	table.Properties = &TableProperties{}

	for {
		token, err := decoder.Token()
		if err != nil {
			return WrapError("parse_table_properties", err)
		}

		switch t := token.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "tblW":
				w := getAttributeValue(t.Attr, "w")
				wType := getAttributeValue(t.Attr, "type")
				if w != "" || wType != "" {
					table.Properties.TableW = &TableWidth{W: w, Type: wType}
				}
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return err
				}
			case "jc":
				val := getAttributeValue(t.Attr, "val")
				if val != "" {
					table.Properties.TableJc = &TableJc{Val: val}
				}
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return err
				}
			case "tblLook":
				// 解析表格外观
				tableLook := &TableLook{
					Val:      getAttributeValue(t.Attr, "val"),
					FirstRow: getAttributeValue(t.Attr, "firstRow"),
					LastRow:  getAttributeValue(t.Attr, "lastRow"),
					FirstCol: getAttributeValue(t.Attr, "firstColumn"),
					LastCol:  getAttributeValue(t.Attr, "lastColumn"),
					NoHBand:  getAttributeValue(t.Attr, "noHBand"),
					NoVBand:  getAttributeValue(t.Attr, "noVBand"),
				}
				table.Properties.TableLook = tableLook
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return err
				}
			case "tblStyle":
				// 解析表格样式
				val := getAttributeValue(t.Attr, "val")
				if val != "" {
					table.Properties.TableStyle = &TableStyle{Val: val}
				}
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return err
				}
			case "tblBorders":
				// 解析表格边框
				borders, err := d.parseTableBorders(decoder)
				if err != nil {
					return err
				}
				table.Properties.TableBorders = borders
			case "shd":
				// 解析表格底纹
				shd := &TableShading{
					Val:       getAttributeValue(t.Attr, "val"),
					Color:     getAttributeValue(t.Attr, "color"),
					Fill:      getAttributeValue(t.Attr, "fill"),
					ThemeFill: getAttributeValue(t.Attr, "themeFill"),
				}
				table.Properties.Shd = shd
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return err
				}
			case "tblCellMar":
				// 解析表格单元格边距
				margins, err := d.parseTableCellMargins(decoder)
				if err != nil {
					return err
				}
				table.Properties.TableCellMar = margins
			case "tblLayout":
				// 解析表格布局
				layoutType := getAttributeValue(t.Attr, "type")
				if layoutType != "" {
					table.Properties.TableLayout = &TableLayoutType{Type: layoutType}
				}
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return err
				}
			case "tblInd":
				// 解析表格缩进
				w := getAttributeValue(t.Attr, "w")
				indType := getAttributeValue(t.Attr, "type")
				if w != "" || indType != "" {
					table.Properties.TableInd = &TableIndentation{W: w, Type: indType}
				}
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return err
				}
			default:
				// 跳过其他表格属性，可以根据需要扩展
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return err
				}
			}
		case xml.EndElement:
			if t.Name.Local == "tblPr" {
				return nil
			}
		}
	}
}

// parseTableGrid 解析表格网格
func (d *Document) parseTableGrid(decoder *xml.Decoder, table *Table) error {
	table.Grid = &TableGrid{
		Cols: make([]TableGridCol, 0),
	}

	for {
		token, err := decoder.Token()
		if err != nil {
			return WrapError("parse_table_grid", err)
		}

		switch t := token.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "gridCol":
				w := getAttributeValue(t.Attr, "w")
				col := TableGridCol{W: w}
				table.Grid.Cols = append(table.Grid.Cols, col)
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return err
				}
			default:
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return err
				}
			}
		case xml.EndElement:
			if t.Name.Local == "tblGrid" {
				return nil
			}
		}
	}
}

// parseTableRow 解析表格行
func (d *Document) parseTableRow(decoder *xml.Decoder, startElement xml.StartElement) (*TableRow, error) {
	row := &TableRow{
		Cells: make([]TableCell, 0),
	}

	for {
		token, err := decoder.Token()
		if err != nil {
			return nil, WrapError("parse_table_row", err)
		}

		switch t := token.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "trPr":
				// 解析行属性
				props, err := d.parseTableRowProperties(decoder)
				if err != nil {
					return nil, err
				}
				row.Properties = props
			case "tc":
				// 解析表格单元格
				cell, err := d.parseTableCell(decoder, t)
				if err != nil {
					return nil, err
				}
				if cell != nil {
					row.Cells = append(row.Cells, *cell)
				}
			default:
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return nil, err
				}
			}
		case xml.EndElement:
			if t.Name.Local == "tr" {
				return row, nil
			}
		}
	}
}

// parseTableCell 解析表格单元格
func (d *Document) parseTableCell(decoder *xml.Decoder, startElement xml.StartElement) (*TableCell, error) {
	cell := &TableCell{
		Paragraphs: make([]Paragraph, 0),
	}

	for {
		token, err := decoder.Token()
		if err != nil {
			return nil, WrapError("parse_table_cell", err)
		}

		switch t := token.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "tcPr":
				// 解析单元格属性
				props, err := d.parseTableCellProperties(decoder)
				if err != nil {
					return nil, err
				}
				cell.Properties = props
			case "p":
				// 解析段落
				para, err := d.parseParagraph(decoder, t)
				if err != nil {
					return nil, err
				}
				if para != nil {
					cell.Paragraphs = append(cell.Paragraphs, *para)
				}
			default:
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return nil, err
				}
			}
		case xml.EndElement:
			if t.Name.Local == "tc" {
				return cell, nil
			}
		}
	}
}

// parseSectionProperties 解析节属性
func (d *Document) parseSectionProperties(decoder *xml.Decoder, startElement xml.StartElement) (*SectionProperties, error) {
	sectPr := &SectionProperties{}

	for {
		token, err := decoder.Token()
		if err != nil {
			return nil, WrapError("parse_section_properties", err)
		}

		switch t := token.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "pgSz":
				// 解析页面尺寸
				w := getAttributeValue(t.Attr, "w")
				h := getAttributeValue(t.Attr, "h")
				orient := getAttributeValue(t.Attr, "orient")
				if w != "" || h != "" {
					sectPr.PageSize = &PageSizeXML{W: w, H: h, Orient: orient}
				}
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return nil, err
				}
			case "pgMar":
				// 解析页面边距
				margin := &PageMargin{}
				margin.Top = getAttributeValue(t.Attr, "top")
				margin.Right = getAttributeValue(t.Attr, "right")
				margin.Bottom = getAttributeValue(t.Attr, "bottom")
				margin.Left = getAttributeValue(t.Attr, "left")
				margin.Header = getAttributeValue(t.Attr, "header")
				margin.Footer = getAttributeValue(t.Attr, "footer")
				margin.Gutter = getAttributeValue(t.Attr, "gutter")
				sectPr.PageMargins = margin
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return nil, err
				}
			case "cols":
				// 解析分栏
				space := getAttributeValue(t.Attr, "space")
				num := getAttributeValue(t.Attr, "num")
				if space != "" || num != "" {
					sectPr.Columns = &Columns{Space: space, Num: num}
				}
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return nil, err
				}
			case "docGrid":
				// 解析文档网格
				docGridType := getAttributeValue(t.Attr, "type")
				linePitch := getAttributeValue(t.Attr, "linePitch")
				charSpace := getAttributeValue(t.Attr, "charSpace")
				if docGridType != "" || linePitch != "" || charSpace != "" {
					sectPr.DocGrid = &DocGrid{
						Type:      docGridType,
						LinePitch: linePitch,
						CharSpace: charSpace,
					}
				}
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return nil, err
				}
			default:
				// 跳过其他节属性
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return nil, err
				}
			}
		case xml.EndElement:
			if t.Name.Local == "sectPr" {
				return sectPr, nil
			}
		}
	}
}

// skipElement 跳过元素及其子元素
func (d *Document) skipElement(decoder *xml.Decoder, elementName string) error {
	depth := 1
	for depth > 0 {
		token, err := decoder.Token()
		if err != nil {
			return WrapError("skip_element", err)
		}

		switch token.(type) {
		case xml.StartElement:
			depth++
		case xml.EndElement:
			depth--
		}
	}
	return nil
}

// readElementText 读取元素的文本内容
func (d *Document) readElementText(decoder *xml.Decoder, elementName string) (string, error) {
	var content string
	for {
		token, err := decoder.Token()
		if err != nil {
			return "", WrapError("read_element_text", err)
		}

		switch t := token.(type) {
		case xml.CharData:
			content += string(t)
		case xml.EndElement:
			if t.Name.Local == elementName {
				return content, nil
			}
		}
	}
}

// getAttributeValue 获取属性值
func getAttributeValue(attrs []xml.Attr, name string) string {
	for _, attr := range attrs {
		if attr.Name.Local == name {
			return attr.Value
		}
	}
	return ""
}

// serializeDocument 序列化文档内容
func (d *Document) serializeDocument() error {
	Debugf("开始序列化文档")

	// 创建文档结构
	type documentXML struct {
		XMLName  xml.Name `xml:"w:document"`
		Xmlns    string   `xml:"xmlns:w,attr"`
		XmlnsW15 string   `xml:"xmlns:w15,attr"`
		XmlnsWP  string   `xml:"xmlns:wp,attr"`
		XmlnsA   string   `xml:"xmlns:a,attr"`
		XmlnsPic string   `xml:"xmlns:pic,attr"`
		XmlnsR   string   `xml:"xmlns:r,attr"`
		Body     *Body    `xml:"w:body"`
	}

	doc := documentXML{
		Xmlns:    "http://schemas.openxmlformats.org/wordprocessingml/2006/main",
		XmlnsW15: "http://schemas.microsoft.com/office/word/2012/wordml",
		XmlnsWP:  "http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing",
		XmlnsA:   "http://schemas.openxmlformats.org/drawingml/2006/main",
		XmlnsPic: "http://schemas.openxmlformats.org/drawingml/2006/picture",
		XmlnsR:   "http://schemas.openxmlformats.org/officeDocument/2006/relationships",
		Body:     d.Body,
	}

	// 序列化为XML
	data, err := xml.MarshalIndent(doc, "", "  ")
	if err != nil {
		Errorf("XML序列化失败: %v", err)
		return WrapError("marshal_xml", err)
	}

	// 添加XML声明
	d.parts["word/document.xml"] = append([]byte(xml.Header), data...)

	Debugf("文档序列化完成")
	return nil
}

// serializeContentTypes 序列化内容类型
func (d *Document) serializeContentTypes() {
	data, _ := xml.MarshalIndent(d.contentTypes, "", "  ")
	d.parts["[Content_Types].xml"] = append([]byte(xml.Header), data...)
}

// serializeRelationships 序列化关系
func (d *Document) serializeRelationships() {
	data, _ := xml.MarshalIndent(d.relationships, "", "  ")
	d.parts["_rels/.rels"] = append([]byte(xml.Header), data...)
}

// serializeDocumentRelationships 序列化文档关系
func (d *Document) serializeDocumentRelationships() {
	// 获取已存在的关系，从索引1开始（保留给styles.xml）
	relationships := []Relationship{
		{
			ID:     "rId1",
			Type:   "http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles",
			Target: "styles.xml",
		},
	}

	// 添加动态创建的文档级关系（如页眉、页脚等）
	relationships = append(relationships, d.documentRelationships.Relationships...)

	// 创建文档关系
	docRels := &Relationships{
		Xmlns:         "http://schemas.openxmlformats.org/package/2006/relationships",
		Relationships: relationships,
	}

	data, _ := xml.MarshalIndent(docRels, "", "  ")
	d.parts["word/_rels/document.xml.rels"] = append([]byte(xml.Header), data...)
}

// serializeStyles 序列化样式
func (d *Document) serializeStyles() error {
	Debugf("开始序列化样式")

	// 如果在克隆文档时已经保留了完整的 styles.xml（含 docDefaults 等信息），
	// 这里直接跳过重新生成，避免丢失模板原有的默认段落/字符设置。
	if existing, ok := d.parts["word/styles.xml"]; ok && len(existing) > 0 {
		Debugf("检测到已有 styles.xml，跳过样式重建以保留模板默认样式")
		return nil
	}

	// 创建样式结构，包含完整的命名空间
	type stylesXML struct {
		XMLName     xml.Name       `xml:"w:styles"`
		XmlnsW      string         `xml:"xmlns:w,attr"`
		XmlnsMC     string         `xml:"xmlns:mc,attr"`
		XmlnsO      string         `xml:"xmlns:o,attr"`
		XmlnsR      string         `xml:"xmlns:r,attr"`
		XmlnsM      string         `xml:"xmlns:m,attr"`
		XmlnsV      string         `xml:"xmlns:v,attr"`
		XmlnsW14    string         `xml:"xmlns:w14,attr"`
		XmlnsW10    string         `xml:"xmlns:w10,attr"`
		XmlnsSL     string         `xml:"xmlns:sl,attr"`
		XmlnsWPS    string         `xml:"xmlns:wpsCustomData,attr"`
		MCIgnorable string         `xml:"mc:Ignorable,attr"`
		Styles      []*style.Style `xml:"w:style"`
	}

	doc := stylesXML{
		XmlnsW:      "http://schemas.openxmlformats.org/wordprocessingml/2006/main",
		XmlnsMC:     "http://schemas.openxmlformats.org/markup-compatibility/2006",
		XmlnsO:      "urn:schemas-microsoft-com:office:office",
		XmlnsR:      "http://schemas.openxmlformats.org/officeDocument/2006/relationships",
		XmlnsM:      "http://schemas.openxmlformats.org/officeDocument/2006/math",
		XmlnsV:      "urn:schemas-microsoft-com:vml",
		XmlnsW14:    "http://schemas.microsoft.com/office/word/2010/wordml",
		XmlnsW10:    "urn:schemas-microsoft-com:office:word",
		XmlnsSL:     "http://schemas.openxmlformats.org/schemaLibrary/2006/main",
		XmlnsWPS:    "http://www.wps.cn/officeDocument/2013/wpsCustomData",
		MCIgnorable: "w14",
		Styles:      d.styleManager.GetAllStyles(),
	}

	// 序列化为XML
	data, err := xml.MarshalIndent(doc, "", "  ")
	if err != nil {
		Errorf("XML序列化失败: %v", err)
		return WrapError("marshal_xml", err)
	}

	// 添加XML声明
	d.parts["word/styles.xml"] = append([]byte(xml.Header), data...)

	Debugf("样式序列化完成")
	return nil
}

// parseContentTypes 解析内容类型文件
func (d *Document) parseContentTypes() error {
	Debugf("开始解析内容类型文件")

	// 查找内容类型文件
	contentTypesData, ok := d.parts["[Content_Types].xml"]
	if !ok {
		return WrapError("parse_content_types", fmt.Errorf("内容类型文件不存在"))
	}

	// 解析XML
	var contentTypes ContentTypes
	if err := xml.Unmarshal(contentTypesData, &contentTypes); err != nil {
		return WrapError("parse_content_types", err)
	}

	d.contentTypes = &contentTypes
	Debugf("内容类型解析完成")
	return nil
}

// parseRelationships 解析关系文件
func (d *Document) parseRelationships() error {
	Debugf("开始解析关系文件")

	// 查找关系文件
	relsData, ok := d.parts["_rels/.rels"]
	if !ok {
		return WrapError("parse_relationships", fmt.Errorf("关系文件不存在"))
	}

	// 解析XML
	var relationships Relationships
	if err := xml.Unmarshal(relsData, &relationships); err != nil {
		return WrapError("parse_relationships", err)
	}

	d.relationships = &relationships
	Debugf("关系解析完成")
	return nil
}

// parseStyles 解析样式文件
func (d *Document) parseStyles() error {
	Debugf("开始解析样式文件")

	// 查找样式文件
	stylesData, ok := d.parts["word/styles.xml"]
	if !ok {
		return WrapError("parse_styles", fmt.Errorf("样式文件不存在"))
	}

	// 使用样式管理器解析样式
	if err := d.styleManager.LoadStylesFromDocument(stylesData); err != nil {
		return WrapError("parse_styles", err)
	}

	Debugf("样式解析完成")
	return nil
}

// parseDocumentRelationships 解析文档关系文件（word/_rels/document.xml.rels）
// 该文件包含文档中图片、页眉、页脚等资源的关系
func (d *Document) parseDocumentRelationships() error {
	Debugf("开始解析文档关系文件")

	// 查找文档关系文件
	docRelsData, ok := d.parts["word/_rels/document.xml.rels"]
	if !ok {
		// 文档可能没有关系文件（没有图片等资源），这不是错误
		Debugf("文档关系文件不存在，文档可能不包含图片等资源")
		return nil
	}

	// 解析XML
	var relationships Relationships
	if err := xml.Unmarshal(docRelsData, &relationships); err != nil {
		return WrapError("parse_document_relationships", err)
	}

	// 保存解析的关系（不包括styles.xml，因为它在serializeDocumentRelationships中会自动添加）
	// 过滤掉styles.xml的关系，因为它总是rId1并在保存时自动添加
	filteredRels := make([]Relationship, 0)
	for _, rel := range relationships.Relationships {
		if rel.Type != "http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" {
			filteredRels = append(filteredRels, rel)
		}
	}

	d.documentRelationships.Relationships = filteredRels
	Debugf("文档关系解析完成，共 %d 个关系", len(filteredRels))
	return nil
}

// updateNextImageID 根据已有的图片关系更新nextImageID计数器
// 确保新添加的图片ID不会与现有图片冲突
func (d *Document) updateNextImageID() {
	maxImageID := -1

	// 遍历所有parts，查找已存在的图片文件的最大ID
	for partName := range d.parts {
		// 检查是否是图片文件（word/media/imageN.xxx）
		if len(partName) > 11 && partName[:11] == "word/media/" {
			// 从文件名中提取图片ID（image0.png -> 0, image1.png -> 1等）
			filename := partName[11:] // 去掉"word/media/"前缀
			var id int
			if _, err := fmt.Sscanf(filename, "image%d.", &id); err == nil {
				if id > maxImageID {
					maxImageID = id
				}
			}
		}
	}

	// 设置nextImageID为最大图片ID + 1
	// 如果没有现有图片，maxImageID为-1，nextImageID应该为0
	d.nextImageID = maxImageID + 1

	Debugf("更新图片ID计数器: nextImageID = %d", d.nextImageID)
}

// ToBytes 将文档转换为字节数组
func (d *Document) ToBytes() ([]byte, error) {
	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	// 序列化文档
	if err := d.serializeDocument(); err != nil {
		return nil, err
	}

	// 序列化样式
	if err := d.serializeStyles(); err != nil {
		return nil, err
	}

	// 序列化内容类型
	d.serializeContentTypes()

	// 序列化关系
	d.serializeRelationships()

	// 序列化文档关系
	d.serializeDocumentRelationships()

	// 写入所有部件
	for name, data := range d.parts {
		writer, err := zipWriter.Create(name)
		if err != nil {
			return nil, err
		}
		if _, err := writer.Write(data); err != nil {
			return nil, err
		}
	}

	if err := zipWriter.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// GetParagraphs 获取所有段落
func (b *Body) GetParagraphs() []*Paragraph {
	paragraphs := make([]*Paragraph, 0)
	for _, element := range b.Elements {
		if p, ok := element.(*Paragraph); ok {
			paragraphs = append(paragraphs, p)
		}
	}
	return paragraphs
}

// GetTables 获取所有表格
func (b *Body) GetTables() []*Table {
	tables := make([]*Table, 0)
	for _, element := range b.Elements {
		if t, ok := element.(*Table); ok {
			tables = append(tables, t)
		}
	}
	return tables
}

// AddElement 添加元素到文档主体
func (b *Body) AddElement(element interface{}) {
	b.Elements = append(b.Elements, element)
}

// RemoveParagraph 从文档中删除指定的段落。
//
// 参数 paragraph 是要删除的段落对象。
// 如果段落不存在于文档中，此方法不会产生任何效果。
//
// 返回值表示是否成功删除段落。
//
// 示例:
//
//	doc := document.New()
//	para := doc.AddParagraph("要删除的段落")
//	doc.RemoveParagraph(para)
func (d *Document) RemoveParagraph(paragraph *Paragraph) bool {
	for i, element := range d.Body.Elements {
		if p, ok := element.(*Paragraph); ok && p == paragraph {
			// 删除元素
			d.Body.Elements = append(d.Body.Elements[:i], d.Body.Elements[i+1:]...)
			Debugf("删除段落: 索引 %d", i)
			return true
		}
	}
	Debugf("警告：未找到要删除的段落")
	return false
}

// RemoveParagraphAt 根据索引删除段落。
//
// 参数 index 是要删除的段落在所有段落中的索引（从0开始）。
// 如果索引超出范围，此方法会返回错误。
//
// 返回值表示是否成功删除段落。
//
// 示例:
//
//	doc := document.New()
//	doc.AddParagraph("第一段")
//	doc.AddParagraph("第二段")
//	doc.RemoveParagraphAt(0)  // 删除第一段
func (d *Document) RemoveParagraphAt(index int) bool {
	// 提前验证负数索引
	if index < 0 {
		Debugf("错误：段落索引不能为负数: %d", index)
		return false
	}

	// 优化：单次遍历找到目标段落及其元素索引
	paragraphCount := 0
	for i, element := range d.Body.Elements {
		if _, ok := element.(*Paragraph); ok {
			if paragraphCount == index {
				// 找到目标段落，删除它
				d.Body.Elements = append(d.Body.Elements[:i], d.Body.Elements[i+1:]...)
				Debugf("删除段落: 段落索引 %d, 元素索引 %d", index, i)
				return true
			}
			paragraphCount++
		}
	}

	Debugf("错误：段落索引 %d 超出范围 [0, %d)", index, paragraphCount)
	return false
}

// RemoveElementAt 根据元素索引删除元素（包括段落、表格等）。
//
// 参数 index 是要删除的元素在文档主体中的索引（从0开始）。
// 如果索引超出范围，此方法会返回错误。
//
// 返回值表示是否成功删除元素。
//
// 示例:
//
//	doc := document.New()
//	doc.AddParagraph("段落")
//	doc.AddTable(&document.TableConfig{Rows: 2, Cols: 2})
//	doc.RemoveElementAt(0)  // 删除第一个元素（段落）
func (d *Document) RemoveElementAt(index int) bool {
	if index < 0 || index >= len(d.Body.Elements) {
		Debugf("错误：元素索引 %d 超出范围 [0, %d)", index, len(d.Body.Elements))
		return false
	}

	// 删除元素
	d.Body.Elements = append(d.Body.Elements[:index], d.Body.Elements[index+1:]...)
	Debugf("删除元素: 索引 %d", index)
	return true
}

// parseTableBorders 解析表格边框
func (d *Document) parseTableBorders(decoder *xml.Decoder) (*TableBorders, error) {
	borders := &TableBorders{}

	for {
		token, err := decoder.Token()
		if err != nil {
			return nil, WrapError("parse_table_borders", err)
		}

		switch t := token.(type) {
		case xml.StartElement:
			border := &TableBorder{
				Val:        getAttributeValue(t.Attr, "val"),
				Sz:         getAttributeValue(t.Attr, "sz"),
				Space:      getAttributeValue(t.Attr, "space"),
				Color:      getAttributeValue(t.Attr, "color"),
				ThemeColor: getAttributeValue(t.Attr, "themeColor"),
			}

			switch t.Name.Local {
			case "top":
				borders.Top = border
			case "left":
				borders.Left = border
			case "bottom":
				borders.Bottom = border
			case "right":
				borders.Right = border
			case "insideH":
				borders.InsideH = border
			case "insideV":
				borders.InsideV = border
			}

			if err := d.skipElement(decoder, t.Name.Local); err != nil {
				return nil, err
			}
		case xml.EndElement:
			if t.Name.Local == "tblBorders" {
				return borders, nil
			}
		}
	}
}

// parseTableCellMargins 解析表格单元格边距
func (d *Document) parseTableCellMargins(decoder *xml.Decoder) (*TableCellMargins, error) {
	margins := &TableCellMargins{}

	for {
		token, err := decoder.Token()
		if err != nil {
			return nil, WrapError("parse_table_cell_margins", err)
		}

		switch t := token.(type) {
		case xml.StartElement:
			space := &TableCellSpace{
				W:    getAttributeValue(t.Attr, "w"),
				Type: getAttributeValue(t.Attr, "type"),
			}

			switch t.Name.Local {
			case "top":
				margins.Top = space
			case "left":
				margins.Left = space
			case "bottom":
				margins.Bottom = space
			case "right":
				margins.Right = space
			}

			if err := d.skipElement(decoder, t.Name.Local); err != nil {
				return nil, err
			}
		case xml.EndElement:
			if t.Name.Local == "tblCellMar" {
				return margins, nil
			}
		}
	}
}

// parseTableCellProperties 解析表格单元格属性
func (d *Document) parseTableCellProperties(decoder *xml.Decoder) (*TableCellProperties, error) {
	props := &TableCellProperties{}

	for {
		token, err := decoder.Token()
		if err != nil {
			return nil, WrapError("parse_table_cell_properties", err)
		}

		switch t := token.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "tcW":
				// 解析单元格宽度
				w := getAttributeValue(t.Attr, "w")
				wType := getAttributeValue(t.Attr, "type")
				if w != "" || wType != "" {
					props.TableCellW = &TableCellW{W: w, Type: wType}
				}
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return nil, err
				}
			case "vAlign":
				// 解析垂直对齐
				val := getAttributeValue(t.Attr, "val")
				if val != "" {
					props.VAlign = &VAlign{Val: val}
				}
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return nil, err
				}
			case "gridSpan":
				// 解析网格跨度
				val := getAttributeValue(t.Attr, "val")
				if val != "" {
					props.GridSpan = &GridSpan{Val: val}
				}
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return nil, err
				}
			case "vMerge":
				// 解析垂直合并
				val := getAttributeValue(t.Attr, "val")
				props.VMerge = &VMerge{Val: val}
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return nil, err
				}
			case "textDirection":
				// 解析文字方向
				val := getAttributeValue(t.Attr, "val")
				if val != "" {
					props.TextDirection = &TextDirection{Val: val}
				}
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return nil, err
				}
			case "shd":
				// 解析单元格底纹
				shd := &TableCellShading{
					Val:       getAttributeValue(t.Attr, "val"),
					Color:     getAttributeValue(t.Attr, "color"),
					Fill:      getAttributeValue(t.Attr, "fill"),
					ThemeFill: getAttributeValue(t.Attr, "themeFill"),
				}
				props.Shd = shd
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return nil, err
				}
			case "tcBorders":
				// 解析单元格边框
				borders, err := d.parseTableCellBorders(decoder)
				if err != nil {
					return nil, err
				}
				props.TcBorders = borders
			case "tcMar":
				// 解析单元格边距
				margins, err := d.parseTableCellMarginsCell(decoder)
				if err != nil {
					return nil, err
				}
				props.TcMar = margins
			case "noWrap":
				// 解析禁止换行
				val := getAttributeValue(t.Attr, "val")
				props.NoWrap = &NoWrap{Val: val}
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return nil, err
				}
			case "hideMark":
				// 解析隐藏标记
				val := getAttributeValue(t.Attr, "val")
				props.HideMark = &HideMark{Val: val}
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return nil, err
				}
			default:
				// 跳过其他未处理的单元格属性
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return nil, err
				}
			}
		case xml.EndElement:
			if t.Name.Local == "tcPr" {
				return props, nil
			}
		}
	}
}

// parseTableCellBorders 解析表格单元格边框
func (d *Document) parseTableCellBorders(decoder *xml.Decoder) (*TableCellBorders, error) {
	borders := &TableCellBorders{}

	for {
		token, err := decoder.Token()
		if err != nil {
			return nil, WrapError("parse_table_cell_borders", err)
		}

		switch t := token.(type) {
		case xml.StartElement:
			border := &TableCellBorder{
				Val:        getAttributeValue(t.Attr, "val"),
				Sz:         getAttributeValue(t.Attr, "sz"),
				Space:      getAttributeValue(t.Attr, "space"),
				Color:      getAttributeValue(t.Attr, "color"),
				ThemeColor: getAttributeValue(t.Attr, "themeColor"),
			}

			switch t.Name.Local {
			case "top":
				borders.Top = border
			case "left":
				borders.Left = border
			case "bottom":
				borders.Bottom = border
			case "right":
				borders.Right = border
			case "insideH":
				borders.InsideH = border
			case "insideV":
				borders.InsideV = border
			case "tl2br":
				borders.TL2BR = border
			case "tr2bl":
				borders.TR2BL = border
			}

			if err := d.skipElement(decoder, t.Name.Local); err != nil {
				return nil, err
			}
		case xml.EndElement:
			if t.Name.Local == "tcBorders" {
				return borders, nil
			}
		}
	}
}

// parseTableCellMarginsCell 解析表格单元格边距（单元格级别）
func (d *Document) parseTableCellMarginsCell(decoder *xml.Decoder) (*TableCellMarginsCell, error) {
	margins := &TableCellMarginsCell{}

	for {
		token, err := decoder.Token()
		if err != nil {
			return nil, WrapError("parse_table_cell_margins_cell", err)
		}

		switch t := token.(type) {
		case xml.StartElement:
			space := &TableCellSpaceCell{
				W:    getAttributeValue(t.Attr, "w"),
				Type: getAttributeValue(t.Attr, "type"),
			}

			switch t.Name.Local {
			case "top":
				margins.Top = space
			case "left":
				margins.Left = space
			case "bottom":
				margins.Bottom = space
			case "right":
				margins.Right = space
			}

			if err := d.skipElement(decoder, t.Name.Local); err != nil {
				return nil, err
			}
		case xml.EndElement:
			if t.Name.Local == "tcMar" {
				return margins, nil
			}
		}
	}
}

// parseTableRowProperties 解析表格行属性
func (d *Document) parseTableRowProperties(decoder *xml.Decoder) (*TableRowProperties, error) {
	props := &TableRowProperties{}

	for {
		token, err := decoder.Token()
		if err != nil {
			return nil, WrapError("parse_table_row_properties", err)
		}

		switch t := token.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "trHeight":
				// 解析行高
				val := getAttributeValue(t.Attr, "val")
				hRule := getAttributeValue(t.Attr, "hRule")
				if val != "" || hRule != "" {
					props.TableRowH = &TableRowH{Val: val, HRule: hRule}
				}
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return nil, err
				}
			case "cantSplit":
				// 解析禁止跨页分割
				val := getAttributeValue(t.Attr, "val")
				props.CantSplit = &CantSplit{Val: val}
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return nil, err
				}
			case "tblHeader":
				// 解析标题行重复
				val := getAttributeValue(t.Attr, "val")
				props.TblHeader = &TblHeader{Val: val}
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return nil, err
				}
			default:
				// 跳过其他行属性
				if err := d.skipElement(decoder, t.Name.Local); err != nil {
					return nil, err
				}
			}
		case xml.EndElement:
			if t.Name.Local == "trPr" {
				return props, nil
			}
		}
	}
}
