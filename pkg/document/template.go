// Package document 模板功能实现
package document

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// 模板相关错误
var (
	// ErrTemplateNotFound 模板未找到
	ErrTemplateNotFound = NewDocumentError("template_not_found", fmt.Errorf("template not found"), "")

	// ErrTemplateSyntaxError 模板语法错误
	ErrTemplateSyntaxError = NewDocumentError("template_syntax_error", fmt.Errorf("template syntax error"), "")

	// ErrTemplateRenderError 模板渲染错误
	ErrTemplateRenderError = NewDocumentError("template_render_error", fmt.Errorf("template render error"), "")

	// ErrInvalidTemplateData 无效模板数据
	ErrInvalidTemplateData = NewDocumentError("invalid_template_data", fmt.Errorf("invalid template data"), "")

	// ErrBlockNotFound 块未找到
	ErrBlockNotFound = NewDocumentError("block_not_found", fmt.Errorf("block not found"), "")

	// ErrInvalidBlockDefinition 无效块定义
	ErrInvalidBlockDefinition = NewDocumentError("invalid_block_definition", fmt.Errorf("invalid block definition"), "")
)

// TemplateEngine 模板引擎
type TemplateEngine struct {
	cache    map[string]*Template // 模板缓存
	mutex    sync.RWMutex         // 读写锁
	basePath string               // 基础路径
}

// Template 模板结构
type Template struct {
	Name          string                    // 模板名称
	Content       string                    // 模板内容
	BaseDoc       *Document                 // 基础文档
	Variables     map[string]string         // 模板变量
	Blocks        []*TemplateBlock          // 模板块列表
	Parent        *Template                 // 父模板（用于继承）
	DefinedBlocks map[string]*TemplateBlock // 定义的块映射
}

// TemplateBlock 模板块
type TemplateBlock struct {
	Type           string                 // 块类型：variable, if, each, inherit, block, image
	Name           string                 // 块名称（block类型使用）
	Content        string                 // 块内容
	Condition      string                 // 条件（if块使用）
	Variable       string                 // 变量名（each块使用）
	Children       []*TemplateBlock       // 子块
	Data           map[string]interface{} // 块数据
	DefaultContent string                 // 默认内容（用于可选重写）
	IsOverridden   bool                   // 是否被重写
}

// TemplateData 模板数据
type TemplateData struct {
	Variables  map[string]interface{}        // 变量数据
	Lists      map[string][]interface{}      // 列表数据
	Conditions map[string]bool               // 条件数据
	Images     map[string]*TemplateImageData // 图片数据
}

// TemplateImageData 模板图片数据
type TemplateImageData struct {
	FilePath string       // 图片文件路径
	Data     []byte       // 图片二进制数据（优先使用）
	Config   *ImageConfig // 图片配置（大小、位置、样式等）
	AltText  string       // 图片替代文字
	Title    string       // 图片标题
}

// NewTemplateEngine 创建新的模板引擎
func NewTemplateEngine() *TemplateEngine {
	return &TemplateEngine{
		cache: make(map[string]*Template),
		mutex: sync.RWMutex{},
	}
}

// SetBasePath 设置模板基础路径
func (te *TemplateEngine) SetBasePath(path string) {
	te.mutex.Lock()
	defer te.mutex.Unlock()
	te.basePath = path
}

// LoadTemplate 从字符串加载模板
func (te *TemplateEngine) LoadTemplate(name, content string) (*Template, error) {
	te.mutex.Lock()
	defer te.mutex.Unlock()

	template := &Template{
		Name:          name,
		Content:       content,
		Variables:     make(map[string]string),
		Blocks:        make([]*TemplateBlock, 0),
		DefinedBlocks: make(map[string]*TemplateBlock),
	}

	// 解析模板内容
	if err := te.parseTemplate(template); err != nil {
		return nil, WrapErrorWithContext("load_template", err, name)
	}

	// 缓存模板
	te.cache[name] = template

	return template, nil
}

// LoadTemplateFromDocument 从现有文档创建模板
func (te *TemplateEngine) LoadTemplateFromDocument(name string, doc *Document) (*Template, error) {
	te.mutex.Lock()
	defer te.mutex.Unlock()

	// 从文档中提取模板内容
	content, err := te.extractTemplateContentFromDocument(doc)
	if err != nil {
		return nil, WrapErrorWithContext("load_template_from_document", err, name)
	}

	template := &Template{
		Name:          name,
		Content:       content,
		BaseDoc:       doc,
		Variables:     make(map[string]string),
		Blocks:        make([]*TemplateBlock, 0),
		DefinedBlocks: make(map[string]*TemplateBlock),
	}

	// 解析模板内容
	if err := te.parseTemplate(template); err != nil {
		return nil, WrapErrorWithContext("load_template_from_document", err, name)
	}

	// 缓存模板
	te.cache[name] = template

	return template, nil
}

// GetTemplate 获取缓存的模板
func (te *TemplateEngine) GetTemplate(name string) (*Template, error) {
	te.mutex.RLock()
	defer te.mutex.RUnlock()

	if template, exists := te.cache[name]; exists {
		return template, nil
	}

	return nil, WrapErrorWithContext("get_template", ErrTemplateNotFound.Cause, name)
}

// getTemplateInternal 获取缓存的模板（内部方法，不加锁）
func (te *TemplateEngine) getTemplateInternal(name string) (*Template, error) {
	if template, exists := te.cache[name]; exists {
		return template, nil
	}

	return nil, WrapErrorWithContext("get_template", ErrTemplateNotFound.Cause, name)
}

// ClearCache 清空模板缓存
func (te *TemplateEngine) ClearCache() {
	te.mutex.Lock()
	defer te.mutex.Unlock()
	te.cache = make(map[string]*Template)
}

// RemoveTemplate 移除指定模板
func (te *TemplateEngine) RemoveTemplate(name string) {
	te.mutex.Lock()
	defer te.mutex.Unlock()
	delete(te.cache, name)
}

// parseTemplate 解析模板内容
func (te *TemplateEngine) parseTemplate(template *Template) error {
	content := template.Content

	// 解析变量: {{变量名}}
	varPattern := regexp.MustCompile(`\{\{(\w+)\}\}`)
	varMatches := varPattern.FindAllStringSubmatch(content, -1)
	for _, match := range varMatches {
		if len(match) >= 2 {
			varName := match[1]
			template.Variables[varName] = ""
		}
	}

	// 解析块定义: {{#block "blockName"}}...{{/block}}
	blockPattern := regexp.MustCompile(`(?s)\{\{#block\s+"([^"]+)"\}\}(.*?)\{\{/block\}\}`)
	blockMatches := blockPattern.FindAllStringSubmatch(content, -1)
	for _, match := range blockMatches {
		if len(match) >= 3 {
			blockName := match[1]
			blockContent := match[2]

			block := &TemplateBlock{
				Type:           "block",
				Name:           blockName,
				Content:        blockContent,
				DefaultContent: blockContent,
				Children:       make([]*TemplateBlock, 0),
			}

			template.Blocks = append(template.Blocks, block)
			template.DefinedBlocks[blockName] = block
		}
	}

	// 解析条件语句: {{#if 条件}}...{{/if}} (修复：添加 (?s) 标志以匹配换行符)
	ifPattern := regexp.MustCompile(`(?s)\{\{#if\s+(\w+)\}\}(.*?)\{\{/if\}\}`)
	ifMatches := ifPattern.FindAllStringSubmatch(content, -1)
	for _, match := range ifMatches {
		if len(match) >= 3 {
			condition := match[1]
			blockContent := match[2]

			block := &TemplateBlock{
				Type:      "if",
				Condition: condition,
				Content:   blockContent,
				Children:  make([]*TemplateBlock, 0),
			}

			template.Blocks = append(template.Blocks, block)
		}
	}

	// 解析循环语句: {{#each 列表}}...{{/each}} (修复：添加 (?s) 标志以匹配换行符)
	eachPattern := regexp.MustCompile(`(?s)\{\{#each\s+(\w+)\}\}(.*?)\{\{/each\}\}`)
	eachMatches := eachPattern.FindAllStringSubmatch(content, -1)
	for _, match := range eachMatches {
		if len(match) >= 3 {
			listVar := match[1]
			blockContent := match[2]

			block := &TemplateBlock{
				Type:     "each",
				Variable: listVar,
				Content:  blockContent,
				Children: make([]*TemplateBlock, 0),
			}

			template.Blocks = append(template.Blocks, block)
		}
	}

	// 解析图片占位符: {{#image imageName}}
	imagePattern := regexp.MustCompile(`\{\{#image\s+(\w+)\}\}`)
	imageMatches := imagePattern.FindAllStringSubmatch(content, -1)
	for _, match := range imageMatches {
		if len(match) >= 2 {
			imageName := match[1]

			block := &TemplateBlock{
				Type:     "image",
				Name:     imageName,
				Content:  match[0], // 保存完整的占位符文本
				Children: make([]*TemplateBlock, 0),
			}

			template.Blocks = append(template.Blocks, block)
		}
	}

	// 解析继承: {{extends "base_template"}}
	extendsPattern := regexp.MustCompile(`\{\{extends\s+"([^"]+)"\}\}`)
	extendsMatches := extendsPattern.FindStringSubmatch(content)
	if len(extendsMatches) >= 2 {
		baseName := extendsMatches[1]
		baseTemplate, err := te.getTemplateInternal(baseName)
		if err == nil {
			template.Parent = baseTemplate
			// 处理块重写
			te.processBlockOverrides(template, baseTemplate)
		}
	}

	return nil
}

// processBlockOverrides 处理块重写
func (te *TemplateEngine) processBlockOverrides(childTemplate, parentTemplate *Template) {
	// 遍历子模板的块定义，检查是否重写父模板的块
	for blockName, childBlock := range childTemplate.DefinedBlocks {
		if parentBlock, exists := parentTemplate.DefinedBlocks[blockName]; exists {
			// 标记父模板块被重写
			parentBlock.IsOverridden = true
			parentBlock.Content = childBlock.Content
		}
	}

	// 递归处理父模板的父模板
	if parentTemplate.Parent != nil {
		te.processBlockOverrides(childTemplate, parentTemplate.Parent)
	}
}

// RenderToDocument 渲染模板到新文档
func (te *TemplateEngine) RenderToDocument(templateName string, data *TemplateData) (*Document, error) {
	template, err := te.GetTemplate(templateName)
	if err != nil {
		return nil, WrapErrorWithContext("render_to_document", err, templateName)
	}

	// 创建新文档
	var doc *Document
	if template.BaseDoc != nil {
		// 基于基础文档创建
		doc = te.cloneDocument(template.BaseDoc)
	} else {
		// 创建新文档
		doc = New()
	}

	// 渲染模板内容
	renderedContent, err := te.renderTemplate(template, data)
	if err != nil {
		return nil, WrapErrorWithContext("render_to_document", err, templateName)
	}

	// 将渲染内容应用到文档
	if err := te.applyRenderedContentToDocument(doc, renderedContent); err != nil {
		return nil, WrapErrorWithContext("render_to_document", err, templateName)
	}

	// 处理图片占位符
	if err := te.processImagePlaceholders(doc, data); err != nil {
		return nil, WrapErrorWithContext("render_to_document", err, templateName)
	}

	return doc, nil
}

// renderTemplate 渲染模板
func (te *TemplateEngine) renderTemplate(template *Template, data *TemplateData) (string, error) {
	var content string

	// 处理继承：如果有父模板，使用父模板作为基础
	if template.Parent != nil {
		// 渲染父模板作为基础内容
		parentContent, err := te.renderTemplate(template.Parent, data)
		if err != nil {
			return "", err
		}
		content = parentContent

		// 应用子模板的块重写到父模板内容中
		content = te.applyBlockOverrides(content, template)
	} else {
		// 没有父模板，直接使用当前模板内容
		content = template.Content
	}

	// 渲染块定义
	content = te.renderBlocks(content, template, data)

	// 渲染变量
	content = te.renderVariables(content, data.Variables)

	// 渲染循环语句（先处理循环，循环内部会处理条件语句）
	content = te.renderLoops(content, data.Lists)

	// 渲染条件语句（处理非循环内的条件语句）
	content = te.renderConditionals(content, data.Conditions)

	// 渲染图片占位符
	content = te.renderImages(content, data.Images)

	return content, nil
}

// applyBlockOverrides 将子模板的块重写应用到父模板内容中
func (te *TemplateEngine) applyBlockOverrides(content string, template *Template) string {
	// 将子模板的块内容替换父模板中对应的块占位符
	blockPattern := regexp.MustCompile(`(?s)\{\{#block\s+"([^"]+)"\}\}.*?\{\{/block\}\}`)

	return blockPattern.ReplaceAllStringFunc(content, func(match string) string {
		matches := blockPattern.FindStringSubmatch(match)
		if len(matches) >= 2 {
			blockName := matches[1]
			// 如果子模板中定义了这个块，使用子模板的内容
			if childBlock, exists := template.DefinedBlocks[blockName]; exists {
				return childBlock.Content
			}
		}
		return match // 保持原样
	})
}

// renderBlocks 渲染块定义
func (te *TemplateEngine) renderBlocks(content string, template *Template, data *TemplateData) string {
	blockPattern := regexp.MustCompile(`(?s)\{\{#block\s+"([^"]+)"\}\}(.*?)\{\{/block\}\}`)

	return blockPattern.ReplaceAllStringFunc(content, func(match string) string {
		matches := blockPattern.FindStringSubmatch(match)
		if len(matches) >= 3 {
			blockName := matches[1]
			blockContent := matches[2]

			// 检查是否有定义的块
			if block, exists := template.DefinedBlocks[blockName]; exists {
				// 如果块被重写，使用重写的内容，否则使用默认内容
				if block.IsOverridden {
					return block.Content
				}
				return block.DefaultContent
			}

			// 如果没有定义块，使用原始内容
			return blockContent
		}
		return match
	})
}

// renderVariables 渲染变量
func (te *TemplateEngine) renderVariables(content string, variables map[string]interface{}) string {
	varPattern := regexp.MustCompile(`\{\{(\w+)\}\}`)

	return varPattern.ReplaceAllStringFunc(content, func(match string) string {
		varName := varPattern.FindStringSubmatch(match)[1]
		if value, exists := variables[varName]; exists {
			return te.interfaceToString(value)
		}
		return match // 保持原样
	})
}

// renderConditionals 渲染条件语句（支持if-else语法）
func (te *TemplateEngine) renderConditionals(content string, conditions map[string]bool) string {
	ifElsePattern := regexp.MustCompile(`(?s)\{\{#if\s+(\w+)\}\}(.*?)\{\{/if\}\}`)

	return ifElsePattern.ReplaceAllStringFunc(content, func(match string) string {
		matches := ifElsePattern.FindStringSubmatch(match)
		if len(matches) >= 3 {
			condition := matches[1]
			blockContent := matches[2]

			// 检查是否有else部分
			elsePattern := regexp.MustCompile(`(?s)(.*?)\{\{else\}\}(.*?)`)
			elseMatches := elsePattern.FindStringSubmatch(blockContent)

			if len(elseMatches) >= 3 {
				// 有else部分
				ifContent := elseMatches[1]
				elseContent := elseMatches[2]

				if condValue, exists := conditions[condition]; exists && condValue {
					return ifContent
				} else {
					return elseContent
				}
			} else {
				// 没有else部分，按原逻辑处理
				if condValue, exists := conditions[condition]; exists && condValue {
					return blockContent
				}
			}
		}
		return "" // 条件不满足，返回空字符串
	})
}

// renderLoops 渲染循环语句
func (te *TemplateEngine) renderLoops(content string, lists map[string][]interface{}) string {
	// 使用栈式方法正确处理嵌套循环
	return te.renderLoopsNested(content, lists, 0)
}

// renderLoopsNested 使用递归方式处理嵌套循环
func (te *TemplateEngine) renderLoopsNested(content string, lists map[string][]interface{}, depth int) string {
	// 查找第一个 {{#each}} 标记
	eachStartPattern := regexp.MustCompile(`\{\{#each\s+(\w+)\}\}`)
	startMatch := eachStartPattern.FindStringIndex(content)

	if startMatch == nil {
		// 没有找到循环，直接返回
		return content
	}

	// 找到了循环开始标记，现在需要找到匹配的结束标记
	startPos := startMatch[0]
	listVarMatch := eachStartPattern.FindStringSubmatch(content[startPos:])
	if len(listVarMatch) < 2 {
		return content
	}

	listVar := listVarMatch[1]
	blockStart := startMatch[1] // {{#each xxx}} 之后的位置

	// 使用栈来找到匹配的 {{/each}}
	depth_counter := 1
	pos := blockStart
	blockEnd := -1

	for pos < len(content) {
		// 查找下一个 {{#each}} 或 {{/each}}
		nextEach := eachStartPattern.FindStringIndex(content[pos:])
		endPattern := regexp.MustCompile(`\{\{/each\}\}`)
		nextEnd := endPattern.FindStringIndex(content[pos:])

		if nextEnd == nil {
			// 没有找到结束标记，语法错误
			break
		}

		// 确定下一个是开始还是结束
		if nextEach != nil && nextEach[0] < nextEnd[0] {
			// 下一个是嵌套的开始标记
			depth_counter++
			pos = pos + nextEach[1]
		} else {
			// 下一个是结束标记
			depth_counter--
			if depth_counter == 0 {
				// 找到了匹配的结束标记
				blockEnd = pos + nextEnd[0]
				break
			}
			pos = pos + nextEnd[1]
		}
	}

	if blockEnd == -1 {
		// 没有找到匹配的结束标记
		return content
	}

	// 提取循环块内容
	blockContent := content[blockStart:blockEnd]

	// 处理循环
	var result strings.Builder

	// 添加循环之前的内容
	result.WriteString(content[:startPos])

	// 渲染循环
	if listData, exists := lists[listVar]; exists {
		for i, item := range listData {
			// 创建循环上下文变量
			loopContent := strings.ReplaceAll(blockContent, "{{this}}", te.interfaceToString(item))
			loopContent = strings.ReplaceAll(loopContent, "{{@index}}", strconv.Itoa(i))
			loopContent = strings.ReplaceAll(loopContent, "{{@first}}", strconv.FormatBool(i == 0))
			loopContent = strings.ReplaceAll(loopContent, "{{@last}}", strconv.FormatBool(i == len(listData)-1))

			// 如果item是map，处理属性访问
			if itemMap, ok := item.(map[string]interface{}); ok {
				// 首先处理嵌套的循环（在替换变量之前）
				// 为嵌套循环创建新的lists map，包含当前项的列表数据
				nestedLists := make(map[string][]interface{})
				for key, value := range itemMap {
					// 检查值是否是列表类型
					if listValue, ok := value.([]interface{}); ok {
						nestedLists[key] = listValue
					}
				}

				// 如果有嵌套列表，递归处理嵌套循环
				if len(nestedLists) > 0 {
					loopContent = te.renderLoopsNested(loopContent, nestedLists, depth+1)
				}

				// 然后替换普通变量
				for key, value := range itemMap {
					placeholder := fmt.Sprintf("{{%s}}", key)
					// 只替换非列表类型的值
					if _, isList := value.([]interface{}); !isList {
						loopContent = strings.ReplaceAll(loopContent, placeholder, te.interfaceToString(value))
					}
				}

				// 处理循环内部的条件语句
				loopContent = te.renderLoopConditionals(loopContent, itemMap)
			}

			result.WriteString(loopContent)
		}
	}

	// 添加循环之后的内容，并递归处理剩余内容中的其他循环
	remainingContent := content[blockEnd+len("{{/each}}"):]
	remainingContent = te.renderLoopsNested(remainingContent, lists, depth)
	result.WriteString(remainingContent)

	return result.String()
}

// renderLoopConditionals 渲染循环内部的条件语句（支持if-else语法）
func (te *TemplateEngine) renderLoopConditionals(content string, itemData map[string]interface{}) string {
	ifElsePattern := regexp.MustCompile(`(?s)\{\{#if\s+(\w+)\}\}(.*?)\{\{/if\}\}`)

	return ifElsePattern.ReplaceAllStringFunc(content, func(match string) string {
		matches := ifElsePattern.FindStringSubmatch(match)
		if len(matches) >= 3 {
			condition := matches[1]
			blockContent := matches[2]

			// 检查是否有else部分
			elsePattern := regexp.MustCompile(`(?s)(.*?)\{\{else\}\}(.*?)`)
			elseMatches := elsePattern.FindStringSubmatch(blockContent)

			var ifContent, elseContent string
			hasElse := false

			if len(elseMatches) >= 3 {
				// 有else部分
				ifContent = elseMatches[1]
				elseContent = elseMatches[2]
				hasElse = true
			} else {
				// 没有else部分
				ifContent = blockContent
			}

			// 检查条件是否在当前循环项的数据中
			if condValue, exists := itemData[condition]; exists {
				// 转换为布尔值
				conditionMet := false
				switch v := condValue.(type) {
				case bool:
					conditionMet = v
				case string:
					conditionMet = v == "true" || v == "1" || v == "yes" || v != ""
				case int:
					conditionMet = v != 0
				case int64:
					conditionMet = v != 0
				case float64:
					conditionMet = v != 0.0
				default:
					// 对于其他类型，如果不为nil就认为是true
					conditionMet = v != nil
				}

				if conditionMet {
					return ifContent
				} else if hasElse {
					return elseContent
				}
			} else if hasElse {
				// 条件不存在，返回else部分
				return elseContent
			}
		}
		return "" // 条件不满足且没有else，返回空字符串
	})
}

// interfaceToString 将interface{}转换为字符串
func (te *TemplateEngine) interfaceToString(value interface{}) string {
	if value == nil {
		return ""
	}

	switch v := value.(type) {
	case string:
		return v
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// ValidateTemplate 验证模板语法
func (te *TemplateEngine) ValidateTemplate(template *Template) error {
	content := template.Content

	// 检查括号配对
	if err := te.validateBrackets(content); err != nil {
		return WrapErrorWithContext("validate_template", err, template.Name)
	}

	// 检查块语句配对
	if err := te.validateBlockStatements(content); err != nil {
		return WrapErrorWithContext("validate_template", err, template.Name)
	}

	// 检查if语句配对
	if err := te.validateIfStatements(content); err != nil {
		return WrapErrorWithContext("validate_template", err, template.Name)
	}

	// 检查each语句配对
	if err := te.validateEachStatements(content); err != nil {
		return WrapErrorWithContext("validate_template", err, template.Name)
	}

	return nil
}

// validateBrackets 验证括号配对
func (te *TemplateEngine) validateBrackets(content string) error {
	openCount := strings.Count(content, "{{")
	closeCount := strings.Count(content, "}}")

	if openCount != closeCount {
		return NewValidationError("brackets", content, "mismatched template brackets")
	}

	return nil
}

// validateBlockStatements 验证块语句配对
func (te *TemplateEngine) validateBlockStatements(content string) error {
	blockCount := len(regexp.MustCompile(`\{\{#block\s+"[^"]+"\}\}`).FindAllString(content, -1))
	endblockCount := len(regexp.MustCompile(`\{\{/block\}\}`).FindAllString(content, -1))

	if blockCount != endblockCount {
		return NewValidationError("block_statements", content, "mismatched block/endblock statements")
	}

	return nil
}

// validateIfStatements 验证if语句配对
func (te *TemplateEngine) validateIfStatements(content string) error {
	ifCount := len(regexp.MustCompile(`\{\{#if\s+\w+\}\}`).FindAllString(content, -1))
	endifCount := len(regexp.MustCompile(`\{\{/if\}\}`).FindAllString(content, -1))

	if ifCount != endifCount {
		return NewValidationError("if_statements", content, "mismatched if/endif statements")
	}

	return nil
}

// validateEachStatements 验证each语句配对
func (te *TemplateEngine) validateEachStatements(content string) error {
	eachCount := len(regexp.MustCompile(`\{\{#each\s+\w+\}\}`).FindAllString(content, -1))
	endeachCount := len(regexp.MustCompile(`\{\{/each\}\}`).FindAllString(content, -1))

	if eachCount != endeachCount {
		return NewValidationError("each_statements", content, "mismatched each/endeach statements")
	}

	return nil
}

// documentToTemplateString 将文档转换为模板字符串
func (te *TemplateEngine) documentToTemplateString(doc *Document) (string, error) {
	// 这里不再转换为纯字符串，而是保持原始文档结构
	// 实际上我们应该直接在原文档上进行变量替换
	return "", nil // 将在新的方法中处理
}

// extractTemplateContentFromDocument 从文档中提取模板内容
func (te *TemplateEngine) extractTemplateContentFromDocument(doc *Document) (string, error) {
	var contentBuilder strings.Builder

	// 遍历文档元素，提取文本内容
	for _, element := range doc.Body.Elements {
		switch elem := element.(type) {
		case *Paragraph:
			// 提取段落中的文本
			for _, run := range elem.Runs {
				contentBuilder.WriteString(run.Text.Content)
			}
			contentBuilder.WriteString("\n")

		case *Table:
			// 暂时跳过表格，专注于段落中的模板语法
			// 表格中的模板语法将在RenderTemplateToDocument中处理
			continue
		}
	}

	return contentBuilder.String(), nil
}

// cloneDocument 深度复制文档所有元素和属性
func (te *TemplateEngine) cloneDocument(source *Document) *Document {
	// 创建新文档
	doc := New()

	// 深拷贝文档元素
	for _, element := range source.Body.Elements {
		switch elem := element.(type) {
		case *Paragraph:
			clonedPara := te.cloneParagraph(elem)
			doc.Body.Elements = append(doc.Body.Elements, clonedPara)

		case *Table:
			clonedTable := te.cloneTable(elem)
			doc.Body.Elements = append(doc.Body.Elements, clonedTable)

		default:
			// 其他类型暂时直接复制引用
			doc.Body.Elements = append(doc.Body.Elements, element)
		}
	}

	// 深拷贝样式管理器，确保模板渲染时的样式与原模板一致
	if source.styleManager != nil {
		doc.styleManager = source.styleManager.Clone()
		// 不再强制修改 Normal 样式的段落行距，避免覆盖模板自身的默认行距/段后设置。
		// 如需统一行距，请在模板中显式设置，而非由代码层面硬编码。
	}

	// 复制 styles.xml 等样式相关部件，确保 docDefaults 等信息完整保留
	if doc.parts == nil {
		doc.parts = make(map[string][]byte)
	}
	if data, ok := source.parts["word/styles.xml"]; ok {
		doc.parts["word/styles.xml"] = data
	}

	return doc
}

// cloneParagraph 深度复制段落
func (te *TemplateEngine) cloneParagraph(source *Paragraph) *Paragraph {
	newPara := &Paragraph{
		Properties: te.cloneParagraphProperties(source.Properties),
		Runs:       make([]Run, len(source.Runs)),
	}

	for i, run := range source.Runs {
		newPara.Runs[i] = te.cloneRun(&run)
	}

	return newPara
}

// cloneParagraphProperties 深度复制段落属性
func (te *TemplateEngine) cloneParagraphProperties(source *ParagraphProperties) *ParagraphProperties {
	if source == nil {
		return nil
	}

	props := &ParagraphProperties{}

	// 复制段落样式
	if source.ParagraphStyle != nil {
		props.ParagraphStyle = &ParagraphStyle{
			Val: source.ParagraphStyle.Val,
		}
	}

	// 复制编号属性
	if source.NumberingProperties != nil {
		props.NumberingProperties = &NumberingProperties{}
		if source.NumberingProperties.ILevel != nil {
			props.NumberingProperties.ILevel = &ILevel{Val: source.NumberingProperties.ILevel.Val}
		}
		if source.NumberingProperties.NumID != nil {
			props.NumberingProperties.NumID = &NumID{Val: source.NumberingProperties.NumID.Val}
		}
	}

	// 复制间距
	if source.Spacing != nil {
		props.Spacing = &Spacing{
			Before:   source.Spacing.Before,
			After:    source.Spacing.After,
			Line:     source.Spacing.Line,
			LineRule: source.Spacing.LineRule,
		}
	}

	// 复制对齐方式
	if source.Justification != nil {
		props.Justification = &Justification{
			Val: source.Justification.Val,
		}
	}

	// 复制缩进
	if source.Indentation != nil {
		props.Indentation = &Indentation{
			FirstLine: source.Indentation.FirstLine,
			Left:      source.Indentation.Left,
			Right:     source.Indentation.Right,
		}
	}

	// 复制制表符
	if source.Tabs != nil {
		props.Tabs = &Tabs{
			Tabs: make([]TabDef, len(source.Tabs.Tabs)),
		}
		for i, tab := range source.Tabs.Tabs {
			props.Tabs.Tabs[i] = TabDef{
				Val:    tab.Val,
				Leader: tab.Leader,
				Pos:    tab.Pos,
			}
		}
	}

	return props
}

// cloneRun 深度复制文本运行
func (te *TemplateEngine) cloneRun(source *Run) Run {
	newRun := Run{
		Properties: te.cloneRunProperties(source.Properties),
		Text:       Text{Content: source.Text.Content, Space: source.Text.Space},
	}

	// 复制图像（如果有）
	if source.Drawing != nil {
		// 暂时保持简单复制，图像的深度复制比较复杂
		newRun.Drawing = source.Drawing
	}

	// 复制域字符（如果有）
	if source.FieldChar != nil {
		newRun.FieldChar = source.FieldChar
	}

	// 复制指令文本（如果有）
	if source.InstrText != nil {
		newRun.InstrText = source.InstrText
	}

	return newRun
}

// cloneRunProperties 深度复制文本运行属性
func (te *TemplateEngine) cloneRunProperties(source *RunProperties) *RunProperties {
	if source == nil {
		return nil
	}

	props := &RunProperties{}

	// 复制粗体
	if source.Bold != nil {
		props.Bold = &Bold{}
	}

	// 复制复杂脚本粗体
	if source.BoldCs != nil {
		props.BoldCs = &BoldCs{}
	}

	// 复制斜体
	if source.Italic != nil {
		props.Italic = &Italic{}
	}

	// 复制复杂脚本斜体
	if source.ItalicCs != nil {
		props.ItalicCs = &ItalicCs{}
	}

	// 复制下划线
	if source.Underline != nil {
		props.Underline = &Underline{
			Val: source.Underline.Val,
		}
	}

	// 复制删除线
	if source.Strike != nil {
		props.Strike = &Strike{}
	}

	// 复制字体大小
	if source.FontSize != nil {
		props.FontSize = &FontSize{
			Val: source.FontSize.Val,
		}
	}

	// 复制复杂脚本字体大小
	if source.FontSizeCs != nil {
		props.FontSizeCs = &FontSizeCs{
			Val: source.FontSizeCs.Val,
		}
	}

	// 复制颜色
	if source.Color != nil {
		props.Color = &Color{
			Val: source.Color.Val,
		}
	}

	// 复制背景色
	if source.Highlight != nil {
		props.Highlight = &Highlight{
			Val: source.Highlight.Val,
		}
	}

	// 完整复制字体族属性，包括所有字体设置
	if source.FontFamily != nil {
		props.FontFamily = &FontFamily{
			ASCII:    source.FontFamily.ASCII,
			HAnsi:    source.FontFamily.HAnsi,
			EastAsia: source.FontFamily.EastAsia,
			CS:       source.FontFamily.CS,
			Hint:     source.FontFamily.Hint,
		}
	}

	return props
}

// cloneTable 深度复制表格
func (te *TemplateEngine) cloneTable(source *Table) *Table {
	newTable := &Table{
		Properties: te.cloneTableProperties(source.Properties),
		Grid:       te.cloneTableGrid(source.Grid),
		Rows:       make([]TableRow, len(source.Rows)),
	}

	for i, row := range source.Rows {
		newTable.Rows[i] = *te.cloneTableRow(&row)
	}

	return newTable
}

// cloneTableProperties 深度复制表格属性
func (te *TemplateEngine) cloneTableProperties(source *TableProperties) *TableProperties {
	if source == nil {
		Debug("克隆表格属性: 源属性为空")
		return nil
	}

	props := &TableProperties{}

	// 复制表格宽度
	if source.TableW != nil {
		props.TableW = &TableWidth{
			W:    source.TableW.W,
			Type: source.TableW.Type,
		}
	}

	// 复制表格对齐
	if source.TableJc != nil {
		props.TableJc = &TableJc{
			Val: source.TableJc.Val,
		}
	}

	// 复制表格外观
	if source.TableLook != nil {
		props.TableLook = &TableLook{
			Val:      source.TableLook.Val,
			FirstRow: source.TableLook.FirstRow,
			LastRow:  source.TableLook.LastRow,
			FirstCol: source.TableLook.FirstCol,
			LastCol:  source.TableLook.LastCol,
			NoHBand:  source.TableLook.NoHBand,
			NoVBand:  source.TableLook.NoVBand,
		}
	}

	// 复制表格样式
	if source.TableStyle != nil {
		props.TableStyle = &TableStyle{
			Val: source.TableStyle.Val,
		}
	}

	// 复制表格边框
	if source.TableBorders != nil {
		props.TableBorders = te.cloneTableBorders(source.TableBorders)
	}

	// 复制表格底纹
	if source.Shd != nil {
		props.Shd = &TableShading{
			Val:       source.Shd.Val,
			Color:     source.Shd.Color,
			Fill:      source.Shd.Fill,
			ThemeFill: source.Shd.ThemeFill,
		}
	}

	// 复制表格单元格边距
	if source.TableCellMar != nil {
		props.TableCellMar = te.cloneTableCellMargins(source.TableCellMar)
	}

	// 复制表格布局
	if source.TableLayout != nil {
		props.TableLayout = &TableLayoutType{
			Type: source.TableLayout.Type,
		}
	}

	// 复制表格缩进
	if source.TableInd != nil {
		props.TableInd = &TableIndentation{
			W:    source.TableInd.W,
			Type: source.TableInd.Type,
		}
	}

	return props
}

// cloneTableBorders 深度复制表格边框
func (te *TemplateEngine) cloneTableBorders(source *TableBorders) *TableBorders {
	if source == nil {
		return nil
	}

	borders := &TableBorders{}

	if source.Top != nil {
		borders.Top = &TableBorder{
			Val:        source.Top.Val,
			Sz:         source.Top.Sz,
			Space:      source.Top.Space,
			Color:      source.Top.Color,
			ThemeColor: source.Top.ThemeColor,
		}
	}

	if source.Left != nil {
		borders.Left = &TableBorder{
			Val:        source.Left.Val,
			Sz:         source.Left.Sz,
			Space:      source.Left.Space,
			Color:      source.Left.Color,
			ThemeColor: source.Left.ThemeColor,
		}
	}

	if source.Bottom != nil {
		borders.Bottom = &TableBorder{
			Val:        source.Bottom.Val,
			Sz:         source.Bottom.Sz,
			Space:      source.Bottom.Space,
			Color:      source.Bottom.Color,
			ThemeColor: source.Bottom.ThemeColor,
		}
	}

	if source.Right != nil {
		borders.Right = &TableBorder{
			Val:        source.Right.Val,
			Sz:         source.Right.Sz,
			Space:      source.Right.Space,
			Color:      source.Right.Color,
			ThemeColor: source.Right.ThemeColor,
		}
	}

	if source.InsideH != nil {
		borders.InsideH = &TableBorder{
			Val:        source.InsideH.Val,
			Sz:         source.InsideH.Sz,
			Space:      source.InsideH.Space,
			Color:      source.InsideH.Color,
			ThemeColor: source.InsideH.ThemeColor,
		}
	}

	if source.InsideV != nil {
		borders.InsideV = &TableBorder{
			Val:        source.InsideV.Val,
			Sz:         source.InsideV.Sz,
			Space:      source.InsideV.Space,
			Color:      source.InsideV.Color,
			ThemeColor: source.InsideV.ThemeColor,
		}
	}

	return borders
}

// cloneTableCellMargins 深度复制表格单元格边距
func (te *TemplateEngine) cloneTableCellMargins(source *TableCellMargins) *TableCellMargins {
	if source == nil {
		return nil
	}

	margins := &TableCellMargins{}

	if source.Top != nil {
		margins.Top = &TableCellSpace{
			W:    source.Top.W,
			Type: source.Top.Type,
		}
	}

	if source.Left != nil {
		margins.Left = &TableCellSpace{
			W:    source.Left.W,
			Type: source.Left.Type,
		}
	}

	if source.Bottom != nil {
		margins.Bottom = &TableCellSpace{
			W:    source.Bottom.W,
			Type: source.Bottom.Type,
		}
	}

	if source.Right != nil {
		margins.Right = &TableCellSpace{
			W:    source.Right.W,
			Type: source.Right.Type,
		}
	}

	return margins
}

// cloneTableGrid 深度复制表格网格
func (te *TemplateEngine) cloneTableGrid(source *TableGrid) *TableGrid {
	if source == nil {
		return nil
	}

	grid := &TableGrid{
		Cols: make([]TableGridCol, len(source.Cols)),
	}

	for i, col := range source.Cols {
		grid.Cols[i] = TableGridCol{
			W: col.W,
		}
	}

	return grid
}

// cloneTableCellMarginsCell 深度复制表格单元格边距（单元格级别）
func (te *TemplateEngine) cloneTableCellMarginsCell(source *TableCellMarginsCell) *TableCellMarginsCell {
	if source == nil {
		return nil
	}

	margins := &TableCellMarginsCell{}

	if source.Top != nil {
		margins.Top = &TableCellSpaceCell{
			W:    source.Top.W,
			Type: source.Top.Type,
		}
	}

	if source.Left != nil {
		margins.Left = &TableCellSpaceCell{
			W:    source.Left.W,
			Type: source.Left.Type,
		}
	}

	if source.Bottom != nil {
		margins.Bottom = &TableCellSpaceCell{
			W:    source.Bottom.W,
			Type: source.Bottom.Type,
		}
	}

	if source.Right != nil {
		margins.Right = &TableCellSpaceCell{
			W:    source.Right.W,
			Type: source.Right.Type,
		}
	}

	return margins
}

// cloneTableCellBorders 深度复制表格单元格边框
func (te *TemplateEngine) cloneTableCellBorders(source *TableCellBorders) *TableCellBorders {
	if source == nil {
		return nil
	}

	borders := &TableCellBorders{}

	if source.Top != nil {
		borders.Top = &TableCellBorder{
			Val:        source.Top.Val,
			Sz:         source.Top.Sz,
			Space:      source.Top.Space,
			Color:      source.Top.Color,
			ThemeColor: source.Top.ThemeColor,
		}
	}

	if source.Left != nil {
		borders.Left = &TableCellBorder{
			Val:        source.Left.Val,
			Sz:         source.Left.Sz,
			Space:      source.Left.Space,
			Color:      source.Left.Color,
			ThemeColor: source.Left.ThemeColor,
		}
	}

	if source.Bottom != nil {
		borders.Bottom = &TableCellBorder{
			Val:        source.Bottom.Val,
			Sz:         source.Bottom.Sz,
			Space:      source.Bottom.Space,
			Color:      source.Bottom.Color,
			ThemeColor: source.Bottom.ThemeColor,
		}
	}

	if source.Right != nil {
		borders.Right = &TableCellBorder{
			Val:        source.Right.Val,
			Sz:         source.Right.Sz,
			Space:      source.Right.Space,
			Color:      source.Right.Color,
			ThemeColor: source.Right.ThemeColor,
		}
	}

	if source.InsideH != nil {
		borders.InsideH = &TableCellBorder{
			Val:        source.InsideH.Val,
			Sz:         source.InsideH.Sz,
			Space:      source.InsideH.Space,
			Color:      source.InsideH.Color,
			ThemeColor: source.InsideH.ThemeColor,
		}
	}

	if source.InsideV != nil {
		borders.InsideV = &TableCellBorder{
			Val:        source.InsideV.Val,
			Sz:         source.InsideV.Sz,
			Space:      source.InsideV.Space,
			Color:      source.InsideV.Color,
			ThemeColor: source.InsideV.ThemeColor,
		}
	}

	if source.TL2BR != nil {
		borders.TL2BR = &TableCellBorder{
			Val:        source.TL2BR.Val,
			Sz:         source.TL2BR.Sz,
			Space:      source.TL2BR.Space,
			Color:      source.TL2BR.Color,
			ThemeColor: source.TL2BR.ThemeColor,
		}
	}

	if source.TR2BL != nil {
		borders.TR2BL = &TableCellBorder{
			Val:        source.TR2BL.Val,
			Sz:         source.TR2BL.Sz,
			Space:      source.TR2BL.Space,
			Color:      source.TR2BL.Color,
			ThemeColor: source.TR2BL.ThemeColor,
		}
	}

	return borders
}

// cloneTableRow 深度复制表格行
func (te *TemplateEngine) cloneTableRow(source *TableRow) *TableRow {
	newRow := &TableRow{
		Properties: te.cloneTableRowProperties(source.Properties),
		Cells:      make([]TableCell, len(source.Cells)),
	}

	for i, cell := range source.Cells {
		newRow.Cells[i] = te.cloneTableCell(&cell)
	}

	return newRow
}

// cloneTableRowProperties 深度复制表格行属性
func (te *TemplateEngine) cloneTableRowProperties(source *TableRowProperties) *TableRowProperties {
	if source == nil {
		return nil
	}

	props := &TableRowProperties{}

	// 复制行高
	if source.TableRowH != nil {
		props.TableRowH = &TableRowH{
			Val:   source.TableRowH.Val,
			HRule: source.TableRowH.HRule,
		}
	}

	// 复制禁止跨页分割
	if source.CantSplit != nil {
		props.CantSplit = &CantSplit{
			Val: source.CantSplit.Val,
		}
	}

	// 复制标题行重复
	if source.TblHeader != nil {
		props.TblHeader = &TblHeader{
			Val: source.TblHeader.Val,
		}
	}

	return props
}

// cloneTableCell 深度复制表格单元格
func (te *TemplateEngine) cloneTableCell(source *TableCell) TableCell {
	newCell := TableCell{
		Properties: te.cloneTableCellProperties(source.Properties),
		Paragraphs: make([]Paragraph, len(source.Paragraphs)),
	}

	for i, para := range source.Paragraphs {
		newCell.Paragraphs[i] = *te.cloneParagraph(&para)
	}

	return newCell
}

// cloneTableCellProperties 深度复制表格单元格属性
func (te *TemplateEngine) cloneTableCellProperties(source *TableCellProperties) *TableCellProperties {
	if source == nil {
		return nil
	}

	props := &TableCellProperties{}

	// 复制单元格宽度
	if source.TableCellW != nil {
		props.TableCellW = &TableCellW{
			W:    source.TableCellW.W,
			Type: source.TableCellW.Type,
		}
	}

	// 复制单元格边距
	if source.TcMar != nil {
		props.TcMar = te.cloneTableCellMarginsCell(source.TcMar)
	}

	// 复制单元格边框
	if source.TcBorders != nil {
		props.TcBorders = te.cloneTableCellBorders(source.TcBorders)
	}

	// 复制单元格底纹
	if source.Shd != nil {
		props.Shd = &TableCellShading{
			Val:       source.Shd.Val,
			Color:     source.Shd.Color,
			Fill:      source.Shd.Fill,
			ThemeFill: source.Shd.ThemeFill,
		}
	}

	// 复制单元格垂直对齐
	if source.VAlign != nil {
		props.VAlign = &VAlign{
			Val: source.VAlign.Val,
		}
	}

	// 复制网格跨度
	if source.GridSpan != nil {
		props.GridSpan = &GridSpan{
			Val: source.GridSpan.Val,
		}
	}

	// 复制垂直合并
	if source.VMerge != nil {
		props.VMerge = &VMerge{
			Val: source.VMerge.Val,
		}
	}

	// 复制文字方向
	if source.TextDirection != nil {
		props.TextDirection = &TextDirection{
			Val: source.TextDirection.Val,
		}
	}

	// 复制禁止换行
	if source.NoWrap != nil {
		props.NoWrap = &NoWrap{
			Val: source.NoWrap.Val,
		}
	}

	// 复制隐藏标记
	if source.HideMark != nil {
		props.HideMark = &HideMark{
			Val: source.HideMark.Val,
		}
	}

	return props
}

// applyRenderedContentToDocument 将渲染内容应用到文档
func (te *TemplateEngine) applyRenderedContentToDocument(doc *Document, content string) error {
	// 如果内容为空，直接返回
	if strings.TrimSpace(content) == "" {
		return nil
	}

	// 将内容按行分割并创建段落
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		// 创建新段落，即使是空行也要创建（保持格式）
		para := &Paragraph{
			Properties: &ParagraphProperties{
				ParagraphStyle: &ParagraphStyle{Val: "Normal"},
			},
			Runs: []Run{},
		}

		// 如果行不为空，添加文本内容
		if strings.TrimSpace(line) != "" {
			run := Run{
				Text: Text{Content: line},
				Properties: &RunProperties{
					FontFamily: &FontFamily{
						ASCII:    "仿宋",
						HAnsi:    "仿宋",
						EastAsia: "仿宋",
					},
					FontSize: &FontSize{Val: "24"}, // 12pt = 24 half-points
				},
			}
			para.Runs = append(para.Runs, run)
		} else {
			// 空行也需要一个空的Run来保持段落结构
			run := Run{
				Text: Text{Content: ""},
				Properties: &RunProperties{
					FontFamily: &FontFamily{
						ASCII:    "仿宋",
						HAnsi:    "仿宋",
						EastAsia: "仿宋",
					},
					FontSize: &FontSize{Val: "24"},
				},
			}
			para.Runs = append(para.Runs, run)
		}

		// 将段落添加到文档
		doc.Body.Elements = append(doc.Body.Elements, para)
	}

	return nil
}

// RenderTemplateToDocument 渲染模板到新文档（新的主要方法）
func (te *TemplateEngine) RenderTemplateToDocument(templateName string, data *TemplateData) (*Document, error) {
	template, err := te.GetTemplate(templateName)
	if err != nil {
		return nil, WrapErrorWithContext("render_template_to_document", err, templateName)
	}

	// 如果有基础文档，克隆它并在其上进行变量替换
	if template.BaseDoc != nil {
		doc := te.cloneDocument(template.BaseDoc)

		// 在文档结构中直接进行变量替换
		err := te.replaceVariablesInDocument(doc, data)
		if err != nil {
			return nil, WrapErrorWithContext("render_template_to_document", err, templateName)
		}

		return doc, nil
	}

	// 如果没有基础文档，使用原有的方式
	return te.RenderToDocument(templateName, data)
}

// replaceVariablesInDocument 在文档结构中直接替换变量
func (te *TemplateEngine) replaceVariablesInDocument(doc *Document, data *TemplateData) error {
	// 首先处理文档级别的循环（跨段落）
	err := te.processDocumentLevelLoops(doc, data)
	if err != nil {
		return err
	}

	for _, element := range doc.Body.Elements {
		switch elem := element.(type) {
		case *Paragraph:
			// 处理段落中的变量替换
			err := te.replaceVariablesInParagraph(elem, data)
			if err != nil {
				return err
			}

		case *Table:
			// 处理表格中的变量替换和模板语法
			err := te.replaceVariablesInTable(elem, data)
			if err != nil {
				return err
			}
		}
	}

	// 处理图片占位符
	err = te.processImagePlaceholders(doc, data)
	if err != nil {
		return err
	}

	return nil
}

// processDocumentLevelLoops 处理文档级别的循环（跨段落）
func (te *TemplateEngine) processDocumentLevelLoops(doc *Document, data *TemplateData) error {
	elements := doc.Body.Elements
	newElements := make([]interface{}, 0)

	i := 0
	for i < len(elements) {
		element := elements[i]

		// 检查当前元素是否包含循环开始标记
		if para, ok := element.(*Paragraph); ok {
			// 获取段落的完整文本
			fullText := ""
			for _, run := range para.Runs {
				fullText += run.Text.Content
			}

			// 检查是否包含循环开始标记
			eachPattern := regexp.MustCompile(`\{\{#each\s+(\w+)\}\}`)
			matches := eachPattern.FindStringSubmatch(fullText)

			if len(matches) > 1 {
				listVarName := matches[1]

				// 找到循环结束位置
				loopEndIndex := -1
				templateElements := make([]interface{}, 0)

				// 收集循环模板元素（从当前位置到结束标记）
				for j := i; j < len(elements); j++ {
					templateElements = append(templateElements, elements[j])

					if nextPara, ok := elements[j].(*Paragraph); ok {
						nextText := ""
						for _, run := range nextPara.Runs {
							nextText += run.Text.Content
						}

						if strings.Contains(nextText, "{{/each}}") {
							loopEndIndex = j
							break
						}
					}
				}

				if loopEndIndex >= 0 {
					// 处理循环
					if listData, exists := data.Lists[listVarName]; exists {
						// 为每个数据项生成元素
						for _, item := range listData {
							if itemMap, ok := item.(map[string]interface{}); ok {
								// 复制模板元素并替换变量
								for _, templateElement := range templateElements {
									if templatePara, ok := templateElement.(*Paragraph); ok {
										newPara := te.cloneParagraph(templatePara)

										// 处理段落文本
										fullText := ""
										for _, run := range newPara.Runs {
											fullText += run.Text.Content
										}

										// 移除循环标记
										content := fullText
										content = regexp.MustCompile(`\{\{#each\s+\w+\}\}`).ReplaceAllString(content, "")
										content = regexp.MustCompile(`\{\{/each\}\}`).ReplaceAllString(content, "")

										// 替换变量
										for key, value := range itemMap {
											placeholder := fmt.Sprintf("{{%s}}", key)
											content = strings.ReplaceAll(content, placeholder, te.interfaceToString(value))
										}

										// 如果内容不为空，创建新段落
										if strings.TrimSpace(content) != "" {
											newPara.Runs = []Run{{
												Text: Text{Content: content},
												Properties: &RunProperties{
													FontFamily: &FontFamily{
														ASCII:    "仿宋",
														HAnsi:    "仿宋",
														EastAsia: "仿宋",
													},
													Bold: &Bold{},
												},
											}}
											newElements = append(newElements, newPara)
										}
									}
								}
							}
						}
					}

					// 跳过循环模板元素
					i = loopEndIndex + 1
					continue
				}
			}
		}

		// 不是循环元素，直接添加
		newElements = append(newElements, element)
		i++
	}

	// 更新文档元素
	doc.Body.Elements = newElements
	return nil
}

// replaceVariablesInParagraph 在段落中替换变量（改进版本，更好地保持样式）
func (te *TemplateEngine) replaceVariablesInParagraph(para *Paragraph, data *TemplateData) error {
	// 首先识别所有变量占位符的位置
	fullText := ""
	runInfos := make([]struct {
		startIndex int
		endIndex   int
		run        *Run
	}, 0)

	currentIndex := 0
	for i := range para.Runs {
		runText := para.Runs[i].Text.Content
		if runText != "" {
			runInfos = append(runInfos, struct {
				startIndex int
				endIndex   int
				run        *Run
			}{
				startIndex: currentIndex,
				endIndex:   currentIndex + len(runText),
				run:        &para.Runs[i],
			})
			fullText += runText
			currentIndex += len(runText)
		}
	}

	// 如果没有文本内容，直接返回
	if fullText == "" {
		return nil
	}

	// 先处理循环语句（包括非表格循环）
	processedText, hasLoopChanges := te.processNonTableLoops(fullText, data)
	if hasLoopChanges {
		// 重新构建段落
		para.Runs = []Run{{
			Text: Text{Content: processedText},
			Properties: &RunProperties{
				FontFamily: &FontFamily{
					ASCII:    "仿宋",
					HAnsi:    "仿宋",
					EastAsia: "仿宋",
				},
				Bold: &Bold{},
			},
		}}
		fullText = processedText
	}

	// 使用新的逐个变量替换方法
	newRuns, hasVarChanges := te.replaceVariablesSequentially(runInfos, fullText, data)

	// 如果有变化，更新段落的Run
	if hasVarChanges || hasLoopChanges {
		para.Runs = newRuns
	}

	return nil
}

// processNonTableLoops 处理非表格循环
func (te *TemplateEngine) processNonTableLoops(content string, data *TemplateData) (string, bool) {
	eachPattern := regexp.MustCompile(`(?s)\{\{#each\s+(\w+)\}\}(.*?)\{\{/each\}\}`)
	matches := eachPattern.FindAllStringSubmatchIndex(content, -1)

	if len(matches) == 0 {
		return content, false
	}

	var result strings.Builder
	lastEnd := 0
	hasChanges := false

	for _, match := range matches {
		// 找到变量名和块内容
		fullMatch := content[match[0]:match[1]]
		submatch := eachPattern.FindStringSubmatch(fullMatch)
		if len(submatch) >= 3 {
			listVar := submatch[1]
			blockContent := submatch[2]

			// 添加循环前的内容
			result.WriteString(content[lastEnd:match[0]])

			// 处理循环
			if listData, exists := data.Lists[listVar]; exists {
				for _, item := range listData {
					if itemMap, ok := item.(map[string]interface{}); ok {
						loopContent := blockContent
						for key, value := range itemMap {
							placeholder := fmt.Sprintf("{{%s}}", key)
							loopContent = strings.ReplaceAll(loopContent, placeholder, te.interfaceToString(value))
						}
						result.WriteString(loopContent)
					}
				}
			}

			lastEnd = match[1]
			hasChanges = true
		}
	}

	// 添加剩余内容
	if lastEnd < len(content) {
		result.WriteString(content[lastEnd:])
	}

	return result.String(), hasChanges
}

// replaceVariablesSequentially 逐个替换变量，保持样式
func (te *TemplateEngine) replaceVariablesSequentially(originalRunInfos []struct {
	startIndex int
	endIndex   int
	run        *Run
}, originalText string, data *TemplateData) ([]Run, bool) {

	// 找到所有变量位置
	varPattern := regexp.MustCompile(`\{\{(\w+)\}\}`)
	varMatches := varPattern.FindAllStringSubmatchIndex(originalText, -1)

	if len(varMatches) == 0 {
		// 没有变量，检查条件语句
		return te.processConditionals(originalRunInfos, originalText, data)
	}

	newRuns := make([]Run, 0)
	currentPos := 0
	hasChanges := false

	for _, varMatch := range varMatches {
		varStart := varMatch[0]
		varEnd := varMatch[1]
		varNameStart := varMatch[2]
		varNameEnd := varMatch[3]

		// 添加变量前的文本（保持原样式）
		if varStart > currentPos {
			beforeText := originalText[currentPos:varStart]
			beforeRuns := te.extractRunsForSegment(originalRunInfos, currentPos, varStart, beforeText)
			newRuns = append(newRuns, beforeRuns...)
		}

		// 处理变量替换
		varName := originalText[varNameStart:varNameEnd]
		if value, exists := data.Variables[varName]; exists {
			replacementText := te.interfaceToString(value)

			// 为变量选择合适的样式（使用覆盖变量位置的Run样式）
			varRun := te.findRunForPosition(originalRunInfos, varStart)
			if varRun != nil {
				newRun := te.cloneRun(varRun)
				newRun.Text.Content = replacementText
				newRuns = append(newRuns, newRun)
				hasChanges = true
			}
		} else {
			// 变量不存在，保持原始占位符
			varText := originalText[varStart:varEnd]
			varRun := te.findRunForPosition(originalRunInfos, varStart)
			if varRun != nil {
				newRun := te.cloneRun(varRun)
				newRun.Text.Content = varText
				newRuns = append(newRuns, newRun)
			}
		}

		currentPos = varEnd
	}

	// 添加最后剩余的文本
	if currentPos < len(originalText) {
		afterText := originalText[currentPos:]
		afterRuns := te.extractRunsForSegment(originalRunInfos, currentPos, len(originalText), afterText)
		newRuns = append(newRuns, afterRuns...)
	}

	// 如果没有找到任何变量但文本发生了变化，处理条件语句
	if !hasChanges {
		return te.processConditionals(originalRunInfos, originalText, data)
	}

	// 对结果处理条件语句（但要保持每个Run的独立性）
	if hasChanges {
		finalRuns := te.processConditionalsPreservingRuns(newRuns, data)
		return finalRuns, true
	}

	return newRuns, hasChanges
}

// processConditionalsPreservingRuns 处理条件语句但保持Run的独立性
func (te *TemplateEngine) processConditionalsPreservingRuns(runs []Run, data *TemplateData) []Run {
	finalRuns := make([]Run, 0)

	for _, run := range runs {
		originalContent := run.Text.Content
		processedContent := te.renderConditionals(originalContent, data.Conditions)

		// 如果内容发生变化，更新这个Run
		if processedContent != originalContent {
			newRun := run // 复制Run结构
			newRun.Text.Content = processedContent
			finalRuns = append(finalRuns, newRun)
		} else {
			// 内容没有变化，保持原样
			finalRuns = append(finalRuns, run)
		}
	}

	return finalRuns
}

// processConditionals 处理条件语句
func (te *TemplateEngine) processConditionals(originalRunInfos []struct {
	startIndex int
	endIndex   int
	run        *Run
}, originalText string, data *TemplateData) ([]Run, bool) {

	processedText := te.renderConditionals(originalText, data.Conditions)

	if processedText == originalText {
		// 没有变化，返回原始Runs
		newRuns := make([]Run, len(originalRunInfos))
		for i, runInfo := range originalRunInfos {
			newRuns[i] = te.cloneRun(runInfo.run)
		}
		return newRuns, false
	}

	// 有条件语句被处理，简化处理
	if len(originalRunInfos) == 1 {
		newRun := te.cloneRun(originalRunInfos[0].run)
		newRun.Text.Content = processedText
		return []Run{newRun}, true
	}

	// 多个Run的情况，使用第一个Run的样式
	newRun := te.cloneRun(originalRunInfos[0].run)
	newRun.Text.Content = processedText
	return []Run{newRun}, true
}

// extractRunsForSegment 为文本片段提取相应的Run（改进版本）
func (te *TemplateEngine) extractRunsForSegment(originalRunInfos []struct {
	startIndex int
	endIndex   int
	run        *Run
}, segmentStart, segmentEnd int, segmentText string) []Run {
	runs := make([]Run, 0)

	for _, runInfo := range originalRunInfos {
		// 检查Run是否与文本段有重叠
		if runInfo.endIndex > segmentStart && runInfo.startIndex < segmentEnd {
			overlapStart := max(runInfo.startIndex, segmentStart)
			overlapEnd := min(runInfo.endIndex, segmentEnd)

			if overlapEnd > overlapStart {
				newRun := te.cloneRun(runInfo.run)
				// 计算在分段文本中的相对位置
				relativeStart := overlapStart - segmentStart
				relativeEnd := overlapEnd - segmentStart

				// 确保索引在有效范围内
				if relativeStart >= 0 && relativeEnd <= len(segmentText) && relativeStart < relativeEnd {
					newRun.Text.Content = segmentText[relativeStart:relativeEnd]
					if newRun.Text.Content != "" {
						runs = append(runs, newRun)
					}
				}
			}
		}
	}

	return runs
}

// findRunForPosition 找到覆盖指定位置的Run
func (te *TemplateEngine) findRunForPosition(originalRunInfos []struct {
	startIndex int
	endIndex   int
	run        *Run
}, position int) *Run {
	for _, runInfo := range originalRunInfos {
		if position >= runInfo.startIndex && position < runInfo.endIndex {
			return runInfo.run
		}
	}
	// 如果没找到，返回第一个Run
	if len(originalRunInfos) > 0 {
		return originalRunInfos[0].run
	}
	return nil
}

// max 返回两个整数中的较大值
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// replaceVariablesInTable 在表格中替换变量和处理表格模板
func (te *TemplateEngine) replaceVariablesInTable(table *Table, data *TemplateData) error {
	// 检查是否有表格循环模板
	if len(table.Rows) > 0 && te.isTableTemplate(table) {
		return te.renderTableTemplate(table, data)
	}

	// 普通表格变量替换
	for i := range table.Rows {
		for j := range table.Rows[i].Cells {
			for k := range table.Rows[i].Cells[j].Paragraphs {
				err := te.replaceVariablesInParagraph(&table.Rows[i].Cells[j].Paragraphs[k], data)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// isTableTemplate 检查表格是否包含模板语法（支持跨Run检测）
func (te *TemplateEngine) isTableTemplate(table *Table) bool {
	if len(table.Rows) == 0 {
		return false
	}

	// 检查所有行是否包含循环语法，支持跨Run检测
	for _, row := range table.Rows {
		for _, cell := range row.Cells {
			for _, para := range cell.Paragraphs {
				// 使用新的跨Run检测方法
				if te.containsTemplateLoopInRuns(para.Runs) {
					return true
				}
			}
		}
	}

	return false
}

// containsTemplateLoop 检查文本是否包含循环模板语法（支持跨Run检测）
func (te *TemplateEngine) containsTemplateLoop(text string) bool {
	eachPattern := regexp.MustCompile(`\{\{#each\s+\w+\}\}`)
	return eachPattern.MatchString(text)
}

// containsTemplateLoopInRuns 检查Run列表中是否包含循环模板语法（跨Run检测）
func (te *TemplateEngine) containsTemplateLoopInRuns(runs []Run) bool {
	// 合并所有Run的文本
	fullText := ""
	for _, run := range runs {
		fullText += run.Text.Content
	}

	return te.containsTemplateLoop(fullText)
}

// renderTableTemplate 渲染表格模板
func (te *TemplateEngine) renderTableTemplate(table *Table, data *TemplateData) error {
	if len(table.Rows) == 0 {
		return nil
	}

	// 找到模板行（包含循环语法的行）
	templateRowIndex := -1
	var listVarName string

	for i, row := range table.Rows {
		found := false
		// 检查整行的所有单元格，合并文本来解决跨Run的变量问题
		for _, cell := range row.Cells {
			for _, para := range cell.Paragraphs {
				// 合并所有Run的文本
				fullText := ""
				for _, run := range para.Runs {
					fullText += run.Text.Content
				}

				// 检查合并后的文本中是否包含循环语法
				eachPattern := regexp.MustCompile(`\{\{#each\s+(\w+)\}\}`)
				matches := eachPattern.FindStringSubmatch(fullText)
				if len(matches) > 1 {
					templateRowIndex = i
					listVarName = matches[1]
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if found {
			break
		}
	}

	if templateRowIndex < 0 || listVarName == "" {
		return nil
	}

	// 获取列表数据
	listData, exists := data.Lists[listVarName]
	if !exists || len(listData) == 0 {
		// 删除模板行
		table.Rows = append(table.Rows[:templateRowIndex], table.Rows[templateRowIndex+1:]...)
		return nil
	}

	// 保存模板行
	templateRow := table.Rows[templateRowIndex]
	newRows := make([]TableRow, 0)

	// 保留模板行之前的行（深度克隆以保持样式）
	for _, row := range table.Rows[:templateRowIndex] {
		clonedRow := te.cloneTableRow(&row)
		newRows = append(newRows, *clonedRow)
	}

	// 为每个数据项生成新行
	for _, item := range listData {
		newRow := te.cloneTableRow(&templateRow)

		// 在新行中替换变量
		if itemMap, ok := item.(map[string]interface{}); ok {
			for i := range newRow.Cells {
				for j := range newRow.Cells[i].Paragraphs {
					// 合并所有Run的文本
					fullText := ""
					originalRuns := newRow.Cells[i].Paragraphs[j].Runs
					for _, run := range originalRuns {
						fullText += run.Text.Content
					}

					// 移除模板语法标记
					content := fullText
					content = regexp.MustCompile(`\{\{#each\s+\w+\}\}`).ReplaceAllString(content, "")
					content = regexp.MustCompile(`\{\{/each\}\}`).ReplaceAllString(content, "")

					// 替换变量
					for key, value := range itemMap {
						placeholder := fmt.Sprintf("{{%s}}", key)
						content = strings.ReplaceAll(content, placeholder, te.interfaceToString(value))
					}

					// 处理条件语句
					content = te.renderLoopConditionals(content, itemMap)

					// 重建Run结构，更好地保持样式继承
					if len(originalRuns) > 0 {
						// 寻找第一个有实际内容或样式的Run作为样式模板
						var templateRun *Run
						for k := range originalRuns {
							if originalRuns[k].Properties != nil || originalRuns[k].Text.Content != "" {
								templateRun = &originalRuns[k]
								break
							}
						}

						if templateRun != nil {
							newRun := te.cloneRun(templateRun)
							newRun.Text.Content = content
							newRow.Cells[i].Paragraphs[j].Runs = []Run{newRun}
						} else {
							// 使用第一个Run但确保基本样式
							newRun := te.cloneRun(&originalRuns[0])
							newRun.Text.Content = content
							// 确保基本的字体设置
							if newRun.Properties == nil {
								newRun.Properties = &RunProperties{}
							}
							if newRun.Properties.FontFamily == nil {
								newRun.Properties.FontFamily = &FontFamily{
									ASCII:    "仿宋",
									HAnsi:    "仿宋",
									EastAsia: "仿宋",
								}
							}
							newRow.Cells[i].Paragraphs[j].Runs = []Run{newRun}
						}
					} else {
						// 如果没有原始Run，创建新的但尝试继承段落样式
						newRun := Run{
							Text: Text{Content: content},
							Properties: &RunProperties{
								FontFamily: &FontFamily{
									ASCII:    "仿宋",
									HAnsi:    "仿宋",
									EastAsia: "仿宋",
								},
								Bold: &Bold{},
							},
						}

						// 如果段落有默认的Run属性，尝试继承
						if len(templateRow.Cells) > i && len(templateRow.Cells[i].Paragraphs) > j {
							templatePara := &templateRow.Cells[i].Paragraphs[j]
							if len(templatePara.Runs) > 0 && templatePara.Runs[0].Properties != nil {
								newRun.Properties = te.cloneRunProperties(templatePara.Runs[0].Properties)
							}
						}

						newRow.Cells[i].Paragraphs[j].Runs = []Run{newRun}
					}
				}
			}
		}

		newRows = append(newRows, *newRow)
	}

	// 保留模板行之后的行（深度克隆以保持样式）
	for _, row := range table.Rows[templateRowIndex+1:] {
		clonedRow := te.cloneTableRow(&row)
		newRows = append(newRows, *clonedRow)
	}

	// 更新表格行
	table.Rows = newRows

	return nil
}

// NewTemplateData 创建新的模板数据
func NewTemplateData() *TemplateData {
	return &TemplateData{
		Variables:  make(map[string]interface{}),
		Lists:      make(map[string][]interface{}),
		Conditions: make(map[string]bool),
		Images:     make(map[string]*TemplateImageData),
	}
}

// SetVariable 设置变量
func (td *TemplateData) SetVariable(name string, value interface{}) {
	td.Variables[name] = value
}

// SetList 设置列表
func (td *TemplateData) SetList(name string, list []interface{}) {
	td.Lists[name] = list
}

// SetCondition 设置条件
func (td *TemplateData) SetCondition(name string, value bool) {
	td.Conditions[name] = value
}

// SetVariables 批量设置变量
func (td *TemplateData) SetVariables(variables map[string]interface{}) {
	for name, value := range variables {
		td.Variables[name] = value
	}
}

// GetVariable 获取变量
func (td *TemplateData) GetVariable(name string) (interface{}, bool) {
	value, exists := td.Variables[name]
	return value, exists
}

// GetList 获取列表
func (td *TemplateData) GetList(name string) ([]interface{}, bool) {
	list, exists := td.Lists[name]
	return list, exists
}

// GetCondition 获取条件
func (td *TemplateData) GetCondition(name string) (bool, bool) {
	value, exists := td.Conditions[name]
	return value, exists
}

// GetImage 获取图片数据
func (td *TemplateData) GetImage(name string) (*TemplateImageData, bool) {
	value, exists := td.Images[name]
	return value, exists
}

// Merge 合并模板数据
func (td *TemplateData) Merge(other *TemplateData) {
	// 合并变量
	for key, value := range other.Variables {
		td.Variables[key] = value
	}

	// 合并列表
	for key, value := range other.Lists {
		td.Lists[key] = value
	}

	// 合并条件
	for key, value := range other.Conditions {
		td.Conditions[key] = value
	}

	// 合并图片
	for key, value := range other.Images {
		td.Images[key] = value
	}
}

// Clear 清空模板数据
func (td *TemplateData) Clear() {
	td.Variables = make(map[string]interface{})
	td.Lists = make(map[string][]interface{})
	td.Conditions = make(map[string]bool)
	td.Images = make(map[string]*TemplateImageData)
}

// FromStruct 从结构体生成模板数据
func (td *TemplateData) FromStruct(data interface{}) error {
	value := reflect.ValueOf(data)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	if value.Kind() != reflect.Struct {
		return NewValidationError("data_type", "struct", "expected struct type")
	}

	typ := value.Type()
	for i := 0; i < value.NumField(); i++ {
		field := typ.Field(i)
		fieldValue := value.Field(i)

		// 跳过不可导出的字段
		if !fieldValue.CanInterface() {
			continue
		}

		fieldName := strings.ToLower(field.Name)
		td.Variables[fieldName] = fieldValue.Interface()
	}

	return nil
}

// SetImage 设置图片数据（通过文件路径）
func (td *TemplateData) SetImage(name, filePath string, config *ImageConfig) {
	imageData := &TemplateImageData{
		FilePath: filePath,
		Config:   config,
	}
	td.Images[name] = imageData
}

// SetImageFromData 设置图片数据（通过二进制数据）
func (td *TemplateData) SetImageFromData(name string, data []byte, config *ImageConfig) {
	imageData := &TemplateImageData{
		Data:   data,
		Config: config,
	}
	td.Images[name] = imageData
}

// SetImageWithDetails 设置图片数据（完整配置）
func (td *TemplateData) SetImageWithDetails(name, filePath string, data []byte, config *ImageConfig, altText, title string) {
	imageData := &TemplateImageData{
		FilePath: filePath,
		Data:     data,
		Config:   config,
		AltText:  altText,
		Title:    title,
	}
	td.Images[name] = imageData
}

// renderImages 渲染图片占位符
func (te *TemplateEngine) renderImages(content string, images map[string]*TemplateImageData) string {
	// 图片占位符模式: {{#image imageName}}
	imagePattern := regexp.MustCompile(`\{\{#image\s+(\w+)\}\}`)

	return imagePattern.ReplaceAllStringFunc(content, func(match string) string {
		matches := imagePattern.FindStringSubmatch(match)
		if len(matches) >= 2 {
			imageName := matches[1]

			// 查找图片数据
			if _, exists := images[imageName]; exists {
				// 在传统的字符串模板中，我们返回图片占位符标记
				// 实际的图片处理将在RenderTemplateToDocument方法中完成
				return fmt.Sprintf("[IMAGE:%s]", imageName)
			}
		}
		// 如果找不到图片数据，保持原样或返回错误信息
		return fmt.Sprintf("[IMAGE_NOT_FOUND:%s]", matches[1])
	})
}

// processImagePlaceholders 处理文档中的图片占位符
func (te *TemplateEngine) processImagePlaceholders(doc *Document, data *TemplateData) error {
	// 遍历文档元素，查找并替换图片占位符
	for i, element := range doc.Body.Elements {
		switch elem := element.(type) {
		case *Paragraph:
			// 检查段落是否包含图片占位符
			newElements, err := te.processImagePlaceholdersInParagraph(elem, data, doc)
			if err != nil {
				return err
			}

			// 如果有图片替换，更新文档元素
			if len(newElements) > 1 || (len(newElements) == 1 && newElements[0] != elem) {
				// 移除原段落，插入新元素（可能包含图片段落）
				doc.Body.Elements = append(doc.Body.Elements[:i], append(newElements, doc.Body.Elements[i+1:]...)...)
			}
		}
	}
	return nil
}

// processImagePlaceholdersInParagraph 处理段落中的图片占位符
func (te *TemplateEngine) processImagePlaceholdersInParagraph(para *Paragraph, data *TemplateData, doc *Document) ([]interface{}, error) {
	// 获取段落的完整文本
	fullText := ""
	for _, run := range para.Runs {
		fullText += run.Text.Content
	}

	// 检查是否包含图片占位符（支持两种格式）
	// 1. 原始模板格式：{{#image imageName}}
	// 2. 渲染后格式：[IMAGE:imageName]
	originalImagePattern := regexp.MustCompile(`\{\{#image\s+(\w+)\}\}`)
	renderedImagePattern := regexp.MustCompile(`\[IMAGE:(\w+)\]`)

	originalMatches := originalImagePattern.FindAllStringSubmatch(fullText, -1)
	renderedMatches := renderedImagePattern.FindAllStringSubmatch(fullText, -1)

	// 合并两种格式的匹配结果
	allMatches := make([][2]string, 0)
	for _, match := range originalMatches {
		allMatches = append(allMatches, [2]string{match[0], match[1]})
	}
	for _, match := range renderedMatches {
		allMatches = append(allMatches, [2]string{match[0], match[1]})
	}

	if len(allMatches) == 0 {
		// 没有图片占位符，返回原段落
		return []interface{}{para}, nil
	}

	result := make([]interface{}, 0)
	lastEnd := 0

	// 处理每个图片占位符
	for _, match := range allMatches {
		imageName := match[1]
		matchStart := strings.Index(fullText[lastEnd:], match[0]) + lastEnd
		matchEnd := matchStart + len(match[0])

		// 添加图片占位符前的文本（如果有）
		if matchStart > lastEnd {
			beforeText := fullText[lastEnd:matchStart]
			if strings.TrimSpace(beforeText) != "" {
				beforePara := te.createTextParagraph(beforeText, para)
				result = append(result, beforePara)
			}
		}

		// 创建图片段落
		if imageData, exists := data.Images[imageName]; exists {
			imagePara, err := te.createImageParagraph(imageData, doc)
			if err != nil {
				return nil, fmt.Errorf("创建图片段落失败: %v", err)
			}
			result = append(result, imagePara)
		} else {
			// 图片数据不存在，创建错误文本段落
			errorPara := te.createTextParagraph(fmt.Sprintf("[图片未找到: %s]", imageName), para)
			result = append(result, errorPara)
		}

		lastEnd = matchEnd
	}

	// 添加最后剩余的文本（如果有）
	if lastEnd < len(fullText) {
		afterText := fullText[lastEnd:]
		if strings.TrimSpace(afterText) != "" {
			afterPara := te.createTextParagraph(afterText, para)
			result = append(result, afterPara)
		}
	}

	// 如果没有任何内容，返回空段落
	if len(result) == 0 {
		emptyPara := te.createTextParagraph("", para)
		result = append(result, emptyPara)
	}

	return result, nil
}

// createTextParagraph 创建文本段落（保持原段落样式）
func (te *TemplateEngine) createTextParagraph(text string, originalPara *Paragraph) *Paragraph {
	newPara := te.cloneParagraph(originalPara)

	// 设置文本内容，保持原有样式
	if len(newPara.Runs) > 0 {
		newPara.Runs[0].Text.Content = text
		newPara.Runs = newPara.Runs[:1] // 只保留第一个run
	} else {
		// 如果原段落没有runs，创建一个默认的
		newPara.Runs = []Run{{
			Text: Text{Content: text},
		}}
	}

	return newPara
}

// createImageParagraph 创建图片段落
func (te *TemplateEngine) createImageParagraph(imageData *TemplateImageData, doc *Document) (*Paragraph, error) {
	// 创建图片配置
	config := imageData.Config
	if config == nil {
		config = &ImageConfig{
			Position:  ImagePositionInline,
			Alignment: AlignCenter,
		}
	}

	// 添加图片到文档
	var imageInfo *ImageInfo
	var err error

	if len(imageData.Data) > 0 {
		// 使用二进制数据
		var format ImageFormat
		format, err = detectImageFormat(imageData.Data)
		if err != nil {
			return nil, fmt.Errorf("检测图片格式失败: %v", err)
		}

		var width, height int
		width, height, err = getImageDimensions(imageData.Data, format)
		if err != nil {
			return nil, fmt.Errorf("获取图片尺寸失败: %v", err)
		}

		// 使用唯一的文件名，包含图片ID计数器
		fileName := fmt.Sprintf("image_%d.%s", doc.nextImageID, string(format))
		// 使用不创建段落元素的方法，由模板引擎自行管理段落
		imageInfo, err = doc.AddImageFromDataWithoutElement(imageData.Data, fileName, format, width, height, config)
	} else if imageData.FilePath != "" {
		// 使用文件路径，但需要先读取数据，然后使用AddImageFromDataWithoutElement
		data, readErr := os.ReadFile(imageData.FilePath)
		if readErr != nil {
			return nil, fmt.Errorf("读取图片文件失败: %v", readErr)
		}

		var format ImageFormat
		format, err = detectImageFormat(data)
		if err != nil {
			return nil, fmt.Errorf("检测图片格式失败: %v", err)
		}

		var width, height int
		width, height, err = getImageDimensions(data, format)
		if err != nil {
			return nil, fmt.Errorf("获取图片尺寸失败: %v", err)
		}

		fileName := filepath.Base(imageData.FilePath)
		imageInfo, err = doc.AddImageFromDataWithoutElement(data, fileName, format, width, height, config)
	} else {
		return nil, fmt.Errorf("图片数据和文件路径都为空")
	}

	if err != nil {
		return nil, fmt.Errorf("添加图片失败: %v", err)
	}

	// 设置图片描述和标题
	if imageData.AltText != "" {
		doc.SetImageAltText(imageInfo, imageData.AltText)
	}
	if imageData.Title != "" {
		doc.SetImageTitle(imageInfo, imageData.Title)
	}

	// 创建包含图片的段落
	imagePara := doc.createImageParagraph(imageInfo)
	return imagePara, nil
}
