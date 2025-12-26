// Package main 模板功能演示示例
package main

import (
	"fmt"
	"log"
	"time"

	"github.com/ZeroHawkeye/wordZero/pkg/document"
)

func main() {
	fmt.Println("WordZero 模板功能演示")
	fmt.Println("=====================================")

	// 演示1: 基础变量替换
	fmt.Println("\n1. 基础变量替换演示")
	demonstrateVariableReplacement()

	// 演示2: 条件语句
	fmt.Println("\n2. 条件语句演示")
	demonstrateConditionalStatements()

	// 演示3: 循环语句
	fmt.Println("\n3. 循环语句演示")
	demonstrateLoopStatements()

	// 演示4: 模板继承
	fmt.Println("\n4. 模板继承演示")
	demonstrateTemplateInheritance()

	// 演示5: 复杂模板综合应用
	fmt.Println("\n5. 复杂模板综合应用")
	demonstrateComplexTemplate()

	// 演示6: 从现有文档创建模板
	fmt.Println("\n6. 从现有文档创建模板演示")
	demonstrateDocumentToTemplate()

	// 演示7: 结构体数据绑定
	fmt.Println("\n7. 结构体数据绑定演示")
	demonstrateStructDataBinding()

	fmt.Println("\n=====================================")
	fmt.Println("模板功能演示完成！")
	fmt.Println("生成的文档保存在 examples/output/ 目录下")
}

// demonstrateVariableReplacement 演示基础变量替换功能
func demonstrateVariableReplacement() {
	// 创建模板引擎
	engine := document.NewTemplateEngine()

	// 创建包含变量的模板
	templateContent := `尊敬的 {{customerName}} 先生/女士：

感谢您选择 {{companyName}}！

您的订单号是：{{orderNumber}}
订单金额：{{amount}} 元
下单时间：{{orderDate}}

我们将在 {{deliveryDays}} 个工作日内为您发货。

如有任何问题，请联系我们的客服热线：{{servicePhone}}

祝您生活愉快！

{{companyName}}
{{currentDate}}`

	// 加载模板
	template, err := engine.LoadTemplate("order_confirmation", templateContent)
	if err != nil {
		log.Fatalf("加载模板失败: %v", err)
	}

	fmt.Printf("解析到 %d 个变量\n", len(template.Variables))

	// 创建模板数据
	data := document.NewTemplateData()
	data.SetVariable("customerName", "张三")
	data.SetVariable("companyName", "WordZero科技有限公司")
	data.SetVariable("orderNumber", "WZ20241201001")
	data.SetVariable("amount", "1299.00")
	data.SetVariable("orderDate", "2024年12月1日 14:30")
	data.SetVariable("deliveryDays", "3-5")
	data.SetVariable("servicePhone", "400-123-4567")
	data.SetVariable("currentDate", time.Now().Format("2006年01月02日"))

	// 渲染模板
	doc, err := engine.RenderToDocument("order_confirmation", data)
	if err != nil {
		log.Fatalf("渲染模板失败: %v", err)
	}

	// 保存文档
	err = doc.Save("examples/output/template_variable_demo.docx")
	if err != nil {
		log.Fatalf("保存文档失败: %v", err)
	}

	fmt.Println("✓ 变量替换演示完成，文档已保存为 template_variable_demo.docx")
}

// demonstrateConditionalStatements 演示条件语句功能
func demonstrateConditionalStatements() {
	engine := document.NewTemplateEngine()

	// 创建包含条件语句的模板
	templateContent := `产品推荐信

尊敬的客户：

{{#if isVipCustomer}}
作为我们的VIP客户，您将享受以下特殊优惠：
- 全场商品9折优惠
- 免费包邮服务
- 优先客服支持
{{/if}}

{{#if hasNewProducts}}
最新产品推荐：
我们刚刚推出了一系列新产品，相信您会喜欢。
{{/if}}

{{#if showDiscount}}
限时优惠：
现在购买任意商品，立享8折优惠！
优惠码：SAVE20
{{/if}}

{{#if needSupport}}
如需技术支持，请联系我们的专业团队。
支持热线：400-888-9999
{{/if}}

感谢您的信任与支持！

WordZero团队`

	// 加载模板
	_, err := engine.LoadTemplate("product_recommendation", templateContent)
	if err != nil {
		log.Fatalf("加载模板失败: %v", err)
	}

	// 测试不同条件组合
	scenarios := []struct {
		name         string
		isVip        bool
		hasNew       bool
		showDiscount bool
		needSupport  bool
		filename     string
	}{
		{"VIP客户场景", true, true, false, true, "template_conditional_vip.docx"},
		{"普通客户场景", false, true, true, false, "template_conditional_normal.docx"},
		{"简单推荐场景", false, false, false, false, "template_conditional_simple.docx"},
	}

	for _, scenario := range scenarios {
		fmt.Printf("生成 %s...\n", scenario.name)

		data := document.NewTemplateData()
		data.SetCondition("isVipCustomer", scenario.isVip)
		data.SetCondition("hasNewProducts", scenario.hasNew)
		data.SetCondition("showDiscount", scenario.showDiscount)
		data.SetCondition("needSupport", scenario.needSupport)

		doc, err := engine.RenderToDocument("product_recommendation", data)
		if err != nil {
			log.Fatalf("渲染模板失败: %v", err)
		}

		err = doc.Save("examples/output/" + scenario.filename)
		if err != nil {
			log.Fatalf("保存文档失败: %v", err)
		}

		fmt.Printf("✓ %s 完成\n", scenario.name)
	}
}

// demonstrateLoopStatements 演示循环语句功能
func demonstrateLoopStatements() {
	engine := document.NewTemplateEngine()

	// 创建包含循环语句的模板
	templateContent := `销售报告

报告时间：{{reportDate}}
销售部门：{{department}}

产品销售明细：
{{#each products}}
{{@index}}. 产品名称：{{name}}
   销售数量：{{quantity}} 件
   单价：{{price}} 元
   销售金额：{{total}} 元
   {{#if isTopSeller}}🏆 热销产品{{/if}}

{{/each}}

销售统计：
总销售金额：{{totalAmount}} 元
平均客单价：{{averagePrice}} 元

{{#each salespeople}}
销售员：{{name}} - 业绩：{{performance}} 元
{{/each}}

备注：
{{#each notes}}
- {{this}}
{{/each}}`

	// 加载模板
	_, err := engine.LoadTemplate("sales_report", templateContent)
	if err != nil {
		log.Fatalf("加载模板失败: %v", err)
	}

	// 创建模板数据
	data := document.NewTemplateData()
	data.SetVariable("reportDate", "2024年12月1日")
	data.SetVariable("department", "华东区销售部")
	data.SetVariable("totalAmount", "89,650")
	data.SetVariable("averagePrice", "1,245")

	// 设置产品列表
	products := []interface{}{
		map[string]interface{}{
			"name":        "iPhone 15 Pro",
			"quantity":    25,
			"price":       8999,
			"total":       224975,
			"isTopSeller": true,
		},
		map[string]interface{}{
			"name":        "iPad Air",
			"quantity":    18,
			"price":       4999,
			"total":       89982,
			"isTopSeller": false,
		},
		map[string]interface{}{
			"name":        "MacBook Pro",
			"quantity":    8,
			"price":       16999,
			"total":       135992,
			"isTopSeller": true,
		},
	}
	data.SetList("products", products)

	// 设置销售员列表
	salespeople := []interface{}{
		map[string]interface{}{
			"name":        "王小明",
			"performance": 156800,
		},
		map[string]interface{}{
			"name":        "李小红",
			"performance": 134500,
		},
		map[string]interface{}{
			"name":        "张小强",
			"performance": 98750,
		},
	}
	data.SetList("salespeople", salespeople)

	// 设置备注列表
	notes := []interface{}{
		"本月销售表现优异，超额完成目标",
		"iPhone 15 Pro 持续热销",
		"建议增加库存以满足需求",
	}
	data.SetList("notes", notes)

	// 渲染模板
	doc, err := engine.RenderToDocument("sales_report", data)
	if err != nil {
		log.Fatalf("渲染模板失败: %v", err)
	}

	// 保存文档
	err = doc.Save("examples/output/template_loop_demo.docx")
	if err != nil {
		log.Fatalf("保存文档失败: %v", err)
	}

	fmt.Println("✓ 循环语句演示完成，文档已保存为 template_loop_demo.docx")
}

// demonstrateTemplateInheritance 演示模板继承功能
func demonstrateTemplateInheritance() {
	engine := document.NewTemplateEngine()

	// 创建基础模板
	baseTemplateContent := `{{companyName}} 官方文档

文档标题：{{title}}
创建时间：{{createDate}}
版本号：{{version}}

---

文档内容：`

	// 加载基础模板
	_, err := engine.LoadTemplate("base_document", baseTemplateContent)
	if err != nil {
		log.Fatalf("加载基础模板失败: %v", err)
	}

	// 创建继承模板 - 用户手册
	userManualContent := `{{extends "base_document"}}

用户手册

第一章：快速开始
欢迎使用我们的产品！本章将帮助您快速上手。

第二章：基础功能
介绍产品的基础功能和使用方法。

第三章：高级功能
深入了解产品的高级特性。

第四章：常见问题
解答用户常见的问题和疑惑。

如需更多帮助，请联系技术支持。`

	// 加载用户手册模板
	_, err = engine.LoadTemplate("user_manual", userManualContent)
	if err != nil {
		log.Fatalf("加载用户手册模板失败: %v", err)
	}

	// 创建继承模板 - API文档
	apiDocContent := `{{extends "base_document"}}

API接口文档

接口概述：
本文档提供了完整的API接口说明。

认证方式：
使用API Key进行身份验证。

接口列表：
1. GET /api/users - 获取用户列表
2. POST /api/users - 创建新用户
3. PUT /api/users/{id} - 更新用户信息
4. DELETE /api/users/{id} - 删除用户

错误代码：
- 400: 请求参数错误
- 401: 认证失败
- 404: 资源不存在
- 500: 服务器内部错误`

	// 加载API文档模板
	_, err = engine.LoadTemplate("api_document", apiDocContent)
	if err != nil {
		log.Fatalf("加载API文档模板失败: %v", err)
	}

	// 创建通用数据
	commonData := document.NewTemplateData()
	commonData.SetVariable("companyName", "WordZero科技")
	commonData.SetVariable("createDate", time.Now().Format("2006年01月02日"))
	commonData.SetVariable("version", "v1.0")

	// 生成用户手册
	userManualData := document.NewTemplateData()
	userManualData.Merge(commonData)
	userManualData.SetVariable("title", "产品用户手册")

	userManualDoc, err := engine.RenderToDocument("user_manual", userManualData)
	if err != nil {
		log.Fatalf("渲染用户手册失败: %v", err)
	}

	err = userManualDoc.Save("examples/output/template_inheritance_user_manual.docx")
	if err != nil {
		log.Fatalf("保存用户手册失败: %v", err)
	}

	// 生成API文档
	apiDocData := document.NewTemplateData()
	apiDocData.Merge(commonData)
	apiDocData.SetVariable("title", "API接口文档")

	apiDoc, err := engine.RenderToDocument("api_document", apiDocData)
	if err != nil {
		log.Fatalf("渲染API文档失败: %v", err)
	}

	err = apiDoc.Save("examples/output/template_inheritance_api_doc.docx")
	if err != nil {
		log.Fatalf("保存API文档失败: %v", err)
	}

	fmt.Println("✓ 模板继承演示完成")
	fmt.Println("  - 用户手册已保存为 template_inheritance_user_manual.docx")
	fmt.Println("  - API文档已保存为 template_inheritance_api_doc.docx")
}

// demonstrateComplexTemplate 演示复杂模板综合应用
func demonstrateComplexTemplate() {
	engine := document.NewTemplateEngine()

	// 创建复杂的项目报告模板
	complexTemplateContent := `{{companyName}} 项目报告

项目名称：{{projectName}}
项目经理：{{projectManager}}
报告日期：{{reportDate}}

===================================

项目概要：
{{projectDescription}}

项目状态：{{projectStatus}}

{{#if showTeamMembers}}
项目团队：
{{#each teamMembers}}
{{@index}}. 姓名：{{name}}
   职位：{{position}}
   工作内容：{{responsibility}}
   {{#if isLeader}}👑 团队负责人{{/if}}

{{/each}}
{{/if}}

{{#if showTasks}}
任务清单：
{{#each tasks}}
任务 {{@index}}: {{title}}
状态：{{status}}
{{#if isCompleted}}✅ 已完成{{/if}}
{{#if isInProgress}}🔄 进行中{{/if}}
{{#if isPending}}⏳ 待开始{{/if}}

描述：{{description}}

{{/each}}
{{/if}}

{{#if showMilestones}}
项目里程碑：
{{#each milestones}}
{{date}} - {{title}}
{{#if isCompleted}}✅ 已完成{{/if}}
{{#if isCurrent}}🎯 当前阶段{{/if}}

{{/each}}
{{/if}}

项目风险：
{{#each risks}}
- 风险：{{description}}
  等级：{{level}}
  应对措施：{{mitigation}}

{{/each}}

{{#if showBudget}}
预算信息：
总预算：{{totalBudget}} 万元
已使用：{{usedBudget}} 万元
剩余：{{remainingBudget}} 万元
{{/if}}

下一步计划：
{{#each nextSteps}}
- {{this}}
{{/each}}

===================================

报告人：{{reporter}}
审核人：{{reviewer}}`

	// 加载模板
	_, err := engine.LoadTemplate("project_report", complexTemplateContent)
	if err != nil {
		log.Fatalf("加载复杂模板失败: %v", err)
	}

	// 创建复杂数据
	data := document.NewTemplateData()

	// 基础信息
	data.SetVariable("companyName", "WordZero科技有限公司")
	data.SetVariable("projectName", "新一代文档处理系统")
	data.SetVariable("projectManager", "李项目")
	data.SetVariable("reportDate", "2024年12月1日")
	data.SetVariable("projectDescription", "开发一个功能强大、易于使用的Word文档操作库，支持模板引擎、样式管理等高级功能。")
	data.SetVariable("projectStatus", "进行中 - 80%完成")
	data.SetVariable("reporter", "李项目")
	data.SetVariable("reviewer", "王总监")

	// 条件设置
	data.SetCondition("showTeamMembers", true)
	data.SetCondition("showTasks", true)
	data.SetCondition("showMilestones", true)
	data.SetCondition("showBudget", true)

	// 团队成员
	teamMembers := []interface{}{
		map[string]interface{}{
			"name":           "张开发",
			"position":       "高级开发工程师",
			"responsibility": "核心功能开发",
			"isLeader":       true,
		},
		map[string]interface{}{
			"name":           "王测试",
			"position":       "测试工程师",
			"responsibility": "功能测试与质量保证",
			"isLeader":       false,
		},
		map[string]interface{}{
			"name":           "刘设计",
			"position":       "UI/UX设计师",
			"responsibility": "用户界面设计",
			"isLeader":       false,
		},
	}
	data.SetList("teamMembers", teamMembers)

	// 任务清单
	tasks := []interface{}{
		map[string]interface{}{
			"title":        "模板引擎开发",
			"status":       "已完成",
			"description":  "实现变量替换、条件语句、循环语句等功能",
			"isCompleted":  true,
			"isInProgress": false,
			"isPending":    false,
		},
		map[string]interface{}{
			"title":        "样式管理系统",
			"status":       "进行中",
			"description":  "完善样式继承和应用机制",
			"isCompleted":  false,
			"isInProgress": true,
			"isPending":    false,
		},
		map[string]interface{}{
			"title":        "性能优化",
			"status":       "待开始",
			"description":  "优化大文档处理性能",
			"isCompleted":  false,
			"isInProgress": false,
			"isPending":    true,
		},
	}
	data.SetList("tasks", tasks)

	// 项目里程碑
	milestones := []interface{}{
		map[string]interface{}{
			"date":        "2024年10月15日",
			"title":       "项目启动",
			"isCompleted": true,
			"isCurrent":   false,
		},
		map[string]interface{}{
			"date":        "2024年11月30日",
			"title":       "核心功能完成",
			"isCompleted": true,
			"isCurrent":   false,
		},
		map[string]interface{}{
			"date":        "2024年12月15日",
			"title":       "测试阶段",
			"isCompleted": false,
			"isCurrent":   true,
		},
	}
	data.SetList("milestones", milestones)

	// 项目风险
	risks := []interface{}{
		map[string]interface{}{
			"description": "技术难度超预期",
			"level":       "中等",
			"mitigation":  "增加技术调研时间，寻求外部专家支持",
		},
		map[string]interface{}{
			"description": "人员流动风险",
			"level":       "低",
			"mitigation":  "建立完善的文档和知识传承机制",
		},
	}
	data.SetList("risks", risks)

	// 预算信息
	data.SetVariable("totalBudget", "50")
	data.SetVariable("usedBudget", "35")
	data.SetVariable("remainingBudget", "15")

	// 下一步计划
	nextSteps := []interface{}{
		"完成剩余功能开发",
		"进行全面测试",
		"编写使用文档",
		"准备产品发布",
	}
	data.SetList("nextSteps", nextSteps)

	// 渲染模板
	doc, err := engine.RenderToDocument("project_report", data)
	if err != nil {
		log.Fatalf("渲染复杂模板失败: %v", err)
	}

	// 保存文档
	err = doc.Save("examples/output/template_complex_demo.docx")
	if err != nil {
		log.Fatalf("保存复杂模板文档失败: %v", err)
	}

	fmt.Println("✓ 复杂模板演示完成，文档已保存为 template_complex_demo.docx")
}

// demonstrateDocumentToTemplate 演示从现有文档创建模板
func demonstrateDocumentToTemplate() {
	// 创建一个基础文档
	doc := document.New()
	doc.AddParagraph("公司：{{companyName}}")
	doc.AddParagraph("部门：{{department}}")
	doc.AddParagraph("")
	doc.AddParagraph("员工信息：")
	doc.AddParagraph("姓名：{{employeeName}}")
	doc.AddParagraph("职位：{{position}}")
	doc.AddParagraph("入职日期：{{hireDate}}")

	// 创建模板引擎
	engine := document.NewTemplateEngine()

	// 从文档创建模板
	template, err := engine.LoadTemplateFromDocument("employee_template", doc)
	if err != nil {
		log.Fatalf("从文档创建模板失败: %v", err)
	}

	fmt.Printf("从文档解析到 %d 个变量\n", len(template.Variables))

	// 创建员工数据
	data := document.NewTemplateData()
	data.SetVariable("companyName", "WordZero科技有限公司")
	data.SetVariable("department", "研发部")
	data.SetVariable("employeeName", "李小明")
	data.SetVariable("position", "软件工程师")
	data.SetVariable("hireDate", "2024年12月1日")

	// 渲染模板
	renderedDoc, err := engine.RenderToDocument("employee_template", data)
	if err != nil {
		log.Fatalf("渲染员工模板失败: %v", err)
	}

	// 保存文档
	err = renderedDoc.Save("examples/output/template_from_document_demo.docx")
	if err != nil {
		log.Fatalf("保存文档失败: %v", err)
	}

	fmt.Println("✓ 从文档创建模板演示完成，文档已保存为 template_from_document_demo.docx")
}

// demonstrateStructDataBinding 演示结构体数据绑定
func demonstrateStructDataBinding() {
	// 定义数据结构
	type Employee struct {
		Name       string
		Position   string
		Department string
		Salary     int
		IsManager  bool
		HireDate   string
	}

	type Company struct {
		Name    string
		Address string
		Phone   string
		Website string
		Founded int
	}

	// 创建数据实例
	employee := Employee{
		Name:       "王小红",
		Position:   "产品经理",
		Department: "产品部",
		Salary:     15000,
		IsManager:  true,
		HireDate:   "2023年3月15日",
	}

	company := Company{
		Name:    "WordZero科技有限公司",
		Address: "上海市浦东新区科技园区1号楼",
		Phone:   "021-12345678",
		Website: "www.wordzero.com",
		Founded: 2023,
	}

	// 创建模板引擎
	engine := document.NewTemplateEngine()

	// 创建员工档案模板
	templateContent := `员工档案

公司信息：
公司名称：{{name}}
公司地址：{{address}}
联系电话：{{phone}}
公司网站：{{website}}
成立年份：{{founded}}

员工信息：
姓名：{{name}}
职位：{{position}}
部门：{{department}}
薪资：{{salary}} 元
入职日期：{{hiredate}}

{{#if ismanager}}
管理职责：
作为部门经理，负责团队管理和项目协调。
{{/if}}`

	// 加载模板
	_, err := engine.LoadTemplate("employee_profile", templateContent)
	if err != nil {
		log.Fatalf("加载员工档案模板失败: %v", err)
	}

	// 创建模板数据并从结构体填充
	data := document.NewTemplateData()

	// 从公司结构体创建数据
	err = data.FromStruct(company)
	if err != nil {
		log.Fatalf("从公司结构体创建数据失败: %v", err)
	}

	// 创建临时数据用于员工信息（避免字段名冲突）
	employeeData := document.NewTemplateData()
	err = employeeData.FromStruct(employee)
	if err != nil {
		log.Fatalf("从员工结构体创建数据失败: %v", err)
	}

	// 手动设置员工相关变量（处理字段名冲突）
	data.SetVariable("name", employee.Name)
	data.SetVariable("position", employee.Position)
	data.SetVariable("department", employee.Department)
	data.SetVariable("salary", employee.Salary)
	data.SetVariable("hiredate", employee.HireDate)
	data.SetCondition("ismanager", employee.IsManager)

	// 设置公司相关变量
	data.SetVariable("name", company.Name)
	data.SetVariable("address", company.Address)
	data.SetVariable("phone", company.Phone)
	data.SetVariable("website", company.Website)
	data.SetVariable("founded", company.Founded)

	// 渲染模板
	doc, err := engine.RenderToDocument("employee_profile", data)
	if err != nil {
		log.Fatalf("渲染员工档案失败: %v", err)
	}

	// 保存文档
	err = doc.Save("examples/output/template_struct_binding_demo.docx")
	if err != nil {
		log.Fatalf("保存员工档案失败: %v", err)
	}

	fmt.Println("✓ 结构体数据绑定演示完成，文档已保存为 template_struct_binding_demo.docx")
}
