// Package main 演示WordZero高级功能
package main

import (
	"fmt"
	"log"

	"github.com/ZeroHawkeye/wordZero/pkg/document"
)

func main() {
	fmt.Println("正在创建高级功能演示文档...")

	// 创建新文档
	doc := document.New()

	// 1. 设置文档标题和副标题
	title := doc.AddFormattedParagraph("高级功能演示文档", &document.TextFormat{
		Bold:       true,
		FontSize:   18,
		FontColor:  "2F5496",
		FontFamily: "微软雅黑",
	})
	title.SetAlignment(document.AlignCenter)
	title.SetSpacing(&document.SpacingConfig{
		AfterPara: 12,
	})

	subtitle := doc.AddFormattedParagraph("包含目录、表格、页眉页脚和各种格式", &document.TextFormat{
		Italic:     true,
		FontSize:   12,
		FontColor:  "7030A0",
		FontFamily: "微软雅黑",
	})
	subtitle.SetAlignment(document.AlignCenter)
	subtitle.SetSpacing(&document.SpacingConfig{
		AfterPara: 18,
	})

	// 2. 添加多级标题以生成层级目录
	fmt.Println("添加多级标题...")

	// 一级标题
	h1_1 := doc.AddHeadingParagraph("第一章 文档基础功能", 1)
	h1_1.SetSpacing(&document.SpacingConfig{
		BeforePara: 18,
		AfterPara:  12,
	})

	// 二级标题
	h2_1 := doc.AddHeadingParagraph("1.1 文本格式化", 2)
	h2_1.SetSpacing(&document.SpacingConfig{
		BeforePara: 12,
		AfterPara:  6,
	})

	// 三级标题
	h3_1 := doc.AddHeadingParagraph("1.1.1 字体设置", 3)
	h3_1.SetSpacing(&document.SpacingConfig{
		BeforePara: 6,
		AfterPara:  6,
	})

	// 添加一些内容段落
	doc.AddParagraph("这里演示了字体设置的功能，包括字体大小、颜色、粗体、斜体等各种格式选项。")

	h3_2 := doc.AddHeadingParagraph("1.1.2 段落格式", 3)
	h3_2.SetSpacing(&document.SpacingConfig{
		BeforePara: 6,
		AfterPara:  6,
	})

	doc.AddParagraph("段落格式包括对齐方式、行间距、段间距、缩进等设置。")

	// 二级标题
	h2_2 := doc.AddHeadingParagraph("1.2 样式管理", 2)
	h2_2.SetSpacing(&document.SpacingConfig{
		BeforePara: 12,
		AfterPara:  6,
	})

	doc.AddParagraph("样式管理系统提供了预定义样式和自定义样式功能。")

	// 一级标题
	h1_2 := doc.AddHeadingParagraph("第二章 表格功能", 1)
	h1_2.SetSpacing(&document.SpacingConfig{
		BeforePara: 18,
		AfterPara:  12,
	})

	// 二级标题
	h2_3 := doc.AddHeadingParagraph("2.1 表格创建", 2)
	h2_3.SetSpacing(&document.SpacingConfig{
		BeforePara: 12,
		AfterPara:  6,
	})

	// 三级标题
	h3_3 := doc.AddHeadingParagraph("2.1.1 基础表格", 3)
	h3_3.SetSpacing(&document.SpacingConfig{
		BeforePara: 6,
		AfterPara:  6,
	})

	doc.AddParagraph("演示基础表格创建功能。")

	h3_4 := doc.AddHeadingParagraph("2.1.2 高级表格", 3)
	h3_4.SetSpacing(&document.SpacingConfig{
		BeforePara: 6,
		AfterPara:  6,
	})

	doc.AddParagraph("演示高级表格功能，包括合并单元格、样式设置等。")

	// 二级标题
	h2_4 := doc.AddHeadingParagraph("2.2 表格样式", 2)
	h2_4.SetSpacing(&document.SpacingConfig{
		BeforePara: 12,
		AfterPara:  6,
	})

	doc.AddParagraph("表格样式设置和格式化选项。")

	// 一级标题
	h1_3 := doc.AddHeadingParagraph("第三章 高级功能", 1)
	h1_3.SetSpacing(&document.SpacingConfig{
		BeforePara: 18,
		AfterPara:  12,
	})

	// 二级标题
	h2_5 := doc.AddHeadingParagraph("3.1 页面设置", 2)
	h2_5.SetSpacing(&document.SpacingConfig{
		BeforePara: 12,
		AfterPara:  6,
	})

	doc.AddParagraph("页面大小、边距、方向等设置功能。")

	h2_6 := doc.AddHeadingParagraph("3.2 目录生成", 2)
	h2_6.SetSpacing(&document.SpacingConfig{
		BeforePara: 12,
		AfterPara:  6,
	})

	doc.AddParagraph("自动生成目录功能，支持多级标题和正确的缩进显示。")

	// 3. 在文档开头生成目录
	fmt.Println("生成自动目录...")

	config := &document.TOCConfig{
		Title:        "目录",
		MaxLevel:     3,
		ShowPageNum:  true,
		RightAlign:   true,
		UseHyperlink: true,
		DotLeader:    true,
	}

	err := doc.AutoGenerateTOC(config)
	if err != nil {
		log.Printf("生成目录失败: %v", err)
	} else {
		fmt.Println("目录生成成功！")
	}

	// 4. 设置页面属性 - 暂时跳过，因为API可能尚未实现
	fmt.Println("设置页面属性...")
	// err = doc.SetPageSize(&document.PageSize{
	// 	Width:       210, // A4宽度
	// 	Height:      297, // A4高度
	// 	Orientation: document.OrientationPortrait,
	// })
	// if err != nil {
	// 	log.Printf("设置页面大小失败: %v", err)
	// }

	// err = doc.SetPageMargins(25, 25, 30, 20)  // 上下左右边距
	// if err != nil {
	// 	log.Printf("设置页面边距失败: %v", err)
	// }

	// 5. 添加页眉页脚
	fmt.Println("添加页眉页脚...")
	err = doc.AddHeader(document.HeaderFooterTypeDefault, "高级功能演示文档")
	if err != nil {
		log.Printf("添加页眉失败: %v", err)
	}

	err = doc.AddFooterWithPageNumber(document.HeaderFooterTypeDefault, "", true)
	if err != nil {
		log.Printf("添加页脚失败: %v", err)
	}

	// 6. 创建演示表格
	fmt.Println("创建演示表格...")

	// 在文档末尾添加表格说明
	doc.AddParagraph("") // 空行
	tableTitle := doc.AddFormattedParagraph("演示表格", &document.TextFormat{
		Bold:     true,
		FontSize: 14,
	})
	tableTitle.SetAlignment(document.AlignCenter)

	// 创建3x4的表格
	table, _ := doc.AddTable(&document.TableConfig{
		Rows:  4,
		Cols:  3,
		Width: 9000,
		Data: [][]string{
			{"项目", "描述", "状态"},
			{"文本格式化", "支持字体、大小、颜色等设置", "✅ 完成"},
			{"段落格式", "支持对齐、间距、缩进等", "✅ 完成"},
			{"目录生成", "自动生成多级目录", "🔧 已修复缩进"},
		},
	})

	// 设置表格样式
	table.SetTableAlignment(document.TableAlignCenter)

	// 设置标题行格式
	for j := 0; j < 3; j++ {
		table.SetCellFormat(0, j, &document.CellFormat{
			TextFormat: &document.TextFormat{
				Bold:      true,
				FontColor: "FFFFFF",
			},
			BackgroundColor: "2F5496",
			HorizontalAlign: document.CellAlignCenter,
			VerticalAlign:   document.CellVAlignCenter,
		})
	}

	// 7. 添加脚注说明
	fmt.Println("添加脚注...")
	footnoteText := doc.AddParagraph("本文档演示了WordZero库的主要功能特性")
	footnoteText.AddFormattedText("¹", &document.TextFormat{
		FontSize: 8,
	})

	// 暂时跳过脚注功能，如果API不可用
	// err = doc.AddFootnote("脚注示例", "这是一个脚注示例，展示了脚注功能的使用。")
	// if err != nil {
	// 	log.Printf("添加脚注失败: %v", err)
	// }

	// 9. 保存文档
	filename := "examples/output/advanced_features_demo.docx"
	fmt.Printf("正在保存文档到: %s\n", filename)

	err = doc.Save(filename)
	if err != nil {
		log.Fatalf("保存文档失败: %v", err)
	}

	fmt.Println("✅ 高级功能演示文档创建完成！")
	fmt.Println("📊 文档统计信息:")

	// 获取标题统计
	headingCount := doc.GetHeadingCount()
	for level := 1; level <= 3; level++ {
		if count, exists := headingCount[level]; exists {
			fmt.Printf("   - %d级标题: %d个\n", level, count)
		}
	}

	// 列出所有标题
	fmt.Println("📋 标题列表:")
	headings := doc.ListHeadings()
	for _, heading := range headings {
		indent := ""
		for i := 1; i < heading.Level; i++ {
			indent += "  "
		}
		fmt.Printf("   %s%d. %s\n", indent, heading.Level, heading.Text)
	}

	fmt.Printf("\n🎉 文档已成功保存到: %s\n", filename)
	fmt.Println("💡 提示：打开Word文档，检查目录是否显示正确的层级缩进！")
}
