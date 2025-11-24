# 嵌套循环示例 (Nested Loop Demo)

## 概述

本示例演示了WordZero模板引擎如何处理嵌套的`{{#each}}`循环。这是一个常见的使用场景，例如在生成会议纪要时，需要列出每个参会人员及其对应的任务列表。

## 功能特性

- ✅ 支持任意层级的嵌套循环
- ✅ 在嵌套循环中使用`{{@index}}`自动编号
- ✅ 在循环内部使用条件判断`{{#if}}`
- ✅ 访问外层循环的数据

## 使用场景

嵌套循环适用于以下场景：

1. **会议纪要**：参会人员 → 每人的任务列表
2. **组织架构**：部门 → 团队 → 成员
3. **产品目录**：分类 → 子分类 → 产品列表
4. **项目报告**：项目 → 模块 → 功能点

## 模板语法

### 基本嵌套循环

```
{{#each outerList}}
外层项目：{{outerField}}
  {{#each innerList}}
  内层项目：{{innerField}}
  {{/each}}
{{/each}}
```

### 使用索引

```
{{#each items}}
第 {{@index}} 项：{{name}}
  {{#each subitems}}
  子项 {{@index}}：{{subname}}
  {{/each}}
{{/each}}
```

### 结合条件判断

```
{{#each items}}
项目：{{name}}
  {{#each tasks}}
  任务：{{taskName}}
  {{#if isCompleted}}✅ 已完成{{/if}}
  {{/each}}
{{/each}}
```

## 运行示例

```bash
cd examples/nested_loop_demo
go run main.go
```

生成的文档将保存在`examples/output/`目录下。

## 数据结构

示例使用的数据结构：

```go
attendees := []interface{}{
    map[string]interface{}{
        "name": "张三",
        "role": "项目经理",
        "tasks": []interface{}{
            map[string]interface{}{
                "taskName": "完成项目进度报告",
                "priority": "高",
                "status":   "进行中",
            },
            // ... 更多任务
        },
    },
    // ... 更多参会人员
}
```

## 注意事项

1. **性能考虑**：嵌套循环会增加文档生成时间，建议合理控制数据量
2. **数据结构**：内层列表应该是外层map中的一个字段，类型为`[]interface{}`
3. **变量作用域**：内层循环可以访问外层循环的变量
4. **索引编号**：每一层循环都有自己独立的`{{@index}}`

## 更多示例

请参考：
- `examples/template_demo/main.go` - 基础模板功能演示
- `examples/enhanced_template_demo/enhanced_template_demo.go` - 增强模板功能
- `test/template_test.go` - 嵌套循环单元测试
