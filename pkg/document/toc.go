// Package document 提供Word文档目录生成功能
package document

import (
	"encoding/xml"
	"fmt"
	"strings"
)

// TOCConfig 目录配置
type TOCConfig struct {
	Title        string // 目录标题，默认为"目录"
	MaxLevel     int    // 最大级别，默认为3（显示1-3级标题）
	ShowPageNum  bool   // 是否显示页码，默认为true
	RightAlign   bool   // 页码是否右对齐，默认为true
	UseHyperlink bool   // 是否使用超链接，默认为true
	DotLeader    bool   // 是否使用点状引导线，默认为true
}

// TOCEntry 目录条目
type TOCEntry struct {
	Text       string // 条目文本
	Level      int    // 级别（1-9）
	PageNum    int    // 页码
	BookmarkID string // 书签ID（用于超链接）
}

// TOCField 目录域
type TOCField struct {
	XMLName xml.Name `xml:"w:fldSimple"`
	Instr   string   `xml:"w:instr,attr"`
	Runs    []Run    `xml:"w:r"`
}

// Hyperlink 超链接结构
type Hyperlink struct {
	XMLName xml.Name `xml:"w:hyperlink"`
	Anchor  string   `xml:"w:anchor,attr,omitempty"`
	Runs    []Run    `xml:"w:r"`
}

// BookmarkEnd 书签结束
type BookmarkEnd struct {
	XMLName xml.Name `xml:"w:bookmarkEnd"`
	ID      string   `xml:"w:id,attr"`
}

// ElementType 返回书签结束元素类型
func (b *BookmarkEnd) ElementType() string {
	return "bookmarkEnd"
}

// BookmarkStart 书签开始
type BookmarkStart struct {
	XMLName xml.Name `xml:"w:bookmarkStart"`
	ID      string   `xml:"w:id,attr"`
	Name    string   `xml:"w:name,attr"`
}

// ElementType 返回书签开始元素类型
func (b *BookmarkStart) ElementType() string {
	return "bookmarkStart"
}

// DefaultTOCConfig 返回默认目录配置
func DefaultTOCConfig() *TOCConfig {
	return &TOCConfig{
		Title:        "目录",
		MaxLevel:     3,
		ShowPageNum:  true,
		RightAlign:   true,
		UseHyperlink: true,
		DotLeader:    true,
	}
}

// GenerateTOC 生成目录
func (d *Document) GenerateTOC(config *TOCConfig) error {
	if config == nil {
		config = DefaultTOCConfig()
	}

	// 收集标题信息
	entries := d.collectHeadings(config.MaxLevel)

	// 创建目录SDT
	tocSDT := d.CreateTOCSDT(config.Title, config.MaxLevel)

	// 为每个标题条目添加到目录中
	for i, entry := range entries {
		entryID := fmt.Sprintf("14746%d", 3000+i)
		tocSDT.AddTOCEntry(entry.Text, entry.Level, entry.PageNum, entryID)
	}

	// 完成目录SDT构建
	tocSDT.FinalizeTOCSDT()

	// 添加到文档中
	d.Body.Elements = append(d.Body.Elements, tocSDT)

	return nil
}

// UpdateTOC 更新目录
func (d *Document) UpdateTOC() error {
	// 查找现有目录SDT
	tocSDT, tocIndex := d.findTOCSDT()
	if tocSDT == nil {
		// 如果没有找到SDT类型的TOC，尝试查找段落类型的TOC
		tocStart := d.findTOCStart()
		if tocStart == -1 {
			return fmt.Errorf("未找到目录")
		}

		// 删除现有目录条目
		d.removeTOCEntries(tocStart)

		// 重新生成目录条目
		config := DefaultTOCConfig()
		entries := d.collectHeadings(config.MaxLevel)
		for _, entry := range entries {
			if err := d.addTOCEntry(entry, config); err != nil {
				return fmt.Errorf("更新目录条目失败: %v", err)
			}
		}
		return nil
	}

	// 处理SDT类型的TOC
	// 从TOC配置中获取MaxLevel，默认为3
	config := DefaultTOCConfig()
	if tocSDT.Properties != nil && tocSDT.Properties.DocPartObj != nil {
		// 可以从SDT属性中提取配置，这里使用默认配置
		config = DefaultTOCConfig()
	}

	// 重新收集标题信息
	entries := d.collectHeadings(config.MaxLevel)

	// 清空SDT内容并重建
	tocSDT.Content.Elements = []interface{}{}

	// 添加目录标题段落
	titlePara := &Paragraph{
		Properties: &ParagraphProperties{
			Spacing: &Spacing{
				Before: "0",
				After:  "0",
				Line:   "240",
			},
			Indentation: &Indentation{
				Left:      "0",
				Right:     "0",
				FirstLine: "0",
			},
			Justification: &Justification{Val: "center"},
		},
		Runs: []Run{
			{
				Text: Text{Content: config.Title},
				Properties: &RunProperties{
					FontFamily: &FontFamily{ASCII: "宋体"},
					FontSize:   &FontSize{Val: "21"},
				},
			},
		},
	}

	// 添加书签开始
	bookmarkStart := &BookmarkStart{
		ID:   "0",
		Name: "_Toc11693_WPSOffice_Type3",
	}

	tocSDT.Content.Elements = append(tocSDT.Content.Elements, bookmarkStart, titlePara)

	// 为每个标题条目添加到目录中
	for i, entry := range entries {
		entryID := fmt.Sprintf("14746%d", 3000+i)
		tocSDT.AddTOCEntry(entry.Text, entry.Level, entry.PageNum, entryID)
	}

	// 完成目录SDT构建
	tocSDT.FinalizeTOCSDT()

	// 更新文档中的SDT
	d.Body.Elements[tocIndex] = tocSDT

	return nil
}

// AddHeadingWithBookmark 添加带书签的标题
func (d *Document) AddHeadingWithBookmark(text string, level int, bookmarkName string) *Paragraph {
	if bookmarkName == "" {
		bookmarkName = fmt.Sprintf("_Toc_%s", strings.ReplaceAll(text, " ", "_"))
	}

	// 添加书签开始
	bookmarkID := fmt.Sprintf("%d", len(d.Body.Elements))
	bookmark := &BookmarkStart{
		ID:   bookmarkID,
		Name: bookmarkName,
	}

	// 创建标题段落
	paragraph := d.AddHeadingParagraph(text, level)

	// 在段落的Run中插入书签
	if len(paragraph.Runs) > 0 {
		// 在第一个Run前插入书签开始
		bookmarkRun := Run{
			Properties: &RunProperties{},
		}
		// 这里需要一个特殊的XML序列化处理来插入书签元素
		paragraph.Runs = append([]Run{bookmarkRun}, paragraph.Runs...)
	}

	// 添加书签结束
	bookmarkEnd := &BookmarkEnd{
		ID: bookmarkID,
	}

	// 将书签添加到文档中（简化处理）
	_ = bookmark // 标记已使用
	d.Body.Elements = append(d.Body.Elements, bookmarkEnd)

	return paragraph
}

// collectHeadings 收集标题信息
func (d *Document) collectHeadings(maxLevel int) []TOCEntry {
	var entries []TOCEntry
	pageNum := 1 // 简化处理，实际需要计算真实页码

	for _, element := range d.Body.Elements {
		if paragraph, ok := element.(*Paragraph); ok {
			level := d.getHeadingLevel(paragraph)
			if level > 0 && level <= maxLevel {
				text := d.extractParagraphText(paragraph)
				if text != "" {
					entry := TOCEntry{
						Text:       text,
						Level:      level,
						PageNum:    pageNum,
						BookmarkID: fmt.Sprintf("_Toc_%s", strings.ReplaceAll(text, " ", "_")),
					}
					entries = append(entries, entry)
				}
			}
		}
	}

	return entries
}

// getHeadingLevel 获取段落的标题级别
func (d *Document) getHeadingLevel(paragraph *Paragraph) int {
	if paragraph.Properties != nil && paragraph.Properties.ParagraphStyle != nil {
		styleVal := paragraph.Properties.ParagraphStyle.Val

		// 根据样式ID映射标题级别 - 支持数字ID
		switch styleVal {
		case "1": // heading 1 (有些文档使用1作为标题1)
			return 1
		case "2": // heading 1 (Word默认使用2作为标题1)
			return 1
		case "3": // heading 2
			return 2
		case "4": // heading 3
			return 3
		case "5": // heading 4
			return 4
		case "6": // heading 5
			return 5
		case "7": // heading 6
			return 6
		case "8": // heading 7
			return 7
		case "9": // heading 8
			return 8
		case "10": // heading 9
			return 9
		}

		// 支持标准样式名称匹配
		switch styleVal {
		case "Heading1", "heading1", "Title1":
			return 1
		case "Heading2", "heading2", "Title2":
			return 2
		case "Heading3", "heading3", "Title3":
			return 3
		case "Heading4", "heading4", "Title4":
			return 4
		case "Heading5", "heading5", "Title5":
			return 5
		case "Heading6", "heading6", "Title6":
			return 6
		case "Heading7", "heading7", "Title7":
			return 7
		case "Heading8", "heading8", "Title8":
			return 8
		case "Heading9", "heading9", "Title9":
			return 9
		}

		// 支持通用模式匹配（处理Heading后面跟数字的情况）
		if strings.HasPrefix(strings.ToLower(styleVal), "heading") {
			// 提取数字部分
			numStr := strings.TrimPrefix(strings.ToLower(styleVal), "heading")
			if numStr != "" {
				if level := parseInt(numStr); level >= 1 && level <= 9 {
					return level
				}
			}
		}
	}
	return 0
}

// parseInt 简单的字符串转整数函数
func parseInt(s string) int {
	switch s {
	case "1":
		return 1
	case "2":
		return 2
	case "3":
		return 3
	case "4":
		return 4
	case "5":
		return 5
	case "6":
		return 6
	case "7":
		return 7
	case "8":
		return 8
	case "9":
		return 9
	default:
		return 0
	}
}

// extractParagraphText 提取段落文本
func (d *Document) extractParagraphText(paragraph *Paragraph) string {
	var text strings.Builder
	for _, run := range paragraph.Runs {
		text.WriteString(run.Text.Content)
	}
	return text.String()
}

// insertTOCField 插入目录域
func (d *Document) insertTOCField(config *TOCConfig) error {
	// 构建TOC指令
	instr := fmt.Sprintf("TOC \\o \"1-%d\"", config.MaxLevel)
	if config.UseHyperlink {
		instr += " \\h"
	}
	if !config.ShowPageNum {
		instr += " \\n"
	}

	// 创建目录域段落
	tocPara := &Paragraph{
		Properties: &ParagraphProperties{
			ParagraphStyle: &ParagraphStyle{Val: "TOC1"},
		},
	}

	// 添加域开始
	fieldStart := Run{
		Properties: &RunProperties{},
		Text:       Text{Content: ""}, // 域开始标记
	}

	// 添加域指令
	fieldInstr := Run{
		Properties: &RunProperties{},
		Text:       Text{Content: instr},
	}

	// 添加域结束
	fieldEnd := Run{
		Properties: &RunProperties{},
		Text:       Text{Content: ""}, // 域结束标记
	}

	tocPara.Runs = append(tocPara.Runs, fieldStart, fieldInstr, fieldEnd)
	d.Body.Elements = append(d.Body.Elements, tocPara)

	return nil
}

// addTOCEntry 添加目录条目
func (d *Document) addTOCEntry(entry TOCEntry, config *TOCConfig) error {
	// 创建目录条目段落
	entryPara := &Paragraph{
		Properties: &ParagraphProperties{
			ParagraphStyle: &ParagraphStyle{Val: fmt.Sprintf("TOC%d", entry.Level)},
		},
	}

	if config.UseHyperlink {
		// 创建超链接
		hyperlink := &Hyperlink{
			Anchor: entry.BookmarkID,
		}

		// 标题文本
		titleRun := Run{
			Properties: &RunProperties{},
			Text:       Text{Content: entry.Text},
		}
		hyperlink.Runs = append(hyperlink.Runs, titleRun)

		// 如果显示页码，添加引导线和页码
		if config.ShowPageNum {
			if config.DotLeader {
				// 添加点状引导线
				leaderRun := Run{
					Properties: &RunProperties{},
					Text:       Text{Content: strings.Repeat(".", 20)}, // 简化处理
				}
				hyperlink.Runs = append(hyperlink.Runs, leaderRun)
			}

			// 添加页码
			pageRun := Run{
				Properties: &RunProperties{},
				Text:       Text{Content: fmt.Sprintf("%d", entry.PageNum)},
			}
			hyperlink.Runs = append(hyperlink.Runs, pageRun)
		}

		// 将超链接添加到段落中
		// 这里需要特殊处理，因为Hyperlink不是标准的Run
		// 简化处理，直接作为文本添加
		hyperlinkRun := Run{
			Properties: &RunProperties{},
			Text:       Text{Content: entry.Text},
		}
		entryPara.Runs = append(entryPara.Runs, hyperlinkRun)

		if config.ShowPageNum {
			pageRun := Run{
				Properties: &RunProperties{},
				Text:       Text{Content: fmt.Sprintf("\t%d", entry.PageNum)},
			}
			entryPara.Runs = append(entryPara.Runs, pageRun)
		}
	} else {
		// 不使用超链接的简单文本
		titleRun := Run{
			Properties: &RunProperties{},
			Text:       Text{Content: entry.Text},
		}
		entryPara.Runs = append(entryPara.Runs, titleRun)

		if config.ShowPageNum {
			pageRun := Run{
				Properties: &RunProperties{},
				Text:       Text{Content: fmt.Sprintf("\t%d", entry.PageNum)},
			}
			entryPara.Runs = append(entryPara.Runs, pageRun)
		}
	}

	d.Body.Elements = append(d.Body.Elements, entryPara)
	return nil
}

// findTOCStart 查找目录开始位置
func (d *Document) findTOCStart() int {
	for i, element := range d.Body.Elements {
		if paragraph, ok := element.(*Paragraph); ok {
			if paragraph.Properties != nil && paragraph.Properties.ParagraphStyle != nil {
				if strings.HasPrefix(paragraph.Properties.ParagraphStyle.Val, "TOC") {
					return i
				}
			}
		}
	}
	return -1
}

// findTOCSDT 查找目录SDT结构
func (d *Document) findTOCSDT() (*SDT, int) {
	for i, element := range d.Body.Elements {
		if sdt, ok := element.(*SDT); ok {
			// 检查是否是目录SDT
			if sdt.Properties != nil && sdt.Properties.DocPartObj != nil {
				if sdt.Properties.DocPartObj.DocPartGallery != nil {
					if sdt.Properties.DocPartObj.DocPartGallery.Val == "Table of Contents" {
						return sdt, i
					}
				}
			}
		}
	}
	return nil, -1
}

// removeTOCEntries 删除现有目录条目
func (d *Document) removeTOCEntries(startIndex int) {
	// 简化处理：从startIndex开始查找并删除所有TOC样式的段落
	var newElements []interface{}

	// 保留start之前的元素
	newElements = append(newElements, d.Body.Elements[:startIndex]...)

	// 跳过TOC相关的元素
	for i := startIndex; i < len(d.Body.Elements); i++ {
		element := d.Body.Elements[i]
		if paragraph, ok := element.(*Paragraph); ok {
			if paragraph.Properties != nil && paragraph.Properties.ParagraphStyle != nil {
				if !strings.HasPrefix(paragraph.Properties.ParagraphStyle.Val, "TOC") {
					// 不是TOC样式，保留后续所有元素
					newElements = append(newElements, d.Body.Elements[i:]...)
					break
				}
			}
		}
	}

	d.Body.Elements = newElements
}

// SetTOCStyle 设置目录样式
func (d *Document) SetTOCStyle(level int, style *TextFormat) error {
	if level < 1 || level > 9 {
		return fmt.Errorf("目录级别必须在1-9之间")
	}

	styleName := fmt.Sprintf("TOC%d", level)

	// 通过样式管理器设置目录样式
	styleManager := d.GetStyleManager()

	// 创建段落样式（这里需要与样式系统集成）
	// 简化处理，实际需要创建完整的样式定义
	_ = styleManager
	_ = styleName
	_ = style

	return nil
}

// AutoGenerateTOC 自动生成目录，检测现有文档中的标题
func (d *Document) AutoGenerateTOC(config *TOCConfig) error {
	if config == nil {
		config = DefaultTOCConfig()
	}

	// 查找现有目录位置
	tocStart := d.findTOCStart()
	var insertIndex int

	if tocStart != -1 {
		// 如果已有目录，删除现有目录条目
		d.removeTOCEntries(tocStart)
		insertIndex = tocStart
	} else {
		// 如果没有目录，在文档开头插入
		insertIndex = 0
	}

	// 收集文档中的所有标题并为它们添加书签
	entries := d.collectHeadingsAndAddBookmarks(config.MaxLevel)

	if len(entries) == 0 {
		return fmt.Errorf("文档中未找到标题（样式ID为2-10的段落）")
	}

	// 使用真正的Word域字段生成目录，而不是简化的SDT
	tocElements := d.createWordFieldTOC(config, entries)

	// 将目录插入到指定位置
	if insertIndex == 0 {
		// 在开头插入
		d.Body.Elements = append(tocElements, d.Body.Elements...)
	} else {
		// 在指定位置替换
		newElements := make([]interface{}, 0, len(d.Body.Elements)+len(tocElements))
		newElements = append(newElements, d.Body.Elements[:insertIndex]...)
		newElements = append(newElements, tocElements...)
		newElements = append(newElements, d.Body.Elements[insertIndex:]...)
		d.Body.Elements = newElements
	}

	return nil
}

// GetHeadingCount 获取文档中标题的数量，用于调试
func (d *Document) GetHeadingCount() map[int]int {
	counts := make(map[int]int)

	for _, element := range d.Body.Elements {
		if paragraph, ok := element.(*Paragraph); ok {
			level := d.getHeadingLevel(paragraph)
			if level > 0 {
				counts[level]++
			}
		}
	}

	return counts
}

// ListHeadings 列出文档中所有的标题，用于调试
func (d *Document) ListHeadings() []TOCEntry {
	return d.collectHeadings(9) // 获取所有级别的标题
}

// createWordFieldTOC 创建使用真正Word域字段的目录
func (d *Document) createWordFieldTOC(config *TOCConfig, entries []TOCEntry) []interface{} {
	var elements []interface{}

	// 创建目录SDT容器
	tocSDT := &SDT{
		Properties: &SDTProperties{
			RunPr: &RunProperties{
				FontFamily: &FontFamily{ASCII: "宋体", HAnsi: "宋体", EastAsia: "宋体", CS: "Times New Roman"},
				FontSize:   &FontSize{Val: "21"},
			},
			ID:    &SDTID{Val: "147458718"},
			Color: &SDTColor{Val: "DBDBDB"},
			DocPartObj: &DocPartObj{
				DocPartGallery: &DocPartGallery{Val: "Table of Contents"},
				DocPartUnique:  &DocPartUnique{},
			},
		},
		EndPr: &SDTEndPr{
			RunPr: &RunProperties{
				FontFamily: &FontFamily{ASCII: "Calibri", HAnsi: "Calibri", EastAsia: "宋体", CS: "Times New Roman"},
				Bold:       &Bold{},
				Color:      &Color{Val: "2F5496"},
				FontSize:   &FontSize{Val: "32"},
			},
		},
		Content: &SDTContent{
			Elements: []interface{}{},
		},
	}

	// 添加目录标题段落
	titlePara := &Paragraph{
		Properties: &ParagraphProperties{
			Spacing: &Spacing{
				Before: "0",
				After:  "0",
				Line:   "240",
			},
			Justification: &Justification{Val: "center"},
			Indentation: &Indentation{
				Left:      "0",
				Right:     "0",
				FirstLine: "0",
			},
		},
		Runs: []Run{
			{
				Text: Text{Content: config.Title},
				Properties: &RunProperties{
					FontFamily: &FontFamily{ASCII: "宋体"},
					FontSize:   &FontSize{Val: "21"},
				},
			},
		},
	}

	tocSDT.Content.Elements = append(tocSDT.Content.Elements, titlePara)

	// 创建主TOC域段落
	tocFieldPara := &Paragraph{
		Properties: &ParagraphProperties{
			ParagraphStyle: &ParagraphStyle{Val: "12"}, // TOC样式
			Tabs: &Tabs{
				Tabs: []TabDef{
					{
						Val:    "right",
						Leader: "dot",
						Pos:    "8640",
					},
				},
			},
		},
		Runs: []Run{},
	}

	// 添加TOC域开始
	tocFieldPara.Runs = append(tocFieldPara.Runs, Run{
		Properties: &RunProperties{
			Bold:     &Bold{},
			Color:    &Color{Val: "2F5496"},
			FontSize: &FontSize{Val: "32"},
		},
		FieldChar: &FieldChar{
			FieldCharType: "begin",
		},
	})

	// 添加TOC指令
	instrContent := fmt.Sprintf("TOC \\o \"1-%d\" \\h \\u", config.MaxLevel)
	tocFieldPara.Runs = append(tocFieldPara.Runs, Run{
		Properties: &RunProperties{
			Bold:     &Bold{},
			Color:    &Color{Val: "2F5496"},
			FontSize: &FontSize{Val: "32"},
		},
		InstrText: &InstrText{
			Space:   "preserve",
			Content: instrContent,
		},
	})

	// 添加TOC域分隔符
	tocFieldPara.Runs = append(tocFieldPara.Runs, Run{
		Properties: &RunProperties{
			Bold:     &Bold{},
			Color:    &Color{Val: "2F5496"},
			FontSize: &FontSize{Val: "32"},
		},
		FieldChar: &FieldChar{
			FieldCharType: "separate",
		},
	})

	tocSDT.Content.Elements = append(tocSDT.Content.Elements, tocFieldPara)

	// 为每个条目创建超链接段落
	for _, entry := range entries {
		entryPara := d.createTOCEntryWithFields(entry, config)
		tocSDT.Content.Elements = append(tocSDT.Content.Elements, entryPara)
	}

	// 添加TOC域结束段落
	endPara := &Paragraph{
		Properties: &ParagraphProperties{
			ParagraphStyle: &ParagraphStyle{Val: "2"},
			Spacing: &Spacing{
				Before: "240",
				After:  "0",
			},
		},
		Runs: []Run{
			{
				Properties: &RunProperties{
					Color: &Color{Val: "2F5496"},
				},
				FieldChar: &FieldChar{
					FieldCharType: "end",
				},
			},
		},
	}

	tocSDT.Content.Elements = append(tocSDT.Content.Elements, endPara)
	elements = append(elements, tocSDT)

	return elements
}

// createTOCEntryWithFields 创建带域字段的目录条目
func (d *Document) createTOCEntryWithFields(entry TOCEntry, config *TOCConfig) *Paragraph {
	// 确定目录样式ID
	var styleVal string
	switch entry.Level {
	case 1:
		styleVal = "13" // TOC 1
	case 2:
		styleVal = "14" // TOC 2
	case 3:
		styleVal = "15" // TOC 3
	default:
		styleVal = fmt.Sprintf("%d", 12+entry.Level)
	}

	para := &Paragraph{
		Properties: &ParagraphProperties{
			ParagraphStyle: &ParagraphStyle{Val: styleVal},
			Tabs: &Tabs{
				Tabs: []TabDef{
					{
						Val:    "right",
						Leader: "dot",
						Pos:    "8640",
					},
				},
			},
		},
		Runs: []Run{},
	}

	// 为每个条目生成唯一的书签ID
	anchor := fmt.Sprintf("_Toc%d", generateUniqueID(entry.Text))

	// 创建超链接域开始
	para.Runs = append(para.Runs, Run{
		Properties: &RunProperties{
			Color: &Color{Val: "2F5496"},
		},
		FieldChar: &FieldChar{
			FieldCharType: "begin",
		},
	})

	// 添加超链接指令
	para.Runs = append(para.Runs, Run{
		InstrText: &InstrText{
			Space:   "preserve",
			Content: fmt.Sprintf(" HYPERLINK \\l %s ", anchor),
		},
	})

	// 超链接域分隔符
	para.Runs = append(para.Runs, Run{
		FieldChar: &FieldChar{
			FieldCharType: "separate",
		},
	})

	// 添加标题文本
	para.Runs = append(para.Runs, Run{
		Text: Text{Content: entry.Text},
	})

	// 添加制表符
	para.Runs = append(para.Runs, Run{
		Text: Text{Content: "\t"},
	})

	// 添加页码引用域
	para.Runs = append(para.Runs, Run{
		FieldChar: &FieldChar{
			FieldCharType: "begin",
		},
	})

	para.Runs = append(para.Runs, Run{
		InstrText: &InstrText{
			Space:   "preserve",
			Content: fmt.Sprintf(" PAGEREF %s \\h ", anchor),
		},
	})

	para.Runs = append(para.Runs, Run{
		FieldChar: &FieldChar{
			FieldCharType: "separate",
		},
	})

	// 页码文本
	para.Runs = append(para.Runs, Run{
		Text: Text{Content: fmt.Sprintf("%d", entry.PageNum)},
	})

	// 页码域结束
	para.Runs = append(para.Runs, Run{
		FieldChar: &FieldChar{
			FieldCharType: "end",
		},
	})

	// 超链接域结束
	para.Runs = append(para.Runs, Run{
		Properties: &RunProperties{
			Color: &Color{Val: "2F5496"},
		},
		FieldChar: &FieldChar{
			FieldCharType: "end",
		},
	})

	return para
}

// generateUniqueID 基于文本内容生成唯一ID
func generateUniqueID(text string) int {
	// 使用简单的哈希算法生成唯一ID
	hash := 0
	for _, char := range text {
		hash = hash*31 + int(char)
	}
	// 确保是正数并限制在合理范围内
	if hash < 0 {
		hash = -hash
	}
	return (hash % 90000) + 10000 // 生成10000-99999之间的数字
}

// collectHeadingsAndAddBookmarks 收集标题信息并添加书签
func (d *Document) collectHeadingsAndAddBookmarks(maxLevel int) []TOCEntry {
	var entries []TOCEntry
	pageNum := 1 // 简化处理，实际需要计算真实页码

	// 需要一个新的Elements切片来插入书签
	newElements := make([]interface{}, 0, len(d.Body.Elements)*2)
	entryIndex := 0

	for _, element := range d.Body.Elements {
		if paragraph, ok := element.(*Paragraph); ok {
			level := d.getHeadingLevel(paragraph)
			if level > 0 && level <= maxLevel {
				text := d.extractParagraphText(paragraph)
				if text != "" {
					// 为每个条目生成唯一的书签ID（与目录条目中使用的一致）
					anchor := fmt.Sprintf("_Toc%d", generateUniqueID(text))

					entry := TOCEntry{
						Text:       text,
						Level:      level,
						PageNum:    pageNum,
						BookmarkID: anchor,
					}
					entries = append(entries, entry)

					// 在标题段落前添加书签开始标记
					bookmarkStart := &BookmarkStart{
						ID:   fmt.Sprintf("%d", entryIndex),
						Name: anchor,
					}
					newElements = append(newElements, bookmarkStart)

					// 添加原段落
					newElements = append(newElements, element)

					// 在标题段落后添加书签结束标记
					bookmarkEnd := &BookmarkEnd{
						ID: fmt.Sprintf("%d", entryIndex),
					}
					newElements = append(newElements, bookmarkEnd)

					entryIndex++
					continue
				}
			}
		}
		// 非标题段落直接添加
		newElements = append(newElements, element)
	}

	// 更新文档元素
	d.Body.Elements = newElements

	return entries
}
