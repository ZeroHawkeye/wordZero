// Package document 提供Word文档的核心操作功能
package document

import (
	"encoding/xml"
)

// OfficeMath 表示Office数学公式元素
// 对应OMML中的 m:oMath 元素
type OfficeMath struct {
	XMLName xml.Name `xml:"m:oMath"`
	Xmlns   string   `xml:"xmlns:m,attr,omitempty"`
	RawXML  string   `xml:",innerxml"` // 存储内部XML内容
}

// OfficeMathPara 表示Office数学公式段落（用于块级公式）
// 对应OMML中的 m:oMathPara 元素
type OfficeMathPara struct {
	XMLName xml.Name    `xml:"m:oMathPara"`
	Xmlns   string      `xml:"xmlns:m,attr,omitempty"`
	Math    *OfficeMath `xml:"m:oMath"`
}

// MathParagraph 表示包含数学公式的段落
// 用于在文档中嵌入数学公式
type MathParagraph struct {
	XMLName    xml.Name             `xml:"w:p"`
	Properties *ParagraphProperties `xml:"w:pPr,omitempty"`
	Math       *OfficeMath          `xml:"m:oMath,omitempty"`
	MathPara   *OfficeMathPara      `xml:"m:oMathPara,omitempty"`
	Runs       []Run                `xml:"w:r"`
}

// MarshalXML 自定义序列化
func (mp *MathParagraph) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	// 开始段落元素
	if err := e.EncodeToken(start); err != nil {
		return err
	}

	// 序列化段落属性
	if mp.Properties != nil {
		if err := e.Encode(mp.Properties); err != nil {
			return err
		}
	}

	// 序列化Runs（在公式之前的文本）
	for _, run := range mp.Runs {
		if err := e.Encode(run); err != nil {
			return err
		}
	}

	// 序列化数学公式（块级）
	if mp.MathPara != nil {
		if err := e.Encode(mp.MathPara); err != nil {
			return err
		}
	}

	// 序列化数学公式（行内）
	if mp.Math != nil {
		if err := e.Encode(mp.Math); err != nil {
			return err
		}
	}

	// 结束段落元素
	return e.EncodeToken(start.End())
}

// ElementType 返回数学段落元素类型
func (mp *MathParagraph) ElementType() string {
	return "math_paragraph"
}

// AddMathFormula 向文档添加数学公式
// latex: LaTeX格式的数学公式
// isBlock: 是否为块级公式（true为块级，false为行内）
func (d *Document) AddMathFormula(latex string, isBlock bool) *MathParagraph {
	Debugf("添加数学公式: %s (块级: %v)", latex, isBlock)

	mp := &MathParagraph{
		Runs: []Run{},
	}

	// 创建公式内容
	// 注意：这里使用RawXML来存储公式内容，因为OMML结构复杂
	// 实际的LaTeX到OMML转换由markdown包的LaTeXToOMML函数完成
	if isBlock {
		mp.MathPara = &OfficeMathPara{
			Xmlns: "http://schemas.openxmlformats.org/officeDocument/2006/math",
			Math: &OfficeMath{
				Xmlns:  "http://schemas.openxmlformats.org/officeDocument/2006/math",
				RawXML: latex, // 这里存储的是预处理过的OMML内容
			},
		}
	} else {
		mp.Math = &OfficeMath{
			Xmlns:  "http://schemas.openxmlformats.org/officeDocument/2006/math",
			RawXML: latex,
		}
	}

	d.Body.Elements = append(d.Body.Elements, mp)
	return mp
}

// AddInlineMathFormula 向段落中添加行内数学公式
// 这将在当前段落的末尾添加一个数学公式
func (p *Paragraph) AddInlineMath(ommlContent string) {
	Debugf("向段落添加行内数学公式")

	// Create a special Run to contain formula reference
	// Note: In Word, inline formulas are implemented through special oMath elements
	// Here we use a placeholder method, actual implementation needs to modify paragraph serialization logic
	run := Run{
		Text: Text{
			Content: "[Formula]", // Placeholder, actual formula content is processed during serialization
			Space:   "preserve",
		},
	}
	p.Runs = append(p.Runs, run)
}
