// Package markdown 提供Markdown到Word文档的转换功能
package markdown

import (
	"encoding/xml"
	"regexp"
	"strings"
)

// OfficeMath 表示Office数学公式的根元素
// 对应OMML中的 m:oMath 元素
type OfficeMath struct {
	XMLName xml.Name      `xml:"m:oMath"`
	Content []interface{} `xml:"-"` // 使用自定义序列化
}

// MarshalXML 自定义XML序列化
func (o *OfficeMath) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name = xml.Name{Local: "m:oMath"}
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	for _, content := range o.Content {
		if err := e.Encode(content); err != nil {
			return err
		}
	}
	return e.EncodeToken(start.End())
}

// OfficeMathPara 表示Office数学公式段落
// 对应OMML中的 m:oMathPara 元素（用于块级公式）
type OfficeMathPara struct {
	XMLName xml.Name    `xml:"m:oMathPara"`
	Math    *OfficeMath `xml:"m:oMath"`
}

// MathRun 表示数学运行元素
type MathRun struct {
	XMLName xml.Name     `xml:"m:r"`
	Text    *MathText    `xml:"m:t,omitempty"`
	RunPr   *MathRunProp `xml:"m:rPr,omitempty"`
}

// MathText 表示数学文本
type MathText struct {
	XMLName xml.Name `xml:"m:t"`
	Content string   `xml:",chardata"`
}

// MathRunProp 表示数学运行属性
type MathRunProp struct {
	XMLName xml.Name `xml:"m:rPr"`
	Sty     *MathSty `xml:"m:sty,omitempty"`
}

// MathSty 表示数学样式
type MathSty struct {
	XMLName xml.Name `xml:"m:sty"`
	Val     string   `xml:"m:val,attr"`
}

// MathFrac 表示分数
type MathFrac struct {
	XMLName xml.Name  `xml:"m:f"`
	FracPr  *MathFracPr `xml:"m:fPr,omitempty"`
	Num     *MathNum  `xml:"m:num"`
	Den     *MathDen  `xml:"m:den"`
}

// MathFracPr 表示分数属性
type MathFracPr struct {
	XMLName xml.Name `xml:"m:fPr"`
	Type    *MathFracType `xml:"m:type,omitempty"`
}

// MathFracType 表示分数类型
type MathFracType struct {
	XMLName xml.Name `xml:"m:type"`
	Val     string   `xml:"m:val,attr"`
}

// MathNum 表示分子
type MathNum struct {
	XMLName xml.Name      `xml:"m:num"`
	Content []interface{} `xml:"-"`
}

// MarshalXML 自定义XML序列化
func (n *MathNum) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name = xml.Name{Local: "m:num"}
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	for _, content := range n.Content {
		if err := e.Encode(content); err != nil {
			return err
		}
	}
	return e.EncodeToken(start.End())
}

// MathDen 表示分母
type MathDen struct {
	XMLName xml.Name      `xml:"m:den"`
	Content []interface{} `xml:"-"`
}

// MarshalXML 自定义XML序列化
func (d *MathDen) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name = xml.Name{Local: "m:den"}
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	for _, content := range d.Content {
		if err := e.Encode(content); err != nil {
			return err
		}
	}
	return e.EncodeToken(start.End())
}

// MathSup 表示上标
type MathSup struct {
	XMLName xml.Name `xml:"m:sSup"`
	E       *MathE   `xml:"m:e"`
	Sup     *MathSupElement `xml:"m:sup"`
}

// MathE 表示基础元素
type MathE struct {
	XMLName xml.Name      `xml:"m:e"`
	Content []interface{} `xml:"-"`
}

// MarshalXML 自定义XML序列化
func (m *MathE) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name = xml.Name{Local: "m:e"}
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	for _, content := range m.Content {
		if err := e.Encode(content); err != nil {
			return err
		}
	}
	return e.EncodeToken(start.End())
}

// MathSupElement 表示上标元素
type MathSupElement struct {
	XMLName xml.Name      `xml:"m:sup"`
	Content []interface{} `xml:"-"`
}

// MarshalXML 自定义XML序列化
func (s *MathSupElement) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name = xml.Name{Local: "m:sup"}
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	for _, content := range s.Content {
		if err := e.Encode(content); err != nil {
			return err
		}
	}
	return e.EncodeToken(start.End())
}

// MathSub 表示下标
type MathSub struct {
	XMLName xml.Name `xml:"m:sSub"`
	E       *MathE   `xml:"m:e"`
	Sub     *MathSubElement `xml:"m:sub"`
}

// MathSubElement 表示下标元素
type MathSubElement struct {
	XMLName xml.Name      `xml:"m:sub"`
	Content []interface{} `xml:"-"`
}

// MarshalXML 自定义XML序列化
func (s *MathSubElement) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name = xml.Name{Local: "m:sub"}
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	for _, content := range s.Content {
		if err := e.Encode(content); err != nil {
			return err
		}
	}
	return e.EncodeToken(start.End())
}

// MathRad 表示根号
type MathRad struct {
	XMLName xml.Name   `xml:"m:rad"`
	RadPr   *MathRadPr `xml:"m:radPr,omitempty"`
	Deg     *MathDeg   `xml:"m:deg,omitempty"`
	E       *MathE     `xml:"m:e"`
}

// MathRadPr 表示根号属性
type MathRadPr struct {
	XMLName xml.Name     `xml:"m:radPr"`
	DegHide *MathDegHide `xml:"m:degHide,omitempty"`
}

// MathDegHide 表示是否隐藏根指数
type MathDegHide struct {
	XMLName xml.Name `xml:"m:degHide"`
	Val     string   `xml:"m:val,attr"`
}

// MathDeg 表示根指数
type MathDeg struct {
	XMLName xml.Name      `xml:"m:deg"`
	Content []interface{} `xml:"-"`
}

// MarshalXML 自定义XML序列化
func (d *MathDeg) MarshalXML(e *xml.Encoder, start xml.StartElement) error {
	start.Name = xml.Name{Local: "m:deg"}
	if err := e.EncodeToken(start); err != nil {
		return err
	}
	for _, content := range d.Content {
		if err := e.Encode(content); err != nil {
			return err
		}
	}
	return e.EncodeToken(start.End())
}

// MathSubSup 表示上下标
type MathSubSup struct {
	XMLName xml.Name        `xml:"m:sSubSup"`
	E       *MathE          `xml:"m:e"`
	Sub     *MathSubElement `xml:"m:sub"`
	Sup     *MathSupElement `xml:"m:sup"`
}

// MathDelim 表示分隔符（括号等）
type MathDelim struct {
	XMLName xml.Name     `xml:"m:d"`
	DPr     *MathDelimPr `xml:"m:dPr,omitempty"`
	E       *MathE       `xml:"m:e"`
}

// MathDelimPr 表示分隔符属性
type MathDelimPr struct {
	XMLName xml.Name       `xml:"m:dPr"`
	BegChr  *MathDelimChar `xml:"m:begChr,omitempty"`
	EndChr  *MathDelimChar `xml:"m:endChr,omitempty"`
}

// MathDelimChar 表示分隔符字符
type MathDelimChar struct {
	XMLName xml.Name `xml:"m:begChr"`
	Val     string   `xml:"m:val,attr"`
}

// LaTeXToOMML 将LaTeX公式转换为OMML格式
// 这是一个简化版本的转换器，支持常用的LaTeX数学语法
func LaTeXToOMML(latex string) *OfficeMath {
	latex = strings.TrimSpace(latex)
	omath := &OfficeMath{
		Content: []interface{}{},
	}

	// 解析LaTeX并转换为OMML
	content := parseLatex(latex)
	omath.Content = content

	return omath
}

// parseLatex 解析LaTeX字符串并返回OMML元素列表
func parseLatex(latex string) []interface{} {
	var result []interface{}
	latex = strings.TrimSpace(latex)

	// 处理空字符串
	if latex == "" {
		return result
	}

	// 定义正则表达式模式
	fracPattern := regexp.MustCompile(`^\\frac\s*\{([^{}]*(?:\{[^{}]*\}[^{}]*)*)\}\s*\{([^{}]*(?:\{[^{}]*\}[^{}]*)*)\}`)
	sqrtPattern := regexp.MustCompile(`^\\sqrt(?:\[([^\]]*)\])?\s*\{([^{}]*(?:\{[^{}]*\}[^{}]*)*)\}`)
	supPattern := regexp.MustCompile(`^([a-zA-Z0-9])\^(?:\{([^{}]*)\}|([a-zA-Z0-9]))`)
	subPattern := regexp.MustCompile(`^([a-zA-Z0-9])_(?:\{([^{}]*)\}|([a-zA-Z0-9]))`)
	subSupPattern := regexp.MustCompile(`^([a-zA-Z0-9])_(?:\{([^{}]*)\}|([a-zA-Z0-9]))\^(?:\{([^{}]*)\}|([a-zA-Z0-9]))`)
	cmdPattern := regexp.MustCompile(`^\\([a-zA-Z]+)`)
	textPattern := regexp.MustCompile(`^[a-zA-Z0-9.,;:!?\s\+\-\*\/\=\(\)\[\]]+`)

	i := 0
	for i < len(latex) {
		remaining := latex[i:]

		// 检查上下标组合
		if match := subSupPattern.FindStringSubmatch(remaining); match != nil {
			base := match[1]
			sub := match[2]
			if sub == "" {
				sub = match[3]
			}
			sup := match[4]
			if sup == "" {
				sup = match[5]
			}

			result = append(result, &MathSubSup{
				E:   &MathE{Content: []interface{}{createMathRun(base)}},
				Sub: &MathSubElement{Content: parseLatex(sub)},
				Sup: &MathSupElement{Content: parseLatex(sup)},
			})
			i += len(match[0])
			continue
		}

		// 检查分数
		if match := fracPattern.FindStringSubmatch(remaining); match != nil {
			num := match[1]
			den := match[2]
			result = append(result, &MathFrac{
				Num: &MathNum{Content: parseLatex(num)},
				Den: &MathDen{Content: parseLatex(den)},
			})
			i += len(match[0])
			continue
		}

		// 检查根号
		if match := sqrtPattern.FindStringSubmatch(remaining); match != nil {
			deg := match[1]  // 可能为空（平方根）
			content := match[2]
			rad := &MathRad{
				E: &MathE{Content: parseLatex(content)},
			}
			if deg == "" {
				// 平方根，隐藏根指数
				rad.RadPr = &MathRadPr{
					DegHide: &MathDegHide{Val: "1"},
				}
			} else {
				// n次方根
				rad.Deg = &MathDeg{Content: parseLatex(deg)}
			}
			result = append(result, rad)
			i += len(match[0])
			continue
		}

		// 检查上标
		if match := supPattern.FindStringSubmatch(remaining); match != nil {
			base := match[1]
			sup := match[2]
			if sup == "" {
				sup = match[3]
			}
			result = append(result, &MathSup{
				E:   &MathE{Content: []interface{}{createMathRun(base)}},
				Sup: &MathSupElement{Content: parseLatex(sup)},
			})
			i += len(match[0])
			continue
		}

		// 检查下标
		if match := subPattern.FindStringSubmatch(remaining); match != nil {
			base := match[1]
			sub := match[2]
			if sub == "" {
				sub = match[3]
			}
			result = append(result, &MathSub{
				E:   &MathE{Content: []interface{}{createMathRun(base)}},
				Sub: &MathSubElement{Content: parseLatex(sub)},
			})
			i += len(match[0])
			continue
		}

		// 检查LaTeX命令
		if match := cmdPattern.FindStringSubmatch(remaining); match != nil {
			cmd := match[1]
			cmdText := convertLaTeXCommand(cmd)
			result = append(result, createMathRun(cmdText))
			i += len(match[0])
			continue
		}

		// 检查花括号分组
		if remaining[0] == '{' {
			depth := 1
			j := 1
			for j < len(remaining) && depth > 0 {
				if remaining[j] == '{' {
					depth++
				} else if remaining[j] == '}' {
					depth--
				}
				j++
			}
			if depth == 0 {
				inner := remaining[1 : j-1]
				innerContent := parseLatex(inner)
				result = append(result, innerContent...)
				i += j
				continue
			}
		}

		// 检查普通文本
		if match := textPattern.FindString(remaining); match != "" {
			result = append(result, createMathRun(match))
			i += len(match)
			continue
		}

		// 处理单个字符
		if i < len(latex) {
			result = append(result, createMathRun(string(latex[i])))
			i++
		}
	}

	return result
}

// createMathRun 创建数学运行元素
func createMathRun(text string) *MathRun {
	return &MathRun{
		Text: &MathText{Content: text},
	}
}

// convertLaTeXCommand 将LaTeX命令转换为对应的Unicode字符
func convertLaTeXCommand(cmd string) string {
	// 常见LaTeX命令到Unicode的映射
	commands := map[string]string{
		// 希腊字母（小写）
		"alpha":   "α",
		"beta":    "β",
		"gamma":   "γ",
		"delta":   "δ",
		"epsilon": "ε",
		"zeta":    "ζ",
		"eta":     "η",
		"theta":   "θ",
		"iota":    "ι",
		"kappa":   "κ",
		"lambda":  "λ",
		"mu":      "μ",
		"nu":      "ν",
		"xi":      "ξ",
		"pi":      "π",
		"rho":     "ρ",
		"sigma":   "σ",
		"tau":     "τ",
		"upsilon": "υ",
		"phi":     "φ",
		"chi":     "χ",
		"psi":     "ψ",
		"omega":   "ω",

		// 希腊字母（大写）
		"Alpha":   "Α",
		"Beta":    "Β",
		"Gamma":   "Γ",
		"Delta":   "Δ",
		"Epsilon": "Ε",
		"Zeta":    "Ζ",
		"Eta":     "Η",
		"Theta":   "Θ",
		"Iota":    "Ι",
		"Kappa":   "Κ",
		"Lambda":  "Λ",
		"Mu":      "Μ",
		"Nu":      "Ν",
		"Xi":      "Ξ",
		"Pi":      "Π",
		"Rho":     "Ρ",
		"Sigma":   "Σ",
		"Tau":     "Τ",
		"Upsilon": "Υ",
		"Phi":     "Φ",
		"Chi":     "Χ",
		"Psi":     "Ψ",
		"Omega":   "Ω",

		// 运算符
		"times":   "×",
		"div":     "÷",
		"pm":      "±",
		"mp":      "∓",
		"cdot":    "·",
		"ast":     "∗",
		"star":    "⋆",
		"circ":    "∘",
		"bullet":  "∙",
		"oplus":   "⊕",
		"ominus":  "⊖",
		"otimes":  "⊗",
		"oslash":  "⊘",
		"odot":    "⊙",

		// 关系符号
		"leq":     "≤",
		"geq":     "≥",
		"neq":     "≠",
		"approx":  "≈",
		"equiv":   "≡",
		"sim":     "∼",
		"simeq":   "≃",
		"cong":    "≅",
		"propto":   "∝",
		"ll":       "≪",
		"gg":       "≫",
		"subset":   "⊂",
		"supset":   "⊃",
		"subseteq": "⊆",
		"supseteq": "⊇",
		"in":       "∈",
		"notin":    "∉",
		"ni":       "∋",

		// 箭头
		"rightarrow":     "→",
		"leftarrow":      "←",
		"leftrightarrow": "↔",
		"Rightarrow":     "⇒",
		"Leftarrow":      "⇐",
		"Leftrightarrow": "⇔",
		"uparrow":        "↑",
		"downarrow":      "↓",
		"to":             "→",
		"gets":           "←",
		"mapsto":         "↦",

		// 杂项符号
		"infty":    "∞",
		"partial":  "∂",
		"nabla":    "∇",
		"forall":   "∀",
		"exists":   "∃",
		"nexists":  "∄",
		"emptyset": "∅",
		"varnothing": "∅",
		"neg":      "¬",
		"lnot":     "¬",
		"land":     "∧",
		"lor":      "∨",
		"cap":      "∩",
		"cup":      "∪",
		"int":      "∫",
		"iint":     "∬",
		"iiint":    "∭",
		"oint":     "∮",
		"sum":      "∑",
		"prod":     "∏",
		"coprod":   "∐",
		"lim":      "lim",
		"limsup":   "lim sup",
		"liminf":   "lim inf",
		"max":      "max",
		"min":      "min",
		"sup":      "sup",
		"inf":      "inf",
		"sin":      "sin",
		"cos":      "cos",
		"tan":      "tan",
		"cot":      "cot",
		"sec":      "sec",
		"csc":      "csc",
		"arcsin":   "arcsin",
		"arccos":   "arccos",
		"arctan":   "arctan",
		"sinh":     "sinh",
		"cosh":     "cosh",
		"tanh":     "tanh",
		"log":      "log",
		"ln":       "ln",
		"exp":      "exp",
		"deg":      "deg",
		"det":      "det",
		"dim":      "dim",
		"ker":      "ker",
		"hom":      "hom",
		"arg":      "arg",
		"gcd":      "gcd",

		// 括号
		"lbrace":   "{",
		"rbrace":   "}",
		"langle":   "⟨",
		"rangle":   "⟩",
		"lceil":    "⌈",
		"rceil":    "⌉",
		"lfloor":   "⌊",
		"rfloor":   "⌋",
		"left":     "",
		"right":    "",

		// 其他
		"ldots":    "…",
		"cdots":    "⋯",
		"vdots":    "⋮",
		"ddots":    "⋱",
		"quad":     " ",
		"qquad":    "  ",
		"space":    " ",
	}

	if result, ok := commands[cmd]; ok {
		return result
	}
	return "\\" + cmd // 未知命令保持原样
}

// LaTeXToOMMLString 将LaTeX公式转换为OMML XML字符串
func LaTeXToOMMLString(latex string, isBlock bool) (string, error) {
	omath := LaTeXToOMML(latex)

	var result interface{}
	if isBlock {
		result = &OfficeMathPara{Math: omath}
	} else {
		result = omath
	}

	data, err := xml.Marshal(result)
	if err != nil {
		return "", err
	}

	return string(data), nil
}
