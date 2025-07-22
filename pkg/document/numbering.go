// Package document 提供Word文档列表和编号操作功能
package document

import (
	"encoding/xml"
	"fmt"
	"strconv"
)

// ListType 列表类型
type ListType string

const (
	// ListTypeBullet 无序列表（项目符号）
	ListTypeBullet ListType = "bullet"
	// ListTypeNumber 有序列表（数字编号）
	ListTypeNumber ListType = "number"
	// ListTypeDecimal 十进制编号
	ListTypeDecimal ListType = "decimal"
	// ListTypeLowerLetter 小写字母编号
	ListTypeLowerLetter ListType = "lowerLetter"
	// ListTypeUpperLetter 大写字母编号
	ListTypeUpperLetter ListType = "upperLetter"
	// ListTypeLowerRoman 小写罗马数字
	ListTypeLowerRoman ListType = "lowerRoman"
	// ListTypeUpperRoman 大写罗马数字
	ListTypeUpperRoman ListType = "upperRoman"
)

// BulletType 项目符号类型
type BulletType string

const (
	// BulletTypeDot 圆点符号
	BulletTypeDot BulletType = "•"
	// BulletTypeCircle 空心圆
	BulletTypeCircle BulletType = "○"
	// BulletTypeSquare 方块
	BulletTypeSquare BulletType = "■"
	// BulletTypeDash 短横线
	BulletTypeDash BulletType = "–"
	// BulletTypeArrow 箭头
	BulletTypeArrow BulletType = "→"
)

// Numbering 编号定义
type Numbering struct {
	XMLName            xml.Name       `xml:"w:numbering"`
	Xmlns              string         `xml:"xmlns:w,attr"`
	AbstractNums       []*AbstractNum `xml:"w:abstractNum"`
	NumberingInstances []*NumInstance `xml:"w:num"`
}

// AbstractNum 抽象编号定义
type AbstractNum struct {
	XMLName       xml.Name `xml:"w:abstractNum"`
	AbstractNumID string   `xml:"w:abstractNumId,attr"`
	Levels        []*Level `xml:"w:lvl"`
}

// NumInstance 编号实例
type NumInstance struct {
	XMLName       xml.Name              `xml:"w:num"`
	NumID         string                `xml:"w:numId,attr"`
	AbstractNumID *AbstractNumReference `xml:"w:abstractNumId"`
}

// AbstractNumReference 抽象编号引用
type AbstractNumReference struct {
	XMLName xml.Name `xml:"w:abstractNumId"`
	Val     string   `xml:"w:val,attr"`
}

// Level 编号级别
type Level struct {
	XMLName   xml.Name   `xml:"w:lvl"`
	ILevel    string     `xml:"w:ilvl,attr"`
	Start     *Start     `xml:"w:start,omitempty"`
	NumFmt    *NumFmt    `xml:"w:numFmt,omitempty"`
	LevelText *LevelText `xml:"w:lvlText,omitempty"`
	LevelJc   *LevelJc   `xml:"w:lvlJc,omitempty"`
	PPr       *LevelPPr  `xml:"w:pPr,omitempty"`
	RPr       *LevelRPr  `xml:"w:rPr,omitempty"`
}

// Start 起始编号
type Start struct {
	XMLName xml.Name `xml:"w:start"`
	Val     string   `xml:"w:val,attr"`
}

// NumFmt 编号格式
type NumFmt struct {
	XMLName xml.Name `xml:"w:numFmt"`
	Val     string   `xml:"w:val,attr"`
}

// LevelText 级别文本
type LevelText struct {
	XMLName xml.Name `xml:"w:lvlText"`
	Val     string   `xml:"w:val,attr"`
}

// LevelJc 级别对齐
type LevelJc struct {
	XMLName xml.Name `xml:"w:lvlJc"`
	Val     string   `xml:"w:val,attr"`
}

// LevelPPr 级别段落属性
type LevelPPr struct {
	XMLName xml.Name     `xml:"w:pPr"`
	Ind     *LevelIndent `xml:"w:ind,omitempty"`
}

// LevelIndent 级别缩进
type LevelIndent struct {
	XMLName xml.Name `xml:"w:ind"`
	Left    string   `xml:"w:left,attr,omitempty"`
	Hanging string   `xml:"w:hanging,attr,omitempty"`
}

// LevelRPr 级别文本属性
type LevelRPr struct {
	XMLName    xml.Name    `xml:"w:rPr"`
	FontFamily *FontFamily `xml:"w:rFonts,omitempty"`
}

// ListConfig 列表配置
type ListConfig struct {
	Type         ListType   // 列表类型
	BulletSymbol BulletType // 项目符号（仅用于无序列表）
	StartNumber  int        // 起始编号（仅用于有序列表）
	IndentLevel  int        // 缩进级别（0-8）
}

// 全局编号管理器
var globalNumberingManager *NumberingManager

// NumberingManager 编号管理器
type NumberingManager struct {
	nextAbstractNumID int
	nextNumID         int
	abstractNums      map[string]*AbstractNum
	numInstances      map[string]*NumInstance
}

// getNumberingManager 获取全局编号管理器
func getNumberingManager() *NumberingManager {
	if globalNumberingManager == nil {
		globalNumberingManager = &NumberingManager{
			nextAbstractNumID: 0,
			nextNumID:         1,
			abstractNums:      make(map[string]*AbstractNum),
			numInstances:      make(map[string]*NumInstance),
		}
	}
	return globalNumberingManager
}

// AddListItem 添加列表项
func (d *Document) AddListItem(text string, config *ListConfig) *Paragraph {
	if config == nil {
		config = &ListConfig{
			Type:         ListTypeBullet,
			BulletSymbol: BulletTypeDot,
			IndentLevel:  0,
		}
	}

	// 确保编号管理器已初始化
	d.ensureNumberingInitialized()

	// 获取或创建编号定义
	numID := d.getOrCreateNumbering(config)

	// 创建段落
	paragraph := &Paragraph{
		Properties: &ParagraphProperties{
			NumberingProperties: &NumberingProperties{
				ILevel: &ILevel{Val: strconv.Itoa(config.IndentLevel)},
				NumID:  &NumID{Val: numID},
			},
		},
	}

	// 添加文本内容
	if text != "" {
		run := Run{
			Text: Text{
				Content: text,
			},
		}
		paragraph.Runs = append(paragraph.Runs, run)
	}

	// 添加到文档
	d.Body.Elements = append(d.Body.Elements, paragraph)
	return paragraph
}

// AddBulletList 添加无序列表项
func (d *Document) AddBulletList(text string, level int, bulletType BulletType) *Paragraph {
	config := &ListConfig{
		Type:         ListTypeBullet,
		BulletSymbol: bulletType,
		IndentLevel:  level,
	}
	return d.AddListItem(text, config)
}

// AddNumberedList 添加有序列表项
func (d *Document) AddNumberedList(text string, level int, numType ListType) *Paragraph {
	config := &ListConfig{
		Type:        numType,
		IndentLevel: level,
		StartNumber: 1,
	}
	return d.AddListItem(text, config)
}

// CreateMultiLevelList 创建多级列表
func (d *Document) CreateMultiLevelList(items []ListItem) error {
	for _, item := range items {
		config := &ListConfig{
			Type:         item.Type,
			BulletSymbol: item.BulletSymbol,
			IndentLevel:  item.Level,
			StartNumber:  item.StartNumber,
		}
		d.AddListItem(item.Text, config)
	}
	return nil
}

// ListItem 列表项结构
type ListItem struct {
	Text         string     // 文本内容
	Level        int        // 缩进级别
	Type         ListType   // 列表类型
	BulletSymbol BulletType // 项目符号
	StartNumber  int        // 起始编号
}

// ensureNumberingInitialized 确保编号系统已初始化
func (d *Document) ensureNumberingInitialized() {
	// 检查是否已有编号定义
	if _, exists := d.parts["word/numbering.xml"]; !exists {
		d.initializeNumbering()
	}
}

// initializeNumbering 初始化编号系统
func (d *Document) initializeNumbering() {
	numbering := &Numbering{
		Xmlns:              "http://schemas.openxmlformats.org/wordprocessingml/2006/main",
		AbstractNums:       []*AbstractNum{},
		NumberingInstances: []*NumInstance{},
	}

	// 序列化编号定义
	numberingXML, err := xml.MarshalIndent(numbering, "", "  ")
	if err != nil {
		return
	}

	// 添加XML声明
	xmlDeclaration := []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + "\n")
	d.parts["word/numbering.xml"] = append(xmlDeclaration, numberingXML...)

	// 添加内容类型
	d.addContentType("word/numbering.xml", "application/vnd.openxmlformats-officedocument.wordprocessingml.numbering+xml")

	// 添加关系
	d.addNumberingRelationship()
}

// getOrCreateNumbering 获取或创建编号定义
func (d *Document) getOrCreateNumbering(config *ListConfig) string {
	manager := getNumberingManager()

	// 生成抽象编号键
	abstractKey := fmt.Sprintf("%s_%s_%d", config.Type, config.BulletSymbol, config.IndentLevel)

	// 检查是否已存在抽象编号
	var abstractNum *AbstractNum
	if existing, exists := manager.abstractNums[abstractKey]; exists {
		abstractNum = existing
	} else {
		// 创建新的抽象编号
		abstractNumID := strconv.Itoa(manager.nextAbstractNumID)
		manager.nextAbstractNumID++

		abstractNum = d.createAbstractNum(abstractNumID, config)
		manager.abstractNums[abstractKey] = abstractNum
	}

	// 创建编号实例
	numID := strconv.Itoa(manager.nextNumID)
	manager.nextNumID++

	numInstance := &NumInstance{
		NumID: numID,
		AbstractNumID: &AbstractNumReference{
			Val: abstractNum.AbstractNumID,
		},
	}
	manager.numInstances[numID] = numInstance

	// 更新编号定义文件
	d.updateNumberingFile()

	return numID
}

// createAbstractNum 创建抽象编号定义
func (d *Document) createAbstractNum(abstractNumID string, config *ListConfig) *AbstractNum {
	abstractNum := &AbstractNum{
		AbstractNumID: abstractNumID,
		Levels:        []*Level{},
	}

	// 创建多个级别（支持9级列表）
	for i := 0; i <= 8; i++ {
		level := d.createLevel(i, config)
		abstractNum.Levels = append(abstractNum.Levels, level)
	}

	return abstractNum
}

// createLevel 创建编号级别
func (d *Document) createLevel(levelIndex int, config *ListConfig) *Level {
	level := &Level{
		ILevel:  strconv.Itoa(levelIndex),
		Start:   &Start{Val: strconv.Itoa(config.StartNumber)},
		LevelJc: &LevelJc{Val: "left"},
		PPr: &LevelPPr{
			Ind: &LevelIndent{
				Left:    strconv.Itoa((levelIndex + 1) * 720), // 720 twips = 0.5 inch
				Hanging: "360",                                // 360 twips = 0.25 inch
			},
		},
	}

	// 设置编号格式和文本
	switch config.Type {
	case ListTypeBullet:
		level.NumFmt = &NumFmt{Val: "bullet"}
		level.LevelText = &LevelText{Val: string(config.BulletSymbol)}
		level.RPr = &LevelRPr{
			FontFamily: &FontFamily{ASCII: "Symbol"},
		}
	case ListTypeNumber, ListTypeDecimal:
		level.NumFmt = &NumFmt{Val: "decimal"}
		level.LevelText = &LevelText{Val: fmt.Sprintf("%%%d.", levelIndex+1)}
	case ListTypeLowerLetter:
		level.NumFmt = &NumFmt{Val: "lowerLetter"}
		level.LevelText = &LevelText{Val: fmt.Sprintf("%%%d.", levelIndex+1)}
	case ListTypeUpperLetter:
		level.NumFmt = &NumFmt{Val: "upperLetter"}
		level.LevelText = &LevelText{Val: fmt.Sprintf("%%%d.", levelIndex+1)}
	case ListTypeLowerRoman:
		level.NumFmt = &NumFmt{Val: "lowerRoman"}
		level.LevelText = &LevelText{Val: fmt.Sprintf("%%%d.", levelIndex+1)}
	case ListTypeUpperRoman:
		level.NumFmt = &NumFmt{Val: "upperRoman"}
		level.LevelText = &LevelText{Val: fmt.Sprintf("%%%d.", levelIndex+1)}
	}

	return level
}

// updateNumberingFile 更新编号定义文件
func (d *Document) updateNumberingFile() {
	manager := getNumberingManager()

	numbering := &Numbering{
		Xmlns:              "http://schemas.openxmlformats.org/wordprocessingml/2006/main",
		AbstractNums:       []*AbstractNum{},
		NumberingInstances: []*NumInstance{},
	}

	// 添加所有抽象编号
	for _, abstractNum := range manager.abstractNums {
		numbering.AbstractNums = append(numbering.AbstractNums, abstractNum)
	}

	// 添加所有编号实例
	for _, numInstance := range manager.numInstances {
		numbering.NumberingInstances = append(numbering.NumberingInstances, numInstance)
	}

	// 序列化
	numberingXML, err := xml.MarshalIndent(numbering, "", "  ")
	if err != nil {
		return
	}

	// 添加XML声明
	xmlDeclaration := []byte(`<?xml version="1.0" encoding="UTF-8" standalone="yes"?>` + "\n")
	d.parts["word/numbering.xml"] = append(xmlDeclaration, numberingXML...)
}

// addNumberingRelationship 添加编号关系
func (d *Document) addNumberingRelationship() {
	// 生成关系ID
	relationshipID := fmt.Sprintf("rId%d", len(d.documentRelationships.Relationships)+2) // +2 因为已有样式styles.xml定义

	// 添加关系
	relationship := Relationship{
		ID:     relationshipID,
		Type:   "http://schemas.openxmlformats.org/officeDocument/2006/relationships/numbering",
		Target: "numbering.xml",
	}
	d.documentRelationships.Relationships = append(d.documentRelationships.Relationships, relationship)
}

// RestartNumbering 重新开始编号
func (d *Document) RestartNumbering(numID string) {
	// 重置编号计数器
	// 在实际实现中，需要创建新的编号实例来重置计数
	manager := getNumberingManager()

	// 创建新的编号实例
	newNumID := strconv.Itoa(manager.nextNumID)
	manager.nextNumID++

	// 如果存在原有实例，复制其抽象编号引用
	if existing, exists := manager.numInstances[numID]; exists {
		newInstance := &NumInstance{
			NumID: newNumID,
			AbstractNumID: &AbstractNumReference{
				Val: existing.AbstractNumID.Val,
			},
		}
		manager.numInstances[newNumID] = newInstance
		d.updateNumberingFile()
	}
}
