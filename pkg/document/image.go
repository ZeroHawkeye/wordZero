// Package document 提供Word文档的图片操作功能
package document

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"image"
	"image/gif"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strconv"
)

// ImageFormat 图片格式类型
type ImageFormat string

const (
	// 支持的图片格式
	ImageFormatJPEG ImageFormat = "jpeg"
	ImageFormatPNG  ImageFormat = "png"
	ImageFormatGIF  ImageFormat = "gif"
)

// ImagePosition 图片位置类型
type ImagePosition string

const (
	// 图片位置选项
	ImagePositionInline     ImagePosition = "inline"     // 嵌入式（默认）
	ImagePositionFloatLeft  ImagePosition = "floatLeft"  // 左浮动
	ImagePositionFloatRight ImagePosition = "floatRight" // 右浮动
)

// ImageWrapText 文字环绕类型
type ImageWrapText string

const (
	// 文字环绕选项
	ImageWrapNone         ImageWrapText = "none"         // 无环绕
	ImageWrapSquare       ImageWrapText = "square"       // 四周环绕
	ImageWrapTight        ImageWrapText = "tight"        // 紧密环绕
	ImageWrapTopAndBottom ImageWrapText = "topAndBottom" // 上下环绕
)

// ImageSize 图片大小配置
type ImageSize struct {
	Width  float64 // 宽度（毫米）
	Height float64 // 高度（毫米）
	// 是否保持长宽比（当只设置一个维度时）
	KeepAspectRatio bool
}

// ImageConfig 图片配置
type ImageConfig struct {
	// 图片大小
	Size *ImageSize
	// 图片位置
	Position ImagePosition
	// 图片对齐方式（用于嵌入式图片）
	Alignment AlignmentType
	// 文字环绕
	WrapText ImageWrapText
	// 图片描述（替代文字）
	AltText string
	// 图片标题
	Title string
	// 水平偏移（毫米）
	OffsetX float64
	// 垂直偏移（毫米）
	OffsetY float64
}

// ImageInfo 图片信息
type ImageInfo struct {
	ID         string       // 图片ID
	RelationID string       // 关系ID
	Format     ImageFormat  // 图片格式
	Width      int          // 原始宽度（像素）
	Height     int          // 原始高度（像素）
	Data       []byte       // 图片数据
	Config     *ImageConfig // 图片配置
}

// DrawingElement 绘图元素（包含图片）
type DrawingElement struct {
	XMLName xml.Name       `xml:"w:drawing"`
	Inline  *InlineDrawing `xml:"wp:inline,omitempty"`
	Anchor  *AnchorDrawing `xml:"wp:anchor,omitempty"`
}

// InlineDrawing 嵌入式绘图
type InlineDrawing struct {
	XMLName xml.Name        `xml:"wp:inline"`
	DistT   string          `xml:"distT,attr,omitempty"`
	DistB   string          `xml:"distB,attr,omitempty"`
	DistL   string          `xml:"distL,attr,omitempty"`
	DistR   string          `xml:"distR,attr,omitempty"`
	Extent  *DrawingExtent  `xml:"wp:extent"`
	DocPr   *DrawingDocPr   `xml:"wp:docPr"`
	Graphic *DrawingGraphic `xml:"a:graphic"`
}

// AnchorDrawing 浮动绘图
type AnchorDrawing struct {
	XMLName           xml.Name            `xml:"wp:anchor"`
	DistT             string              `xml:"distT,attr,omitempty"`
	DistB             string              `xml:"distB,attr,omitempty"`
	DistL             string              `xml:"distL,attr,omitempty"`
	DistR             string              `xml:"distR,attr,omitempty"`
	SimplePos         string              `xml:"simplePos,attr,omitempty"`
	RelativeHeight    string              `xml:"relativeHeight,attr,omitempty"`
	BehindDoc         string              `xml:"behindDoc,attr,omitempty"`
	Locked            string              `xml:"locked,attr,omitempty"`
	LayoutInCell      string              `xml:"layoutInCell,attr,omitempty"`
	AllowOverlap      string              `xml:"allowOverlap,attr,omitempty"`
	SimplePosition    *SimplePosition     `xml:"wp:simplePos,omitempty"`
	PositionH         *HorizontalPosition `xml:"wp:positionH,omitempty"`
	PositionV         *VerticalPosition   `xml:"wp:positionV,omitempty"`
	Extent            *DrawingExtent      `xml:"wp:extent"`
	EffectExtent      *EffectExtent       `xml:"wp:effectExtent,omitempty"`
	WrapNone          *WrapNone           `xml:"wp:wrapNone,omitempty"`
	WrapSquare        *WrapSquare         `xml:"wp:wrapSquare,omitempty"`
	WrapTight         *WrapTight          `xml:"wp:wrapTight,omitempty"`
	WrapThrough       *WrapThrough        `xml:"wp:wrapThrough,omitempty"`
	WrapTopAndBottom  *WrapTopAndBottom   `xml:"wp:wrapTopAndBottom,omitempty"`
	DocPr             *DrawingDocPr       `xml:"wp:docPr"`
	CNvGraphicFramePr *CNvGraphicFramePr  `xml:"wp:cNvGraphicFramePr,omitempty"`
	Graphic           *DrawingGraphic     `xml:"a:graphic"`
}

// SimplePosition 简单位置
type SimplePosition struct {
	XMLName xml.Name `xml:"wp:simplePos"`
	X       string   `xml:"x,attr"`
	Y       string   `xml:"y,attr"`
}

// HorizontalPosition 水平位置
type HorizontalPosition struct {
	XMLName      xml.Name   `xml:"wp:positionH"`
	RelativeFrom string     `xml:"relativeFrom,attr"`
	Align        *PosAlign  `xml:"wp:align,omitempty"`
	PosOffset    *PosOffset `xml:"wp:posOffset,omitempty"`
}

// VerticalPosition 垂直位置
type VerticalPosition struct {
	XMLName      xml.Name   `xml:"wp:positionV"`
	RelativeFrom string     `xml:"relativeFrom,attr"`
	Align        *PosAlign  `xml:"wp:align,omitempty"`
	PosOffset    *PosOffset `xml:"wp:posOffset,omitempty"`
}

// PosAlign 位置对齐
type PosAlign struct {
	XMLName xml.Name `xml:"wp:align"`
	Value   string   `xml:",chardata"`
}

// PosOffset 位置偏移
type PosOffset struct {
	XMLName xml.Name `xml:"wp:posOffset"`
	Value   string   `xml:",chardata"`
}

// EffectExtent 效果范围
type EffectExtent struct {
	XMLName xml.Name `xml:"wp:effectExtent"`
	L       string   `xml:"l,attr,omitempty"`
	T       string   `xml:"t,attr,omitempty"`
	R       string   `xml:"r,attr,omitempty"`
	B       string   `xml:"b,attr,omitempty"`
}

// WrapNone 无环绕
type WrapNone struct {
	XMLName xml.Name `xml:"wp:wrapNone"`
}

// WrapSquare 四周环绕
type WrapSquare struct {
	XMLName  xml.Name `xml:"wp:wrapSquare"`
	WrapText string   `xml:"wrapText,attr,omitempty"`
	DistT    string   `xml:"distT,attr,omitempty"`
	DistB    string   `xml:"distB,attr,omitempty"`
	DistL    string   `xml:"distL,attr,omitempty"`
	DistR    string   `xml:"distR,attr,omitempty"`
}

// WrapTight 紧密环绕
type WrapTight struct {
	XMLName     xml.Name     `xml:"wp:wrapTight"`
	WrapText    string       `xml:"wrapText,attr,omitempty"`
	DistL       string       `xml:"distL,attr,omitempty"`
	DistR       string       `xml:"distR,attr,omitempty"`
	WrapPolygon *WrapPolygon `xml:"wp:wrapPolygon,omitempty"`
}

// WrapThrough 穿透环绕
type WrapThrough struct {
	XMLName     xml.Name     `xml:"wp:wrapThrough"`
	WrapText    string       `xml:"wrapText,attr,omitempty"`
	DistL       string       `xml:"distL,attr,omitempty"`
	DistR       string       `xml:"distR,attr,omitempty"`
	WrapPolygon *WrapPolygon `xml:"wp:wrapPolygon,omitempty"`
}

// WrapTopAndBottom 上下环绕
type WrapTopAndBottom struct {
	XMLName      xml.Name      `xml:"wp:wrapTopAndBottom"`
	DistT        string        `xml:"distT,attr,omitempty"`
	DistB        string        `xml:"distB,attr,omitempty"`
	EffectExtent *EffectExtent `xml:"wp:effectExtent,omitempty"`
}

// WrapPolygon 环绕多边形
type WrapPolygon struct {
	XMLName xml.Name        `xml:"wp:wrapPolygon"`
	Start   *PolygonStart   `xml:"wp:start"`
	LineTo  []PolygonLineTo `xml:"wp:lineTo"`
}

// PolygonStart 多边形起点
type PolygonStart struct {
	XMLName xml.Name `xml:"wp:start"`
	X       string   `xml:"x,attr"`
	Y       string   `xml:"y,attr"`
}

// PolygonLineTo 多边形线段
type PolygonLineTo struct {
	XMLName xml.Name `xml:"wp:lineTo"`
	X       string   `xml:"x,attr"`
	Y       string   `xml:"y,attr"`
}

// CNvGraphicFramePr 非可视图形框架属性
type CNvGraphicFramePr struct {
	XMLName           xml.Name           `xml:"wp:cNvGraphicFramePr"`
	GraphicFrameLocks *GraphicFrameLocks `xml:"a:graphicFrameLocks,omitempty"`
}

// GraphicFrameLocks 图形框架锁定
type GraphicFrameLocks struct {
	XMLName        xml.Name `xml:"a:graphicFrameLocks"`
	Xmlns          string   `xml:"xmlns:a,attr,omitempty"`
	NoChangeAspect string   `xml:"noChangeAspect,attr,omitempty"`
	NoCrop         string   `xml:"noCrop,attr,omitempty"`
	NoMove         string   `xml:"noMove,attr,omitempty"`
	NoResize       string   `xml:"noResize,attr,omitempty"`
	NoRot          string   `xml:"noRot,attr,omitempty"`
	NoSelect       string   `xml:"noSelect,attr,omitempty"`
}

// DrawingExtent 尺寸
type DrawingExtent struct {
	XMLName xml.Name `xml:"wp:extent"`
	Cx      string   `xml:"cx,attr"`
	Cy      string   `xml:"cy,attr"`
}

// DrawingDocPr 文档属性
type DrawingDocPr struct {
	XMLName xml.Name `xml:"wp:docPr"`
	ID      string   `xml:"id,attr"`
	Name    string   `xml:"name,attr"`
	Descr   string   `xml:"descr,attr,omitempty"`
	Title   string   `xml:"title,attr,omitempty"`
}

// DrawingGraphic 图形
type DrawingGraphic struct {
	XMLName     xml.Name     `xml:"a:graphic"`
	Xmlns       string       `xml:"xmlns:a,attr"`
	GraphicData *GraphicData `xml:"a:graphicData"`
}

// GraphicData 图形数据
type GraphicData struct {
	XMLName xml.Name    `xml:"a:graphicData"`
	Uri     string      `xml:"uri,attr"`
	Pic     *PicElement `xml:"pic:pic"`
}

// PicElement 图片
type PicElement struct {
	XMLName  xml.Name  `xml:"pic:pic"`
	Xmlns    string    `xml:"xmlns:pic,attr"`
	NvPicPr  *NvPicPr  `xml:"pic:nvPicPr"`
	BlipFill *BlipFill `xml:"pic:blipFill"`
	SpPr     *SpPr     `xml:"pic:spPr"`
}

// NvPicPr 非可视图片属性
type NvPicPr struct {
	XMLName  xml.Name  `xml:"pic:nvPicPr"`
	CNvPr    *CNvPr    `xml:"pic:cNvPr"`
	CNvPicPr *CNvPicPr `xml:"pic:cNvPicPr"`
}

// CNvPr 通用非可视属性
type CNvPr struct {
	XMLName xml.Name `xml:"pic:cNvPr"`
	ID      string   `xml:"id,attr"`
	Name    string   `xml:"name,attr"`
	Descr   string   `xml:"descr,attr,omitempty"`
	Title   string   `xml:"title,attr,omitempty"`
}

// CNvPicPr 图片特定非可视属性
type CNvPicPr struct {
	XMLName  xml.Name  `xml:"pic:cNvPicPr"`
	PicLocks *PicLocks `xml:"a:picLocks,omitempty"`
}

// PicLocks 图片锁定属性
type PicLocks struct {
	XMLName            xml.Name `xml:"a:picLocks"`
	NoChangeAspect     string   `xml:"noChangeAspect,attr,omitempty"`
	NoChangeArrowheads string   `xml:"noChangeArrowheads,attr,omitempty"`
}

// BlipFill 图片填充
type BlipFill struct {
	XMLName xml.Name `xml:"pic:blipFill"`
	Blip    *Blip    `xml:"a:blip"`
	Stretch *Stretch `xml:"a:stretch"`
}

// Blip 二进制图片
type Blip struct {
	XMLName xml.Name `xml:"a:blip"`
	Embed   string   `xml:"r:embed,attr"`
}

// Stretch 拉伸
type Stretch struct {
	XMLName  xml.Name  `xml:"a:stretch"`
	FillRect *FillRect `xml:"a:fillRect"`
}

// FillRect 填充矩形
type FillRect struct {
	XMLName xml.Name `xml:"a:fillRect"`
}

// SpPr 形状属性
type SpPr struct {
	XMLName  xml.Name  `xml:"pic:spPr"`
	Xfrm     *Xfrm     `xml:"a:xfrm"`
	PrstGeom *PrstGeom `xml:"a:prstGeom"`
}

// Xfrm 变换
type Xfrm struct {
	XMLName xml.Name `xml:"a:xfrm"`
	Off     *Off     `xml:"a:off,omitempty"`
	Ext     *Ext     `xml:"a:ext"`
}

// Off 偏移
type Off struct {
	XMLName xml.Name `xml:"a:off"`
	X       string   `xml:"x,attr"`
	Y       string   `xml:"y,attr"`
}

// Ext 范围
type Ext struct {
	XMLName xml.Name `xml:"a:ext"`
	Cx      string   `xml:"cx,attr"`
	Cy      string   `xml:"cy,attr"`
}

// PrstGeom 预设几何图形
type PrstGeom struct {
	XMLName xml.Name `xml:"a:prstGeom"`
	Prst    string   `xml:"prst,attr"`
	AvLst   *AvLst   `xml:"a:avLst"`
}

// AvLst 调整值列表
type AvLst struct {
	XMLName xml.Name `xml:"a:avLst"`
}

// AddImageFromFile 从文件添加图片到文档
func (d *Document) AddImageFromFile(filePath string, config *ImageConfig) (*ImageInfo, error) {
	Debugf("开始添加图片文件: %s", filePath)

	// 读取图片文件
	imageData, err := os.ReadFile(filePath)
	if err != nil {
		Errorf("读取图片文件失败 %s: %v", filePath, err)
		return nil, fmt.Errorf("读取图片文件失败: %v", err)
	}

	// 检测图片格式
	format, err := detectImageFormat(imageData)
	if err != nil {
		Errorf("检测图片格式失败 %s: %v", filePath, err)
		return nil, fmt.Errorf("检测图片格式失败: %v", err)
	}

	// 获取图片尺寸
	width, height, err := getImageDimensions(imageData, format)
	if err != nil {
		Errorf("获取图片尺寸失败 %s: %v", filePath, err)
		return nil, fmt.Errorf("获取图片尺寸失败: %v", err)
	}

	fileName := filepath.Base(filePath)
	Infof("成功读取图片: %s (格式: %s, 尺寸: %dx%d, 大小: %d字节)", fileName, format, width, height, len(imageData))
	return d.AddImageFromData(imageData, fileName, format, width, height, config)
}

// generateSafeImageFileName 生成安全的图片文件名
// 将非ASCII字符的文件名转换为安全的ASCII文件名，以确保Microsoft Word兼容性
func generateSafeImageFileName(imageID int, originalFileName string, format ImageFormat) string {
	// 获取文件扩展名
	ext := filepath.Ext(originalFileName)
	if ext == "" {
		// 如果没有扩展名，根据格式添加
		switch format {
		case ImageFormatPNG:
			ext = ".png"
		case ImageFormatJPEG:
			ext = ".jpeg"
		case ImageFormatGIF:
			ext = ".gif"
		default:
			ext = ".png"
		}
	}

	// 使用图片ID生成安全的文件名
	safeFileName := fmt.Sprintf("image%d%s", imageID, ext)
	return safeFileName
}

// AddImageFromData 从数据添加图片到文档
func (d *Document) AddImageFromData(imageData []byte, fileName string, format ImageFormat, width, height int, config *ImageConfig) (*ImageInfo, error) {
	if d.documentRelationships == nil {
		d.documentRelationships = &Relationships{
			Xmlns:         "http://schemas.openxmlformats.org/package/2006/relationships",
			Relationships: []Relationship{},
		}
	}

	// 使用文档级别的图片ID计数器确保ID唯一性
	imageID := d.nextImageID
	d.nextImageID++ // 递增计数器

	// 生成安全的文件名（避免中文等非ASCII字符导致Word打开错误）
	safeFileName := generateSafeImageFileName(imageID, fileName, format)

	// 生成关系ID，注意：rId1保留给styles.xml，图片从rId2开始
	relationID := fmt.Sprintf("rId%d", len(d.documentRelationships.Relationships)+2)

	// 添加图片关系，使用安全文件名
	d.documentRelationships.Relationships = append(d.documentRelationships.Relationships, Relationship{
		ID:     relationID,
		Type:   "http://schemas.openxmlformats.org/officeDocument/2006/relationships/image",
		Target: fmt.Sprintf("media/%s", safeFileName),
	})

	// 存储图片数据，使用安全文件名
	if d.parts == nil {
		d.parts = make(map[string][]byte)
	}
	d.parts[fmt.Sprintf("word/media/%s", safeFileName)] = imageData

	// 更新内容类型
	d.addImageContentType(format)

	// 创建图片信息
	imageInfo := &ImageInfo{
		ID:         strconv.Itoa(imageID),
		RelationID: relationID,
		Format:     format,
		Width:      width,
		Height:     height,
		Data:       imageData,
		Config:     config,
	}

	// 创建图片段落并添加到文档
	paragraph := d.createImageParagraph(imageInfo)
	d.Body.AddElement(paragraph)

	return imageInfo, nil
}

// AddImageFromDataWithoutElement 从数据添加图片到文档但不创建段落元素
// 此方法供模板引擎等需要自行管理图片段落的场景使用
func (d *Document) AddImageFromDataWithoutElement(imageData []byte, fileName string, format ImageFormat, width, height int, config *ImageConfig) (*ImageInfo, error) {
	if d.documentRelationships == nil {
		d.documentRelationships = &Relationships{
			Xmlns:         "http://schemas.openxmlformats.org/package/2006/relationships",
			Relationships: []Relationship{},
		}
	}

	// 使用文档级别的图片ID计数器确保ID唯一性
	imageID := d.nextImageID
	d.nextImageID++ // 递增计数器

	// 生成安全的文件名（避免中文等非ASCII字符导致Word打开错误）
	safeFileName := generateSafeImageFileName(imageID, fileName, format)

	// 生成关系ID，注意：rId1保留给styles.xml，图片从rId2开始
	relationID := fmt.Sprintf("rId%d", len(d.documentRelationships.Relationships)+2)

	// 添加图片关系，使用安全文件名
	d.documentRelationships.Relationships = append(d.documentRelationships.Relationships, Relationship{
		ID:     relationID,
		Type:   "http://schemas.openxmlformats.org/officeDocument/2006/relationships/image",
		Target: fmt.Sprintf("media/%s", safeFileName),
	})

	// 存储图片数据，使用安全文件名
	if d.parts == nil {
		d.parts = make(map[string][]byte)
	}
	d.parts[fmt.Sprintf("word/media/%s", safeFileName)] = imageData

	// 更新内容类型
	d.addImageContentType(format)

	// 创建图片信息
	imageInfo := &ImageInfo{
		ID:         strconv.Itoa(imageID),
		RelationID: relationID,
		Format:     format,
		Width:      width,
		Height:     height,
		Data:       imageData,
		Config:     config,
	}

	// 注意：这个方法不创建段落元素，由调用者负责管理
	return imageInfo, nil
}

// createImageParagraph 创建包含图片的段落
func (d *Document) createImageParagraph(imageInfo *ImageInfo) *Paragraph {
	// 计算图片显示尺寸（EMU单位）
	displayWidth, displayHeight := d.calculateDisplaySize(imageInfo)

	// 获取图片描述和标题
	altText := "图片"
	title := "图片"
	if imageInfo.Config != nil {
		if imageInfo.Config.AltText != "" {
			altText = imageInfo.Config.AltText
		}
		if imageInfo.Config.Title != "" {
			title = imageInfo.Config.Title
		}
	}

	// 创建Drawing元素
	var drawing *DrawingElement

	// 检查是否是浮动图片
	if imageInfo.Config != nil &&
		(imageInfo.Config.Position == ImagePositionFloatLeft ||
			imageInfo.Config.Position == ImagePositionFloatRight) {
		// 创建浮动图片
		drawing = d.createFloatingImageDrawing(imageInfo, displayWidth, displayHeight, altText, title)
	} else {
		// 创建嵌入式图片
		drawing = d.createInlineImageDrawing(imageInfo, displayWidth, displayHeight, altText, title)
	}

	// 创建包含图片的段落
	paragraph := &Paragraph{
		Runs: []Run{
			{
				Drawing: drawing,
			},
		},
	}

	// 为嵌入式图片设置段落对齐方式
	if imageInfo.Config != nil &&
		imageInfo.Config.Position == ImagePositionInline &&
		imageInfo.Config.Alignment != "" {
		paragraph.Properties = &ParagraphProperties{
			Justification: &Justification{Val: string(imageInfo.Config.Alignment)},
		}
	}

	return paragraph
}

// createInlineImageDrawing 创建嵌入式图片绘图元素
func (d *Document) createInlineImageDrawing(imageInfo *ImageInfo, displayWidth, displayHeight int64, altText, title string) *DrawingElement {
	return &DrawingElement{
		Inline: &InlineDrawing{
			DistT: "0",
			DistB: "0",
			DistL: "0",
			DistR: "0",
			Extent: &DrawingExtent{
				Cx: fmt.Sprintf("%d", displayWidth),
				Cy: fmt.Sprintf("%d", displayHeight),
			},
			DocPr: &DrawingDocPr{
				ID:    imageInfo.ID,
				Name:  fmt.Sprintf("图片 %s", imageInfo.ID),
				Descr: altText,
				Title: title,
			},
			Graphic: d.createImageGraphic(imageInfo, displayWidth, displayHeight, altText, title),
		},
	}
}

// createFloatingImageDrawing 创建浮动图片绘图元素
func (d *Document) createFloatingImageDrawing(imageInfo *ImageInfo, displayWidth, displayHeight int64, altText, title string) *DrawingElement {
	config := imageInfo.Config

	// 计算距离（EMU单位）
	distT := "0"
	distB := "0"
	distL := "0"
	distR := "0"

	if config.OffsetX > 0 {
		distL = fmt.Sprintf("%.0f", config.OffsetX*36000) // 毫米转EMU
		distR = fmt.Sprintf("%.0f", config.OffsetX*36000)
	}
	if config.OffsetY > 0 {
		distT = fmt.Sprintf("%.0f", config.OffsetY*36000)
		distB = fmt.Sprintf("%.0f", config.OffsetY*36000)
	}

	anchor := &AnchorDrawing{
		DistT:          distT,
		DistB:          distB,
		DistL:          distL,
		DistR:          distR,
		SimplePos:      "0",
		RelativeHeight: "251658240",
		BehindDoc:      "0",
		Locked:         "0",
		LayoutInCell:   "1",
		AllowOverlap:   "1",
		SimplePosition: &SimplePosition{
			X: "0",
			Y: "0",
		},
		Extent: &DrawingExtent{
			Cx: fmt.Sprintf("%d", displayWidth),
			Cy: fmt.Sprintf("%d", displayHeight),
		},
		EffectExtent: &EffectExtent{
			L: "0",
			T: "0",
			R: "0",
			B: "0",
		},
		DocPr: &DrawingDocPr{
			ID:    imageInfo.ID,
			Name:  fmt.Sprintf("图片 %s", imageInfo.ID),
			Descr: altText,
			Title: title,
		},
		CNvGraphicFramePr: &CNvGraphicFramePr{
			GraphicFrameLocks: &GraphicFrameLocks{
				Xmlns:          "http://schemas.openxmlformats.org/drawingml/2006/main",
				NoChangeAspect: "1",
			},
		},
		Graphic: d.createImageGraphic(imageInfo, displayWidth, displayHeight, altText, title),
	}

	// 设置位置
	d.setFloatingImagePosition(anchor, config)

	// 设置文字环绕
	d.setFloatingImageWrap(anchor, config)

	return &DrawingElement{
		Anchor: anchor,
	}
}

// setFloatingImagePosition 设置浮动图片位置
func (d *Document) setFloatingImagePosition(anchor *AnchorDrawing, config *ImageConfig) {
	if config.Position == ImagePositionFloatLeft {
		// 左浮动
		anchor.PositionH = &HorizontalPosition{
			RelativeFrom: "margin",
			Align: &PosAlign{
				Value: "left",
			},
		}
	} else if config.Position == ImagePositionFloatRight {
		// 右浮动
		anchor.PositionH = &HorizontalPosition{
			RelativeFrom: "margin",
			Align: &PosAlign{
				Value: "right",
			},
		}
	} else {
		// 默认居中
		anchor.PositionH = &HorizontalPosition{
			RelativeFrom: "margin",
			Align: &PosAlign{
				Value: "center",
			},
		}
	}

	// 垂直位置设置为顶部对齐
	anchor.PositionV = &VerticalPosition{
		RelativeFrom: "margin",
		Align: &PosAlign{
			Value: "top",
		},
	}

	// 如果有偏移量，使用偏移而不是对齐
	if config.OffsetX != 0 || config.OffsetY != 0 {
		if config.OffsetX != 0 {
			anchor.PositionH.Align = nil
			anchor.PositionH.PosOffset = &PosOffset{
				Value: fmt.Sprintf("%.0f", config.OffsetX*36000),
			}
		}
		if config.OffsetY != 0 {
			anchor.PositionV.Align = nil
			anchor.PositionV.PosOffset = &PosOffset{
				Value: fmt.Sprintf("%.0f", config.OffsetY*36000),
			}
		}
	}
}

// setFloatingImageWrap 设置浮动图片文字环绕
func (d *Document) setFloatingImageWrap(anchor *AnchorDrawing, config *ImageConfig) {
	// 计算环绕距离
	wrapDistL := "114300" // 默认3毫米
	wrapDistR := "114300"
	wrapDistT := "0"
	wrapDistB := "0"

	if config.OffsetX > 0 {
		wrapDistL = fmt.Sprintf("%.0f", config.OffsetX*36000)
		wrapDistR = fmt.Sprintf("%.0f", config.OffsetX*36000)
	}
	if config.OffsetY > 0 {
		wrapDistT = fmt.Sprintf("%.0f", config.OffsetY*36000)
		wrapDistB = fmt.Sprintf("%.0f", config.OffsetY*36000)
	}

	switch config.WrapText {
	case ImageWrapNone:
		anchor.WrapNone = &WrapNone{}
	case ImageWrapSquare:
		wrapText := "bothSides"
		if config.Position == ImagePositionFloatLeft {
			wrapText = "right"
		} else if config.Position == ImagePositionFloatRight {
			wrapText = "left"
		}
		anchor.WrapSquare = &WrapSquare{
			WrapText: wrapText,
			DistT:    wrapDistT,
			DistB:    wrapDistB,
			DistL:    wrapDistL,
			DistR:    wrapDistR,
		}
	case ImageWrapTight:
		wrapText := "bothSides"
		if config.Position == ImagePositionFloatLeft {
			wrapText = "right"
		} else if config.Position == ImagePositionFloatRight {
			wrapText = "left"
		}
		anchor.WrapTight = &WrapTight{
			WrapText:    wrapText,
			DistL:       wrapDistL,
			DistR:       wrapDistR,
			WrapPolygon: d.createDefaultWrapPolygon(), // 添加必需的WrapPolygon
		}
	case ImageWrapTopAndBottom:
		anchor.WrapTopAndBottom = &WrapTopAndBottom{
			DistT: wrapDistT,
			DistB: wrapDistB,
		}
	default:
		// 默认使用四周环绕
		wrapText := "bothSides"
		if config.Position == ImagePositionFloatLeft {
			wrapText = "right"
		} else if config.Position == ImagePositionFloatRight {
			wrapText = "left"
		}
		anchor.WrapSquare = &WrapSquare{
			WrapText: wrapText,
			DistT:    wrapDistT,
			DistB:    wrapDistB,
			DistL:    wrapDistL,
			DistR:    wrapDistR,
		}
	}
}

// createDefaultWrapPolygon 创建默认的环绕多边形
// 这个方法创建一个矩形的环绕路径，符合OpenXML规范
func (d *Document) createDefaultWrapPolygon() *WrapPolygon {
	return &WrapPolygon{
		Start: &PolygonStart{
			X: "0",
			Y: "0",
		},
		LineTo: []PolygonLineTo{
			{X: "0", Y: "21600"},
			{X: "21600", Y: "21600"},
			{X: "21600", Y: "0"},
			{X: "0", Y: "0"},
		},
	}
}

// createImageGraphic 创建图片图形元素
func (d *Document) createImageGraphic(imageInfo *ImageInfo, displayWidth, displayHeight int64, altText, title string) *DrawingGraphic {
	return &DrawingGraphic{
		Xmlns: "http://schemas.openxmlformats.org/drawingml/2006/main",
		GraphicData: &GraphicData{
			Uri: "http://schemas.openxmlformats.org/drawingml/2006/picture",
			Pic: &PicElement{
				Xmlns: "http://schemas.openxmlformats.org/drawingml/2006/picture",
				NvPicPr: &NvPicPr{
					CNvPr: &CNvPr{
						ID:    imageInfo.ID,
						Name:  fmt.Sprintf("图片 %s", imageInfo.ID),
						Descr: altText,
						Title: title,
					},
					CNvPicPr: &CNvPicPr{
						PicLocks: &PicLocks{
							NoChangeAspect: "1",
						},
					},
				},
				BlipFill: &BlipFill{
					Blip: &Blip{
						Embed: imageInfo.RelationID,
					},
					Stretch: &Stretch{
						FillRect: &FillRect{},
					},
				},
				SpPr: &SpPr{
					Xfrm: &Xfrm{
						Off: &Off{
							X: "0",
							Y: "0",
						},
						Ext: &Ext{
							Cx: fmt.Sprintf("%d", displayWidth),
							Cy: fmt.Sprintf("%d", displayHeight),
						},
					},
					PrstGeom: &PrstGeom{
						Prst:  "rect",
						AvLst: &AvLst{},
					},
				},
			},
		},
	}
}

// calculateDisplaySize 计算图片显示尺寸（EMU单位）
func (d *Document) calculateDisplaySize(imageInfo *ImageInfo) (int64, int64) {
	config := imageInfo.Config
	originalWidth := int64(imageInfo.Width)
	originalHeight := int64(imageInfo.Height)

	// 默认使用原始尺寸（96 DPI）
	// 1像素 = 9525 EMU (at 96 DPI)
	displayWidth := originalWidth * 9525
	displayHeight := originalHeight * 9525

	if config != nil && config.Size != nil {
		if config.Size.Width > 0 && config.Size.Height > 0 {
			// 用户指定了具体尺寸
			displayWidth = int64(config.Size.Width * 36000)   // 毫米转EMU
			displayHeight = int64(config.Size.Height * 36000) // 毫米转EMU
		} else if config.Size.Width > 0 && config.Size.KeepAspectRatio {
			// 只指定宽度，保持长宽比
			displayWidth = int64(config.Size.Width * 36000)
			ratio := float64(originalHeight) / float64(originalWidth)
			displayHeight = int64(float64(displayWidth) * ratio)
		} else if config.Size.Height > 0 && config.Size.KeepAspectRatio {
			// 只指定高度，保持长宽比
			displayHeight = int64(config.Size.Height * 36000)
			ratio := float64(originalWidth) / float64(originalHeight)
			displayWidth = int64(float64(displayHeight) * ratio)
		}
	}

	return displayWidth, displayHeight
}

// detectImageFormat 检测图片格式
func detectImageFormat(data []byte) (ImageFormat, error) {
	if len(data) < 3 {
		return "", fmt.Errorf("图片数据太短")
	}

	// 检测PNG
	if len(data) >= 8 && bytes.Equal(data[:8], []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}) {
		return ImageFormatPNG, nil
	}

	// 检测JPEG
	if len(data) >= 3 && bytes.Equal(data[:3], []byte{0xFF, 0xD8, 0xFF}) {
		return ImageFormatJPEG, nil
	}

	// 检测GIF
	if len(data) >= 6 && (bytes.Equal(data[:6], []byte("GIF87a")) || bytes.Equal(data[:6], []byte("GIF89a"))) {
		return ImageFormatGIF, nil
	}

	return "", fmt.Errorf("不支持的图片格式")
}

// getImageDimensions 获取图片尺寸
func getImageDimensions(data []byte, format ImageFormat) (int, int, error) {
	reader := bytes.NewReader(data)

	var img image.Image
	var err error

	switch format {
	case ImageFormatPNG:
		img, err = png.Decode(reader)
	case ImageFormatJPEG:
		img, err = jpeg.Decode(reader)
	case ImageFormatGIF:
		img, err = gif.Decode(reader)
	default:
		return 0, 0, fmt.Errorf("不支持的图片格式: %s", format)
	}

	if err != nil {
		return 0, 0, fmt.Errorf("解码图片失败: %v", err)
	}

	bounds := img.Bounds()
	return bounds.Dx(), bounds.Dy(), nil
}

// addImageContentType 添加图片内容类型
func (d *Document) addImageContentType(format ImageFormat) {
	if d.contentTypes == nil {
		d.contentTypes = &ContentTypes{
			Xmlns:     "http://schemas.openxmlformats.org/package/2006/content-types",
			Defaults:  []Default{},
			Overrides: []Override{},
		}
	}

	var extension, contentType string
	switch format {
	case ImageFormatPNG:
		extension = "png"
		contentType = "image/png"
	case ImageFormatJPEG:
		extension = "jpeg"
		contentType = "image/jpeg"
	case ImageFormatGIF:
		extension = "gif"
		contentType = "image/gif"
	default:
		return
	}

	// 检查是否已存在相同的默认类型
	for _, def := range d.contentTypes.Defaults {
		if def.Extension == extension {
			return
		}
	}

	// 添加默认内容类型
	d.contentTypes.Defaults = append(d.contentTypes.Defaults, Default{
		Extension:   extension,
		ContentType: contentType,
	})
}

// ResizeImage 调整图片大小
func (d *Document) ResizeImage(imageInfo *ImageInfo, size *ImageSize) error {
	if imageInfo == nil {
		return fmt.Errorf("图片信息不能为空")
	}

	if imageInfo.Config == nil {
		imageInfo.Config = &ImageConfig{}
	}

	imageInfo.Config.Size = size

	// 如果图片已经被添加到文档中，需要重新生成
	// 注意：这是一个简化的实现，实际应用中可能需要更复杂的更新机制
	return nil
}

// SetImagePosition 设置图片位置
func (d *Document) SetImagePosition(imageInfo *ImageInfo, position ImagePosition, offsetX, offsetY float64) error {
	if imageInfo == nil {
		return fmt.Errorf("图片信息不能为空")
	}

	if imageInfo.Config == nil {
		imageInfo.Config = &ImageConfig{}
	}

	imageInfo.Config.Position = position
	imageInfo.Config.OffsetX = offsetX
	imageInfo.Config.OffsetY = offsetY

	// 如果位置发生变化（从inline到float或反之），可能需要重新生成Drawing元素
	// 注意：这是一个简化的实现，实际应用中可能需要更复杂的更新机制
	return nil
}

// SetImageWrapText 设置图片文字环绕
func (d *Document) SetImageWrapText(imageInfo *ImageInfo, wrapText ImageWrapText) error {
	if imageInfo == nil {
		return fmt.Errorf("图片信息不能为空")
	}

	if imageInfo.Config == nil {
		imageInfo.Config = &ImageConfig{}
	}

	imageInfo.Config.WrapText = wrapText
	return nil
}

// SetImageAltText 设置图片替代文字
func (d *Document) SetImageAltText(imageInfo *ImageInfo, altText string) error {
	if imageInfo == nil {
		return fmt.Errorf("图片信息不能为空")
	}

	if imageInfo.Config == nil {
		imageInfo.Config = &ImageConfig{}
	}

	imageInfo.Config.AltText = altText
	return nil
}

// SetImageTitle 设置图片标题
func (d *Document) SetImageTitle(imageInfo *ImageInfo, title string) error {
	if imageInfo == nil {
		return fmt.Errorf("图片信息不能为空")
	}

	if imageInfo.Config == nil {
		imageInfo.Config = &ImageConfig{}
	}

	imageInfo.Config.Title = title
	return nil
}

// SetImageAlignment 设置图片对齐方式
//
// 此方法用于设置嵌入式图片（ImagePositionInline）的对齐方式。
// 对于浮动图片，请使用SetImagePosition方法。
//
// 参数 alignment 指定对齐类型，支持以下值：
//   - AlignLeft: 左对齐
//   - AlignCenter: 居中对齐
//   - AlignRight: 右对齐
//   - AlignJustify: 两端对齐
//
// 示例:
//
//	imageInfo, err := doc.AddImageFromFile("image.png", nil)
//	if err != nil {
//		return err
//	}
//	err = doc.SetImageAlignment(imageInfo, document.AlignCenter)
func (d *Document) SetImageAlignment(imageInfo *ImageInfo, alignment AlignmentType) error {
	if imageInfo == nil {
		return fmt.Errorf("图片信息不能为空")
	}

	if imageInfo.Config == nil {
		imageInfo.Config = &ImageConfig{}
	}

	// 更新配置
	imageInfo.Config.Alignment = alignment

	// 查找包含此图片的段落并更新其对齐方式
	for _, element := range d.Body.Elements {
		if paragraph, ok := element.(*Paragraph); ok {
			// 检查段落中是否包含指定的图片
			for _, run := range paragraph.Runs {
				if run.Drawing != nil && run.Drawing.Inline != nil {
					// 检查docPr ID是否匹配
					if run.Drawing.Inline.DocPr != nil && run.Drawing.Inline.DocPr.ID == imageInfo.ID {
						// 更新段落对齐方式
						if paragraph.Properties == nil {
							paragraph.Properties = &ParagraphProperties{}
						}
						paragraph.Properties.Justification = &Justification{Val: string(alignment)}
						return nil
					}
				}
			}
		}
	}

	// 未找到包含图片的段落：保持配置已更新，返回nil以兼容在图片尚未插入段落前先设置对齐的用例。
	Debugf("未找到包含图片ID %s 的段落，已仅更新配置中的对齐方式", imageInfo.ID)
	return nil
}
