package benchmark

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/ZeroHawkeye/wordZero/pkg/document"
	"github.com/ZeroHawkeye/wordZero/pkg/style"
)

// 统一的测试配置，与JavaScript和Python保持一致
var testIterations = map[string]int{
	"basic":      50, // 基础文档创建
	"complex":    30, // 复杂格式化
	"table":      20, // 表格操作
	"largeTable": 10, // 大表格处理
	"largeDoc":   5,  // 大型文档
	"memory":     10, // 内存使用测试
}

// BenchmarkCreateBasicDocument 基础文档创建性能测试
func BenchmarkCreateBasicDocument(b *testing.B) {
	outputDir := "../results/golang"
	os.MkdirAll(outputDir, 0755)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doc := document.New()
		doc.AddParagraph("这是一个基础性能测试文档")
		doc.AddParagraph("测试内容包括基本的文本添加功能")

		filename := filepath.Join(outputDir, fmt.Sprintf("basic_doc_%d.docx", i))
		doc.Save(filename)
	}
}

// BenchmarkComplexFormatting 复杂格式化性能测试
func BenchmarkComplexFormatting(b *testing.B) {
	outputDir := "../results/golang"
	os.MkdirAll(outputDir, 0755)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doc := document.New()

		// 添加标题
		doc.AddParagraph("性能测试报告").SetStyle(style.StyleHeading1)
		doc.AddParagraph("测试概述").SetStyle(style.StyleHeading2)

		// 添加格式化文本
		para := doc.AddParagraph("")
		para.AddFormattedText("粗体文本", &document.TextFormat{Bold: true})
		para.AddFormattedText(" ", &document.TextFormat{})
		para.AddFormattedText("斜体文本", &document.TextFormat{Italic: true})
		para.AddFormattedText(" ", &document.TextFormat{})
		para.AddFormattedText("彩色文本", &document.TextFormat{FontColor: "FF0000"})

		// 添加不同样式的段落
		for j := 0; j < 10; j++ {
			para2 := doc.AddParagraph(fmt.Sprintf("这是第%d个段落，包含复杂格式化", j+1))
			para2.SetAlignment(document.AlignCenter)
		}

		filename := filepath.Join(outputDir, fmt.Sprintf("complex_formatting_%d.docx", i))
		doc.Save(filename)
	}
}

// BenchmarkTableOperations 表格操作性能测试
func BenchmarkTableOperations(b *testing.B) {
	outputDir := "../results/golang"
	os.MkdirAll(outputDir, 0755)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doc := document.New()
		doc.AddParagraph("表格性能测试").SetStyle(style.StyleHeading1)

		// 创建10行5列的表格
		tableConfig := &document.TableConfig{
			Rows:  10,
			Cols:  5,
			Width: 7200, // 5英寸 = 7200磅
		}
		table, _ := doc.AddTable(tableConfig)

		// 填充表格数据
		for row := 0; row < 10; row++ {
			for col := 0; col < 5; col++ {
				cellText := fmt.Sprintf("R%dC%d", row+1, col+1)
				table.SetCellText(row, col, cellText)
			}
		}

		filename := filepath.Join(outputDir, fmt.Sprintf("table_operations_%d.docx", i))
		doc.Save(filename)
	}
}

// BenchmarkLargeTable 大表格性能测试
func BenchmarkLargeTable(b *testing.B) {
	outputDir := "../results/golang"
	os.MkdirAll(outputDir, 0755)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doc := document.New()
		doc.AddParagraph("大表格性能测试").SetStyle(style.StyleHeading1)

		// 创建100行10列的大表格
		tableConfig := &document.TableConfig{
			Rows:  100,
			Cols:  10,
			Width: 14400, // 10英寸 = 14400磅
		}
		table, _ := doc.AddTable(tableConfig)

		// 填充表格数据
		for row := 0; row < 100; row++ {
			for col := 0; col < 10; col++ {
				cellText := fmt.Sprintf("数据_%d_%d", row+1, col+1)
				table.SetCellText(row, col, cellText)
			}
		}

		filename := filepath.Join(outputDir, fmt.Sprintf("large_table_%d.docx", i))
		doc.Save(filename)
	}
}

// BenchmarkLargeDocument 大型文档性能测试
func BenchmarkLargeDocument(b *testing.B) {
	outputDir := "../results/golang"
	os.MkdirAll(outputDir, 0755)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		doc := document.New()
		doc.AddParagraph("大型文档性能测试").SetStyle(style.StyleHeading1)

		// 添加1000个段落
		for j := 0; j < 1000; j++ {
			if j%10 == 0 {
				// 每10个段落添加一个标题
				doc.AddParagraph(fmt.Sprintf("章节 %d", j/10+1)).SetStyle(style.StyleHeading2)
			}

			doc.AddParagraph(fmt.Sprintf("这是第%d个段落。Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.", j+1))
		}

		// 添加一个中等大小的表格
		tableConfig := &document.TableConfig{
			Rows:  20,
			Cols:  8,
			Width: 11520, // 8英寸 = 11520磅
		}
		table, _ := doc.AddTable(tableConfig)
		for row := 0; row < 20; row++ {
			for col := 0; col < 8; col++ {
				table.SetCellText(row, col, fmt.Sprintf("表格数据%d-%d", row+1, col+1))
			}
		}

		filename := filepath.Join(outputDir, fmt.Sprintf("large_document_%d.docx", i))
		doc.Save(filename)
	}
}

// BenchmarkMemoryUsage 内存使用测试
func BenchmarkMemoryUsage(b *testing.B) {
	outputDir := "../results/golang"
	os.MkdirAll(outputDir, 0755)

	var m1, m2 runtime.MemStats

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		runtime.GC()
		runtime.ReadMemStats(&m1)

		doc := document.New()

		// 创建复杂内容
		for j := 0; j < 100; j++ {
			doc.AddParagraph(fmt.Sprintf("段落%d: 这是一个测试段落，用于测试内存使用情况", j+1))
		}

		tableConfig := &document.TableConfig{
			Rows:  50,
			Cols:  6,
			Width: 8640, // 6英寸 = 8640磅
		}
		table, _ := doc.AddTable(tableConfig)
		for row := 0; row < 50; row++ {
			for col := 0; col < 6; col++ {
				table.SetCellText(row, col, fmt.Sprintf("单元格%d-%d", row+1, col+1))
			}
		}

		runtime.ReadMemStats(&m2)

		// 输出内存使用情况（可选）
		if i == 0 {
			b.Logf("内存使用: %d KB", (m2.Alloc-m1.Alloc)/1024)
		}

		filename := filepath.Join(outputDir, fmt.Sprintf("memory_test_%d.docx", i))
		doc.Save(filename)
	}
}

// === 新增：固定迭代次数的测试函数，与其他语言保持一致 ===

// TestFixedIterationsPerformance 固定迭代次数的性能测试，与JavaScript和Python保持一致
func TestFixedIterationsPerformance(t *testing.T) {
	outputDir := "../results/golang"
	os.MkdirAll(outputDir, 0755)

	tests := []struct {
		name       string
		iterations int
		testFunc   func(int) time.Duration
	}{
		{"基础文档创建", testIterations["basic"], testBasicDocumentCreationFixed},
		{"复杂格式化", testIterations["complex"], testComplexFormattingFixed},
		{"表格操作", testIterations["table"], testTableOperationsFixed},
		{"大表格处理", testIterations["largeTable"], testLargeTableProcessingFixed},
		{"大型文档", testIterations["largeDoc"], testLargeDocumentFixed},
		{"内存使用测试", testIterations["memory"], testMemoryUsageFixed},
	}

	results := make([]map[string]interface{}, 0, len(tests))

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Logf("开始测试: %s (迭代次数: %d)", tt.name, tt.iterations)

			times := make([]float64, 0, tt.iterations)

			for i := 0; i < tt.iterations; i++ {
				duration := tt.testFunc(i)
				times = append(times, float64(duration.Nanoseconds())/1e6) // 转换为毫秒

				if i%max(1, tt.iterations/10) == 0 {
					t.Logf("  进度: %d/%d", i+1, tt.iterations)
				}
			}

			// 计算统计数据
			var total float64
			minTime := times[0]
			maxTime := times[0]

			for _, time := range times {
				total += time
				if time < minTime {
					minTime = time
				}
				if time > maxTime {
					maxTime = time
				}
			}

			avgTime := total / float64(len(times))

			result := map[string]interface{}{
				"name":       tt.name,
				"avgTime":    fmt.Sprintf("%.2f", avgTime),
				"minTime":    fmt.Sprintf("%.2f", minTime),
				"maxTime":    fmt.Sprintf("%.2f", maxTime),
				"iterations": tt.iterations,
			}

			results = append(results, result)

			t.Logf("  平均耗时: %.2fms", avgTime)
			t.Logf("  最小耗时: %.2fms", minTime)
			t.Logf("  最大耗时: %.2fms", maxTime)
		})
	}

	// 生成性能报告（JSON格式，与其他语言保持一致）
	report := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"platform":  "Golang",
		"goVersion": runtime.Version(),
		"results":   results,
	}

	// 保存报告
	reportPath := filepath.Join(outputDir, "performance_report.json")
	file, err := os.Create(reportPath)
	if err != nil {
		t.Fatalf("无法创建报告文件: %v", err)
	}
	defer file.Close()

	fmt.Fprintf(file, "{\n")
	fmt.Fprintf(file, "  \"timestamp\": \"%s\",\n", report["timestamp"])
	fmt.Fprintf(file, "  \"platform\": \"%s\",\n", report["platform"])
	fmt.Fprintf(file, "  \"goVersion\": \"%s\",\n", report["goVersion"])
	fmt.Fprintf(file, "  \"results\": [\n")

	for i, result := range results {
		fmt.Fprintf(file, "    {\n")
		fmt.Fprintf(file, "      \"name\": \"%s\",\n", result["name"])
		fmt.Fprintf(file, "      \"avgTime\": \"%s\",\n", result["avgTime"])
		fmt.Fprintf(file, "      \"minTime\": \"%s\",\n", result["minTime"])
		fmt.Fprintf(file, "      \"maxTime\": \"%s\",\n", result["maxTime"])
		fmt.Fprintf(file, "      \"iterations\": %d\n", result["iterations"])
		if i < len(results)-1 {
			fmt.Fprintf(file, "    },\n")
		} else {
			fmt.Fprintf(file, "    }\n")
		}
	}

	fmt.Fprintf(file, "  ]\n")
	fmt.Fprintf(file, "}\n")

	t.Logf("\n=== Golang 性能测试报告 ===")
	t.Logf("详细报告已保存到: %s", reportPath)
}

// max 辅助函数
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// === 固定迭代次数的测试实现函数 ===

func testBasicDocumentCreationFixed(index int) time.Duration {
	start := time.Now()

	doc := document.New()
	doc.AddParagraph("这是一个基础性能测试文档")
	doc.AddParagraph("测试内容包括基本的文本添加功能")
	doc.Save(fmt.Sprintf("../results/golang/fixed_basic_doc_%d.docx", index))

	return time.Since(start)
}

func testComplexFormattingFixed(index int) time.Duration {
	start := time.Now()

	doc := document.New()
	doc.AddParagraph("性能测试报告").SetStyle(style.StyleHeading1)
	doc.AddParagraph("测试概述").SetStyle(style.StyleHeading2)

	para := doc.AddParagraph("")
	para.AddFormattedText("粗体文本", &document.TextFormat{Bold: true})
	para.AddFormattedText(" ", &document.TextFormat{})
	para.AddFormattedText("斜体文本", &document.TextFormat{Italic: true})
	para.AddFormattedText(" ", &document.TextFormat{})
	para.AddFormattedText("彩色文本", &document.TextFormat{FontColor: "FF0000"})

	// 添加不同样式的段落
	for j := 0; j < 10; j++ {
		para2 := doc.AddParagraph(fmt.Sprintf("这是第%d个段落，包含复杂格式化", j+1))
		para2.SetAlignment(document.AlignCenter)
	}

	doc.Save(fmt.Sprintf("../results/golang/fixed_complex_formatting_%d.docx", index))

	return time.Since(start)
}

func testTableOperationsFixed(index int) time.Duration {
	start := time.Now()

	doc := document.New()
	doc.AddParagraph("表格性能测试").SetStyle(style.StyleHeading1)

	tableConfig := &document.TableConfig{
		Rows:  10,
		Cols:  5,
		Width: 7200, // 5英寸
	}
	table, _ := doc.AddTable(tableConfig)

	for row := 0; row < 10; row++ {
		for col := 0; col < 5; col++ {
			table.SetCellText(row, col, fmt.Sprintf("R%dC%d", row+1, col+1))
		}
	}

	doc.Save(fmt.Sprintf("../results/golang/fixed_table_operations_%d.docx", index))

	return time.Since(start)
}

func testLargeTableProcessingFixed(index int) time.Duration {
	start := time.Now()

	doc := document.New()
	doc.AddParagraph("大表格性能测试").SetStyle(style.StyleHeading1)

	tableConfig := &document.TableConfig{
		Rows:  100,
		Cols:  10,
		Width: 14400, // 10英寸
	}
	table, _ := doc.AddTable(tableConfig)

	for row := 0; row < 100; row++ {
		for col := 0; col < 10; col++ {
			table.SetCellText(row, col, fmt.Sprintf("数据_%d_%d", row+1, col+1))
		}
	}

	doc.Save(fmt.Sprintf("../results/golang/fixed_large_table_%d.docx", index))

	return time.Since(start)
}

func testLargeDocumentFixed(index int) time.Duration {
	start := time.Now()

	doc := document.New()
	doc.AddParagraph("大型文档性能测试").SetStyle(style.StyleHeading1)

	// 添加1000个段落
	for j := 0; j < 1000; j++ {
		if j%10 == 0 {
			// 每10个段落添加一个标题
			doc.AddParagraph(fmt.Sprintf("章节 %d", j/10+1)).SetStyle(style.StyleHeading2)
		}

		doc.AddParagraph(fmt.Sprintf("这是第%d个段落。Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.", j+1))
	}

	// 添加一个中等大小的表格
	tableConfig := &document.TableConfig{
		Rows:  20,
		Cols:  8,
		Width: 11520, // 8英寸
	}
	table, _ := doc.AddTable(tableConfig)
	for row := 0; row < 20; row++ {
		for col := 0; col < 8; col++ {
			table.SetCellText(row, col, fmt.Sprintf("表格数据%d-%d", row+1, col+1))
		}
	}

	doc.Save(fmt.Sprintf("../results/golang/fixed_large_document_%d.docx", index))

	return time.Since(start)
}

func testMemoryUsageFixed(index int) time.Duration {
	start := time.Now()

	var m1, m2 runtime.MemStats
	runtime.GC()
	runtime.ReadMemStats(&m1)

	doc := document.New()

	// 创建复杂内容
	for j := 0; j < 100; j++ {
		doc.AddParagraph(fmt.Sprintf("段落%d: 这是一个测试段落，用于测试内存使用情况", j+1))
	}

	tableConfig := &document.TableConfig{
		Rows:  50,
		Cols:  6,
		Width: 8640, // 6英寸
	}
	table, _ := doc.AddTable(tableConfig)
	for row := 0; row < 50; row++ {
		for col := 0; col < 6; col++ {
			table.SetCellText(row, col, fmt.Sprintf("单元格%d-%d", row+1, col+1))
		}
	}

	runtime.ReadMemStats(&m2)

	doc.Save(fmt.Sprintf("../results/golang/fixed_memory_test_%d.docx", index))

	return time.Since(start)
}

// TestPerformanceComparison 性能对比测试（非基准测试，用于详细分析）
func TestPerformanceComparison(t *testing.T) {
	outputDir := "../results/golang"
	os.MkdirAll(outputDir, 0755)

	tests := []struct {
		name     string
		testFunc func() time.Duration
	}{
		{"基础文档创建", testBasicDocumentCreation},
		{"复杂格式化", testComplexFormatting},
		{"表格操作", testTableOperations},
		{"大表格处理", testLargeTableProcessing},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 运行3次取平均值
			var total time.Duration
			for i := 0; i < 3; i++ {
				duration := tt.testFunc()
				total += duration
			}
			avg := total / 3
			t.Logf("%s 平均耗时: %v", tt.name, avg)
		})
	}
}

func testBasicDocumentCreation() time.Duration {
	start := time.Now()

	doc := document.New()
	doc.AddParagraph("基础文档测试")
	doc.AddParagraph("这是一个性能测试文档")
	doc.Save("../results/golang/perf_basic.docx")

	return time.Since(start)
}

func testComplexFormatting() time.Duration {
	start := time.Now()

	doc := document.New()
	doc.AddParagraph("复杂格式化测试").SetStyle(style.StyleHeading1)

	para := doc.AddParagraph("")
	para.AddFormattedText("粗体", &document.TextFormat{Bold: true})
	para.AddFormattedText(" ", &document.TextFormat{})
	para.AddFormattedText("斜体", &document.TextFormat{Italic: true})
	para.AddFormattedText(" ", &document.TextFormat{})
	para.AddFormattedText("红色", &document.TextFormat{FontColor: "FF0000"})

	doc.Save("../results/golang/perf_complex.docx")

	return time.Since(start)
}

func testTableOperations() time.Duration {
	start := time.Now()

	doc := document.New()
	tableConfig := &document.TableConfig{
		Rows:  20,
		Cols:  5,
		Width: 7200, // 5英寸
	}
	table, _ := doc.AddTable(tableConfig)

	for row := 0; row < 20; row++ {
		for col := 0; col < 5; col++ {
			table.SetCellText(row, col, fmt.Sprintf("R%dC%d", row+1, col+1))
		}
	}

	doc.Save("../results/golang/perf_table.docx")

	return time.Since(start)
}

func testLargeTableProcessing() time.Duration {
	start := time.Now()

	doc := document.New()
	tableConfig := &document.TableConfig{
		Rows:  100,
		Cols:  8,
		Width: 11520, // 8英寸
	}
	table, _ := doc.AddTable(tableConfig)

	for row := 0; row < 100; row++ {
		for col := 0; col < 8; col++ {
			table.SetCellText(row, col, fmt.Sprintf("数据%d-%d", row+1, col+1))
		}
	}

	doc.Save("../results/golang/perf_large_table.docx")

	return time.Since(start)
}
