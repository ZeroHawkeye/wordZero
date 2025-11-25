package markdown

import (
	"path/filepath"
	"regexp"
	"strings"

	"github.com/ZeroHawkeye/wordZero/pkg/document"
	"github.com/yuin/goldmark/ast"

	// 添加goldmark扩展的AST节点支持
	extast "github.com/yuin/goldmark/extension/ast"

	// 添加数学公式支持
	mathjax "github.com/litao91/goldmark-mathjax"
)

// WordRenderer Word文档渲染器
type WordRenderer struct {
	doc       *document.Document
	opts      *ConvertOptions
	source    []byte
	listLevel int // 当前列表嵌套级别
}

// Render 渲染AST为Word文档
func (r *WordRenderer) Render(doc ast.Node) error {
	return ast.Walk(doc, func(node ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch n := node.(type) {
		case *ast.Document:
			// 文档根节点，继续处理子节点
			return ast.WalkContinue, nil

		case *ast.Heading:
			return r.renderHeading(n)

		case *ast.Paragraph:
			return r.renderParagraph(n)

		case *ast.List:
			return r.renderList(n)

		case *ast.ListItem:
			return r.renderListItem(n)

		case *ast.Blockquote:
			return r.renderBlockquote(n)

		case *ast.FencedCodeBlock:
			return r.renderCodeBlock(n)

		case *ast.CodeBlock:
			return r.renderCodeBlock(n)

		case *ast.ThematicBreak:
			return r.renderThematicBreak(n)

		case *ast.Text:
			// Text节点由父节点处理
			return ast.WalkSkipChildren, nil

		case *ast.Emphasis:
			// 强调节点由父节点处理
			return ast.WalkSkipChildren, nil

		case *ast.Link:
			// 链接节点由父节点处理
			return ast.WalkSkipChildren, nil

		case *ast.Image:
			return r.renderImage(n)

		// 表格支持
		case *extast.Table:
			if r.opts.EnableTables {
				return r.renderTable(n)
			}
			return ast.WalkContinue, nil

		case *extast.TableRow:
			// TableRow节点由Table处理
			return ast.WalkSkipChildren, nil

		case *extast.TableCell:
			// TableCell节点由Table处理
			return ast.WalkSkipChildren, nil

		// 任务列表支持
		case *extast.TaskCheckBox:
			if r.opts.EnableTaskList {
				return r.renderTaskCheckBox(n)
			}
			return ast.WalkContinue, nil

		default:
			// 检查是否为数学公式节点
			if r.opts.EnableMath {
				// 检查块级数学公式
				if node.Kind() == mathjax.KindMathBlock {
					return r.renderMathBlock(node)
				}
				// 检查行内数学公式
				if node.Kind() == mathjax.KindInlineMath {
					return r.renderInlineMath(node)
				}
			}
			// 对于不支持的节点类型，记录错误但继续处理
			if r.opts.ErrorCallback != nil {
				r.opts.ErrorCallback(NewConversionError("UnsupportedNode", "unsupported markdown node type", 0, 0, nil))
			}
			return ast.WalkContinue, nil
		}
	})
}

// renderHeading 渲染标题
func (r *WordRenderer) renderHeading(node *ast.Heading) (ast.WalkStatus, error) {
	text := r.extractTextContent(node)
	level := node.Level

	// 限制标题级别
	if level > 6 {
		level = 6
	}

	// 使用现有的API，确保兼容性
	if r.opts.GenerateTOC && level <= r.opts.TOCMaxLevel {
		// 复用现有的AddHeadingWithBookmark方法
		r.doc.AddHeadingWithBookmark(text, level, "")
	} else {
		// 复用现有的AddHeadingParagraph方法
		r.doc.AddHeadingParagraph(text, level)
	}

	return ast.WalkSkipChildren, nil
}

// renderParagraph 渲染段落
func (r *WordRenderer) renderParagraph(node *ast.Paragraph) (ast.WalkStatus, error) {
	// 检查段落是否为空
	if !node.HasChildren() {
		return ast.WalkSkipChildren, nil
	}

	// 创建段落
	para := r.doc.AddParagraph("")

	// 处理段落内容
	r.renderInlineContent(node, para)

	return ast.WalkSkipChildren, nil
}

// renderInlineContent 渲染内联内容（文本、强调、链接等）
func (r *WordRenderer) renderInlineContent(node ast.Node, para *document.Paragraph) {
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		switch n := child.(type) {
		case *ast.Text:
			text := string(n.Segment.Value(r.source))
			para.AddFormattedText(text, nil)
			
			// 处理软换行（单个\n）
			// goldmark将单个\n解析为多个Text节点，第一个节点的SoftLineBreak为true
			// 在Markdown中，软换行通常应该被渲染为空格
			if n.SoftLineBreak() {
				para.AddFormattedText(" ", nil)
			}

		case *ast.Emphasis:
			text := r.extractTextContent(n)
			// goldmark中，level=1是斜体，level=2是粗体
			if n.Level == 2 {
				// 使用粗体格式
				format := &document.TextFormat{Bold: true}
				para.AddFormattedText(text, format)
			} else {
				// 使用斜体格式
				format := &document.TextFormat{Italic: true}
				para.AddFormattedText(text, format)
			}

		case *ast.CodeSpan:
			text := r.extractTextContent(n)
			// 使用CodeChar样式的格式
			format := &document.TextFormat{
				FontFamily: "Consolas",
				FontColor:  "D73A49", // GitHub风格的红色
			}
			para.AddFormattedText(text, format)

		case *ast.Link:
			text := r.extractTextContent(n)
			// 简单处理链接，后续可以扩展为超链接
			format := &document.TextFormat{
				FontColor: "0000FF", // 蓝色
			}
			para.AddFormattedText(text, format)

		case *ast.Image:
			r.renderImageInline(n, para)
		case *extast.Strikethrough:
			// 处理删除线
			text := r.extractTextContent(n)
			format := &document.TextFormat{
				Strike: true,
			}
			para.AddFormattedText(text, format)

		default:
			// 检查是否为行内数学公式
			if r.opts.EnableMath && child.Kind() == mathjax.KindInlineMath {
				r.renderInlineMathToParagraph(child, para)
				continue
			}
			// 对于其他类型，尝试提取文本内容
			text := r.extractTextContent(n)
			if text != "" {
				para.AddFormattedText(text, nil)
			}
		}
	}
}

// renderList 渲染列表
func (r *WordRenderer) renderList(node *ast.List) (ast.WalkStatus, error) {
	r.listLevel++
	defer func() { r.listLevel-- }()

	// 处理列表项
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if listItem, ok := child.(*ast.ListItem); ok {
			r.renderListItem(listItem)
		}
	}

	return ast.WalkSkipChildren, nil
}

// renderListItem 渲染列表项
func (r *WordRenderer) renderListItem(node *ast.ListItem) (ast.WalkStatus, error) {
	// 检查是否包含任务复选框
	hasTaskCheckBox := false
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if _, ok := child.(*extast.TaskCheckBox); ok {
			hasTaskCheckBox = true
			break
		}
	}

	// 如果包含任务复选框且启用了任务列表，让TaskCheckBox节点处理
	if hasTaskCheckBox && r.opts.EnableTaskList {
		// 任务列表项将由TaskCheckBox节点处理
		return ast.WalkContinue, nil
	}

	// 普通列表项处理
	text := r.extractTextContent(node)

	// 简单的列表项处理，后续可以扩展为真正的列表格式
	// 这里暂时使用缩进和符号来模拟列表
	indent := strings.Repeat("  ", r.listLevel-1)
	bulletText := "• " + text

	r.doc.AddParagraph(indent + bulletText)

	return ast.WalkSkipChildren, nil
}

// renderBlockquote 渲染引用块
func (r *WordRenderer) renderBlockquote(node *ast.Blockquote) (ast.WalkStatus, error) {
	text := r.extractTextContent(node)

	// 创建引用段落，使用Quote样式
	para := r.doc.AddParagraph(text)
	para.SetStyle("Quote")

	return ast.WalkSkipChildren, nil
}

// renderCodeBlock 渲染代码块
func (r *WordRenderer) renderCodeBlock(node ast.Node) (ast.WalkStatus, error) {
	// 按行处理代码块内容，保持格式化状态
	lines := r.extractCodeBlockLines(node)

	// 为每行代码创建一个段落，保持换行和缩进
	for _, line := range lines {
		// 处理空行
		if strings.TrimSpace(line) == "" {
			para := r.doc.AddParagraph(" ") // 空行用空格表示
			para.SetStyle("CodeBlock")
			r.applyCodeBlockFormatting(para)
			continue
		}

		// 创建代码行段落
		para := r.doc.AddParagraph(line)
		para.SetStyle("CodeBlock")
		r.applyCodeBlockFormatting(para)
	}

	return ast.WalkSkipChildren, nil
}

// extractCodeBlockLines 按行提取代码块文本，保持格式
func (r *WordRenderer) extractCodeBlockLines(node ast.Node) []string {
	var lines []string

	for i := 0; i < node.Lines().Len(); i++ {
		line := node.Lines().At(i)
		lineText := string(line.Value(r.source))
		// 保持原始格式，包括空格和制表符
		lines = append(lines, lineText)
	}

	return lines
}

// applyCodeBlockFormatting 应用代码块格式
func (r *WordRenderer) applyCodeBlockFormatting(para *document.Paragraph) {
	// 应用额外的代码块格式
	if para.Properties == nil {
		para.Properties = &document.ParagraphProperties{}
	}

	// 设置左缩进（模拟code_template中的样式）
	para.Properties.Indentation = &document.Indentation{
		Left: "360", // 左缩进0.25英寸，与code_template保持一致
	}

	// 设置间距（段前段后间距）
	para.Properties.Spacing = &document.Spacing{
		Before: "60", // 3磅段前间距（减少间距，避免代码行之间空隙太大）
		After:  "60", // 3磅段后间距
	}

	// 设置对齐方式为左对齐
	para.Properties.Justification = &document.Justification{
		Val: "left",
	}
}

// renderThematicBreak 渲染分割线
func (r *WordRenderer) renderThematicBreak(node *ast.ThematicBreak) (ast.WalkStatus, error) {
	// 创建一个空段落用于显示分割线
	para := r.doc.AddParagraph("")

	// 设置水平分割线样式
	// 使用单线样式，中等粗细，黑色
	para.SetHorizontalRule(document.BorderStyleSingle, 12, "000000")

	// 设置段落间距，使分割线在视觉上更明显
	para.SetSpacing(&document.SpacingConfig{
		BeforePara: 6, // 段前6磅间距
		AfterPara:  6, // 段后6磅间距
	})

	return ast.WalkSkipChildren, nil
}

// renderImage 渲染图片
func (r *WordRenderer) renderImage(node *ast.Image) (ast.WalkStatus, error) {
	// 获取图片路径
	src := string(node.Destination)
	alt := r.extractTextContent(node)

	// 处理相对路径
	if !filepath.IsAbs(src) && r.opts.ImageBasePath != "" {
		src = filepath.Join(r.opts.ImageBasePath, src)
	}

	// 尝试添加图片，如果失败则添加替代文本
	// 这里需要后续完善图片处理逻辑
	if alt != "" {
		r.doc.AddParagraph("[图片: " + alt + "]")
	} else {
		r.doc.AddParagraph("[图片: " + src + "]")
	}

	return ast.WalkSkipChildren, nil
}

// renderImageInline 渲染内联图片
func (r *WordRenderer) renderImageInline(node *ast.Image, para *document.Paragraph) {
	src := string(node.Destination)
	alt := r.extractTextContent(node)

	// 处理相对路径
	if !filepath.IsAbs(src) && r.opts.ImageBasePath != "" {
		src = filepath.Join(r.opts.ImageBasePath, src)
	}

	// 内联图片暂时用文本替代
	if alt != "" {
		para.AddFormattedText("[图片: "+alt+"]", nil)
	} else {
		para.AddFormattedText("[图片: "+src+"]", nil)
	}
}

// extractTextContent 提取节点的文本内容
func (r *WordRenderer) extractTextContent(node ast.Node) string {
	var buf strings.Builder
	r.extractTextContentRecursive(node, &buf)
	return buf.String()
}

// extractTextContentRecursive 递归提取文本内容
func (r *WordRenderer) extractTextContentRecursive(node ast.Node, buf *strings.Builder) {
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		switch n := child.(type) {
		case *ast.Text:
			buf.Write(n.Segment.Value(r.source))
		default:
			r.extractTextContentRecursive(child, buf)
		}
	}
}

// cleanText 清理文本内容
func (r *WordRenderer) cleanText(text string) string {
	// 移除多余的空白字符
	re := regexp.MustCompile(`\s+`)
	text = re.ReplaceAllString(text, " ")
	return strings.TrimSpace(text)
}

// renderTable 渲染表格
func (r *WordRenderer) renderTable(node *extast.Table) (ast.WalkStatus, error) {
	// 收集表格数据
	var tableData [][]string
	var alignments []extast.Alignment
	var emphases [][]int

	// 遍历表头
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if row, ok := child.(*extast.TableHeader); ok {
			var rowData []string
			var rowEmphasis []int
			// 遍历表头单元格
			for cellChild := row.FirstChild(); cellChild != nil; cellChild = cellChild.NextSibling() {
				if cell, ok := cellChild.(*extast.TableCell); ok {
					cellText := r.extractTextContent(cell)
					rowData = append(rowData, cellText)
					//表头默认粗体
					rowEmphasis = append(rowEmphasis, 2)
				}
			}
			tableData = append(tableData, rowData)
			emphases = append(emphases, rowEmphasis)
		}
	}

	// 遍历表格行
	for child := node.FirstChild(); child != nil; child = child.NextSibling() {
		if row, ok := child.(*extast.TableRow); ok {
			var rowData []string
			var rowEmphasis []int
			if len(alignments) == 0 {
				// 从第一行获取对齐方式
				alignments = row.Alignments
			}

			// 遍历单元格
			for cellChild := row.FirstChild(); cellChild != nil; cellChild = cellChild.NextSibling() {
				if cell, ok := cellChild.(*extast.TableCell); ok {
					cellText := r.extractTextContent(cell)
					rowData = append(rowData, cellText)
					emphasis := extractCellEmphasis(cell)
					rowEmphasis = append(rowEmphasis, emphasis)
				}
			}
			tableData = append(tableData, rowData)
			emphases = append(emphases, rowEmphasis)
		}
	}

	// 如果没有数据，跳过
	if len(tableData) == 0 {
		return ast.WalkSkipChildren, nil
	}

	// 计算列数
	cols := 0
	for _, row := range tableData {
		if len(row) > cols {
			cols = len(row)
		}
	}

	// 创建表格配置
	config := &document.TableConfig{
		Rows:     len(tableData),
		Cols:     cols,
		Width:    9000, // 默认宽度（磅）
		Data:     tableData,
		Emphases: emphases,
	}

	// 添加表格到文档
	table, err := r.doc.AddTable(config)
	if err != nil && r.opts.ErrorCallback != nil {
		r.opts.ErrorCallback(NewConversionError("AddTable", err.Error(), 0, 0, err))
	}
	if table != nil {
		// 设置表头样式（如果有的话）
		if len(tableData) > 0 {
			// 第一行设为表头样式
			err := table.SetRowAsHeader(0, true)
			if err != nil && r.opts.ErrorCallback != nil {
				r.opts.ErrorCallback(NewConversionError("TableHeader", "failed to set table header", 0, 0, err))
			}
		}

		// 根据对齐方式设置单元格对齐
		for rowIdx, row := range tableData {
			for colIdx := range row {
				if colIdx < len(alignments) {
					var align document.CellAlignment
					switch alignments[colIdx] {
					case extast.AlignLeft:
						align = document.CellAlignLeft
					case extast.AlignCenter:
						align = document.CellAlignCenter
					case extast.AlignRight:
						align = document.CellAlignRight
					default:
						align = document.CellAlignLeft
					}

					format := &document.CellFormat{
						HorizontalAlign: align,
					}
					err := table.SetCellFormat(rowIdx, colIdx, format)
					if err != nil && r.opts.ErrorCallback != nil {
						r.opts.ErrorCallback(NewConversionError("CellFormat", "failed to set cell format", rowIdx, colIdx, err))
					}
				}
			}
		}
	}

	return ast.WalkSkipChildren, nil
}

// 处理单元格样式
func extractCellEmphasis(cell *extast.TableCell) int {
	format := 0 // 0表示无格式
	// 遍历单元格内容
	ast.Walk(cell, func(n ast.Node, entering bool) (ast.WalkStatus, error) {
		if !entering {
			return ast.WalkContinue, nil
		}

		switch node := n.(type) {
		case *ast.Emphasis:
			// 处理强调文本（粗体或斜体）
			format = node.Level
		}

		return ast.WalkContinue, nil
	})

	return format
}

// renderTaskCheckBox 渲染任务列表复选框 ✨ 新增功能
func (r *WordRenderer) renderTaskCheckBox(node *extast.TaskCheckBox) (ast.WalkStatus, error) {
	// 获取复选框状态
	checked := node.IsChecked

	// 根据状态选择符号
	var checkSymbol string
	if checked {
		checkSymbol = "☑" // 选中的复选框
	} else {
		checkSymbol = "☐" // 未选中的复选框
	}

	// 创建一个段落来包含复选框
	para := r.doc.AddParagraph("")

	// 添加复选框符号
	para.AddFormattedText(checkSymbol+" ", nil)

	// 处理任务项文本（通常是父级ListItem中的其他内容）
	// 注意：TaskCheckBox通常是ListItem的第一个子元素
	parent := node.Parent()
	if parent != nil {
		// 提取除TaskCheckBox外的其他文本内容
		r.renderTaskItemContent(parent, para, node)
	}

	return ast.WalkSkipChildren, nil
}

// renderTaskItemContent 渲染任务项内容（除复选框外的文本）
func (r *WordRenderer) renderTaskItemContent(parent ast.Node, para *document.Paragraph, skipNode ast.Node) {
	for child := parent.FirstChild(); child != nil; child = child.NextSibling() {
		// 跳过复选框节点本身
		if child == skipNode {
			continue
		}

		switch n := child.(type) {
		case *ast.Text:
			text := string(n.Segment.Value(r.source))
			para.AddFormattedText(text, nil)
			
			// 处理软换行（单个\n）
			if n.SoftLineBreak() {
				para.AddFormattedText(" ", nil)
			}
		case *ast.Emphasis:
			text := r.extractTextContent(n)
			if n.Level == 2 {
				format := &document.TextFormat{Bold: true}
				para.AddFormattedText(text, format)
			} else {
				format := &document.TextFormat{Italic: true}
				para.AddFormattedText(text, format)
			}
		case *ast.CodeSpan:
			text := r.extractTextContent(n)
			format := &document.TextFormat{
				FontFamily: "Consolas",
			}
			para.AddFormattedText(text, format)
		case *ast.Link:
			text := r.extractTextContent(n)
			format := &document.TextFormat{
				FontColor: "0000FF", // 蓝色
			}
			para.AddFormattedText(text, format)
		default:
			// 检查是否为行内数学公式
			if r.opts.EnableMath && child.Kind() == mathjax.KindInlineMath {
				r.renderInlineMathToParagraph(child, para)
				continue
			}
			// 对于其他类型，尝试提取文本内容
			text := r.extractTextContent(n)
			if text != "" {
				para.AddFormattedText(text, nil)
			}
		}
	}
}

// renderInlineMathToParagraph 将行内数学公式渲染到段落中
func (r *WordRenderer) renderInlineMathToParagraph(node ast.Node, para *document.Paragraph) {
	latex := r.extractMathContent(node)
	para.AddFormattedText(latex, &document.TextFormat{
		FontFamily: "Cambria Math",
	})
}

// renderMathBlock 渲染块级数学公式
func (r *WordRenderer) renderMathBlock(node ast.Node) (ast.WalkStatus, error) {
	// 提取LaTeX内容
	latex := r.extractMathContent(node)

	// 创建包含公式的段落
	para := r.doc.AddParagraph("")
	para.SetAlignment(document.AlignCenter)

	// 添加公式内容（使用Cambria Math字体以获得更好的数学符号显示）
	para.AddFormattedText(latex, &document.TextFormat{
		FontFamily: "Cambria Math",
		FontSize:   12,
	})

	return ast.WalkSkipChildren, nil
}

// renderInlineMath 渲染行内数学公式
func (r *WordRenderer) renderInlineMath(node ast.Node) (ast.WalkStatus, error) {
	// 行内公式通常由父节点处理（在renderInlineContent中）
	// 如果直接调用此方法，创建一个新段落
	para := r.doc.AddParagraph("")
	r.renderInlineMathToParagraph(node, para)

	return ast.WalkSkipChildren, nil
}

// extractMathContent 从数学节点中提取LaTeX内容
func (r *WordRenderer) extractMathContent(node ast.Node) string {
	var content strings.Builder

	// 检查是否为块级节点（MathBlock）
	if node.Kind() == mathjax.KindMathBlock {
		// MathBlock是块级节点，可以安全地访问Lines()
		if blockNode, ok := node.(ast.Node); ok {
			lines := blockNode.Lines()
			if lines != nil {
				for i := 0; i < lines.Len(); i++ {
					line := lines.At(i)
					content.Write(line.Value(r.source))
				}
			}
		}
	}

	// 如果没有行内容（可能是行内公式），尝试从子节点提取
	if content.Len() == 0 {
		for child := node.FirstChild(); child != nil; child = child.NextSibling() {
			if text, ok := child.(*ast.Text); ok {
				content.Write(text.Segment.Value(r.source))
			}
		}
	}

	latex := strings.TrimSpace(content.String())

	// 转换常见的LaTeX命令为Unicode字符以获得更好的显示效果
	latex = convertLaTeXToDisplay(latex)

	return latex
}

// convertLaTeXToDisplay 将LaTeX命令转换为可显示的Unicode字符
func convertLaTeXToDisplay(latex string) string {
	// 处理分数：\frac{a}{b} -> a/b 或 (a)/(b)
	fracPattern := regexp.MustCompile(`\\frac\s*\{([^{}]*(?:\{[^{}]*\}[^{}]*)*)\}\s*\{([^{}]*(?:\{[^{}]*\}[^{}]*)*)\}`)
	for fracPattern.MatchString(latex) {
		latex = fracPattern.ReplaceAllStringFunc(latex, func(match string) string {
			parts := fracPattern.FindStringSubmatch(match)
			if len(parts) == 3 {
				num := convertLaTeXToDisplay(parts[1])
				den := convertLaTeXToDisplay(parts[2])
				// 使用分数斜线字符来表示分数
				return "(" + num + ")/(" + den + ")"
			}
			return match
		})
	}

	// 处理根号：\sqrt{x} -> √x, \sqrt[n]{x} -> ⁿ√x
	sqrtPattern := regexp.MustCompile(`\\sqrt\s*(?:\[([^\]]*)\])?\s*\{([^{}]*(?:\{[^{}]*\}[^{}]*)*)\}`)
	for sqrtPattern.MatchString(latex) {
		latex = sqrtPattern.ReplaceAllStringFunc(latex, func(match string) string {
			parts := sqrtPattern.FindStringSubmatch(match)
			if len(parts) == 3 {
				deg := parts[1]
				content := convertLaTeXToDisplay(parts[2])
				if deg == "" {
					return "√(" + content + ")"
				}
				// 将根指数转换为上标
				degSup := convertToSuperscript(deg)
				return degSup + "√(" + content + ")"
			}
			return match
		})
	}

	// 处理上标：x^{n} -> xⁿ 或 x^n -> xⁿ
	supBracePattern := regexp.MustCompile(`\^\{([^{}]*)\}`)
	latex = supBracePattern.ReplaceAllStringFunc(latex, func(match string) string {
		parts := supBracePattern.FindStringSubmatch(match)
		if len(parts) == 2 {
			return convertToSuperscript(parts[1])
		}
		return match
	})
	supSimplePattern := regexp.MustCompile(`\^([a-zA-Z0-9])`)
	latex = supSimplePattern.ReplaceAllStringFunc(latex, func(match string) string {
		parts := supSimplePattern.FindStringSubmatch(match)
		if len(parts) == 2 {
			return convertToSuperscript(parts[1])
		}
		return match
	})

	// 处理下标：x_{n} -> xₙ 或 x_n -> xₙ
	subBracePattern := regexp.MustCompile(`_\{([^{}]*)\}`)
	latex = subBracePattern.ReplaceAllStringFunc(latex, func(match string) string {
		parts := subBracePattern.FindStringSubmatch(match)
		if len(parts) == 2 {
			return convertToSubscript(parts[1])
		}
		return match
	})
	subSimplePattern := regexp.MustCompile(`_([a-zA-Z0-9])`)
	latex = subSimplePattern.ReplaceAllStringFunc(latex, func(match string) string {
		parts := subSimplePattern.FindStringSubmatch(match)
		if len(parts) == 2 {
			return convertToSubscript(parts[1])
		}
		return match
	})

	// LaTeX命令替换映射
	replacements := map[string]string{
		// 希腊字母（小写）
		`\alpha`:   "α",
		`\beta`:    "β",
		`\gamma`:   "γ",
		`\delta`:   "δ",
		`\epsilon`: "ε",
		`\zeta`:    "ζ",
		`\eta`:     "η",
		`\theta`:   "θ",
		`\iota`:    "ι",
		`\kappa`:   "κ",
		`\lambda`:  "λ",
		`\mu`:      "μ",
		`\nu`:      "ν",
		`\xi`:      "ξ",
		`\pi`:      "π",
		`\rho`:     "ρ",
		`\sigma`:   "σ",
		`\tau`:     "τ",
		`\upsilon`: "υ",
		`\phi`:     "φ",
		`\chi`:     "χ",
		`\psi`:     "ψ",
		`\omega`:   "ω",

		// 希腊字母（大写）
		`\Alpha`:   "Α",
		`\Beta`:    "Β",
		`\Gamma`:   "Γ",
		`\Delta`:   "Δ",
		`\Epsilon`: "Ε",
		`\Zeta`:    "Ζ",
		`\Eta`:     "Η",
		`\Theta`:   "Θ",
		`\Iota`:    "Ι",
		`\Kappa`:   "Κ",
		`\Lambda`:  "Λ",
		`\Mu`:      "Μ",
		`\Nu`:      "Ν",
		`\Xi`:      "Ξ",
		`\Pi`:      "Π",
		`\Rho`:     "Ρ",
		`\Sigma`:   "Σ",
		`\Tau`:     "Τ",
		`\Upsilon`: "Υ",
		`\Phi`:     "Φ",
		`\Chi`:     "Χ",
		`\Psi`:     "Ψ",
		`\Omega`:   "Ω",

		// 运算符
		`\times`:  "×",
		`\div`:    "÷",
		`\pm`:     "±",
		`\mp`:     "∓",
		`\cdot`:   "·",
		`\ast`:    "∗",
		`\star`:   "⋆",
		`\circ`:   "∘",
		`\bullet`: "•",
		`\oplus`:  "⊕",
		`\ominus`: "⊖",
		`\otimes`: "⊗",

		// 关系符号
		`\leq`:      "≤",
		`\le`:       "≤",
		`\geq`:      "≥",
		`\ge`:       "≥",
		`\neq`:      "≠",
		`\ne`:       "≠",
		`\approx`:   "≈",
		`\equiv`:    "≡",
		`\sim`:      "∼",
		`\simeq`:    "≃",
		`\cong`:     "≅",
		`\propto`:   "∝",
		`\ll`:       "≪",
		`\gg`:       "≫",
		`\subset`:   "⊂",
		`\supset`:   "⊃",
		`\subseteq`: "⊆",
		`\supseteq`: "⊇",
		`\in`:       "∈",
		`\notin`:    "∉",
		`\ni`:       "∋",

		// 箭头
		`\rightarrow`:     "→",
		`\leftarrow`:      "←",
		`\leftrightarrow`: "↔",
		`\Rightarrow`:     "⇒",
		`\Leftarrow`:      "⇐",
		`\Leftrightarrow`: "⇔",
		`\uparrow`:        "↑",
		`\downarrow`:      "↓",
		`\to`:             "→",
		`\gets`:           "←",
		`\mapsto`:         "↦",

		// 杂项符号
		`\infty`:      "∞",
		`\partial`:    "∂",
		`\nabla`:      "∇",
		`\forall`:     "∀",
		`\exists`:     "∃",
		`\nexists`:    "∄",
		`\emptyset`:   "∅",
		`\varnothing`: "∅",
		`\neg`:        "¬",
		`\lnot`:       "¬",
		`\land`:       "∧",
		`\lor`:        "∨",
		`\cap`:        "∩",
		`\cup`:        "∪",
		`\int`:        "∫",
		`\iint`:       "∬",
		`\iiint`:      "∭",
		`\oint`:       "∮",
		`\sum`:        "∑",
		`\prod`:       "∏",
		`\coprod`:     "∐",

		// 省略号
		`\ldots`: "…",
		`\cdots`: "⋯",
		`\vdots`: "⋮",
		`\ddots`: "⋱",

		// 空格
		`\quad`:  " ",
		`\qquad`: "  ",
		`\,`:     " ",
		`\;`:     " ",
		`\:`:     " ",
		`\ `:     " ",

		// 括号
		`\{`:      "{",
		`\}`:      "}",
		`\lbrace`: "{",
		`\rbrace`: "}",
		`\langle`: "⟨",
		`\rangle`: "⟩",
		`\lceil`:  "⌈",
		`\rceil`:  "⌉",
		`\lfloor`: "⌊",
		`\rfloor`: "⌋",
		`\left`:   "",
		`\right`:  "",
	}

	// 应用替换
	for latex_cmd, unicode := range replacements {
		latex = strings.ReplaceAll(latex, latex_cmd, unicode)
	}

	// 清理多余的花括号
	latex = strings.ReplaceAll(latex, "{", "")
	latex = strings.ReplaceAll(latex, "}", "")

	return latex
}

// convertToSuperscript 将字符串转换为上标形式
func convertToSuperscript(s string) string {
	superscripts := map[rune]rune{
		'0': '⁰', '1': '¹', '2': '²', '3': '³', '4': '⁴',
		'5': '⁵', '6': '⁶', '7': '⁷', '8': '⁸', '9': '⁹',
		'+': '⁺', '-': '⁻', '=': '⁼', '(': '⁽', ')': '⁾',
		'a': 'ᵃ', 'b': 'ᵇ', 'c': 'ᶜ', 'd': 'ᵈ', 'e': 'ᵉ',
		'f': 'ᶠ', 'g': 'ᵍ', 'h': 'ʰ', 'i': 'ⁱ', 'j': 'ʲ',
		'k': 'ᵏ', 'l': 'ˡ', 'm': 'ᵐ', 'n': 'ⁿ', 'o': 'ᵒ',
		'p': 'ᵖ', 'r': 'ʳ', 's': 'ˢ', 't': 'ᵗ', 'u': 'ᵘ',
		'v': 'ᵛ', 'w': 'ʷ', 'x': 'ˣ', 'y': 'ʸ', 'z': 'ᶻ',
	}

	var result strings.Builder
	for _, r := range s {
		if sup, ok := superscripts[r]; ok {
			result.WriteRune(sup)
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}

// convertToSubscript 将字符串转换为下标形式
func convertToSubscript(s string) string {
	subscripts := map[rune]rune{
		'0': '₀', '1': '₁', '2': '₂', '3': '₃', '4': '₄',
		'5': '₅', '6': '₆', '7': '₇', '8': '₈', '9': '₉',
		'+': '₊', '-': '₋', '=': '₌', '(': '₍', ')': '₎',
		'a': 'ₐ', 'e': 'ₑ', 'h': 'ₕ', 'i': 'ᵢ', 'j': 'ⱼ',
		'k': 'ₖ', 'l': 'ₗ', 'm': 'ₘ', 'n': 'ₙ', 'o': 'ₒ',
		'p': 'ₚ', 'r': 'ᵣ', 's': 'ₛ', 't': 'ₜ', 'u': 'ᵤ',
		'v': 'ᵥ', 'x': 'ₓ',
	}

	var result strings.Builder
	for _, r := range s {
		if sub, ok := subscripts[r]; ok {
			result.WriteRune(sub)
		} else {
			result.WriteRune(r)
		}
	}
	return result.String()
}
