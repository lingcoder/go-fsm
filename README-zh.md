# Go-FSM: Go 语言轻量级有限状态机

[![Go Reference](https://pkg.go.dev/badge/github.com/lingcoder/go-fsm.svg)](https://pkg.go.dev/github.com/lingcoder/go-fsm)
[![Go Report Card](https://goreportcard.com/badge/github.com/lingcoder/go-fsm)](https://goreportcard.com/report/github.com/lingcoder/go-fsm)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Go-FSM 是一个轻量级、高性能、无状态的有限状态机 Go 实现，灵感来自阿里巴巴的 COLA 状态机组件。

<p align="center">
  <a href="README.md">English Documentation</a>
</p>

## ✨ 特性

- 🪶 **轻量级** - 极简的无状态设计，提供高性能
- 🔒 **类型安全** - 使用 Go 泛型实现编译时类型检查
- 🔄 **流畅的 API** - 直观的构建器模式用于定义状态机
- 🔀 **多样化转换** - 支持外部、内部和并行状态转换
- 🧪 **条件逻辑** - 灵活的条件控制何时进行状态转换
- 🎬 **动作执行** - 转换过程中执行的自定义动作
- 🔄 **线程安全** - 为多线程环境下的并发使用而设计
- 📊 **可视化** - 内置支持生成状态机图表

## 📦 安装

```bash
go get github.com/lingcoder/go-fsm
```

## 🚀 使用方法

```go
package main

import (
	"fmt"
	"log"

	"github.com/lingcoder/go-fsm"
)

// 定义状态
type OrderState string

const (
	OrderCreated   OrderState = "CREATED"
	OrderPaid      OrderState = "PAID"
	OrderShipped   OrderState = "SHIPPED"
	OrderDelivered OrderState = "DELIVERED"
	OrderCancelled OrderState = "CANCELLED"
)

// 定义事件
type OrderEvent string

const (
	EventPay     OrderEvent = "PAY"
	EventShip    OrderEvent = "SHIP"
	EventDeliver OrderEvent = "DELIVER"
	EventCancel  OrderEvent = "CANCEL"
)

// 定义载荷
type OrderPayload struct {
	OrderID string
	Amount  float64
}

func main() {
	// 创建构建器
	builder := fsm.NewStateMachineBuilder[OrderState, OrderEvent, OrderPayload]()

	// 定义状态机
	builder.ExternalTransition().
		From(OrderCreated).
		To(OrderPaid).
		On(EventPay).
		WhenFunc(func(payload OrderPayload) bool {
			// 检查金额是否有效
			return payload.Amount > 0
		}).
		PerformFunc(func(from, to OrderState, event OrderEvent, payload OrderPayload) error {
			fmt.Printf("订单 %s 从 %s 状态转换到 %s 状态，触发事件: %s\n",
				payload.OrderID, from, to, event)
			return nil
		})

	builder.ExternalTransition().
		From(OrderPaid).
		To(OrderShipped).
		On(EventShip).
		WhenFunc(func(payload OrderPayload) bool {
			return true
		}).
		PerformFunc(func(from, to OrderState, event OrderEvent, payload OrderPayload) error {
			fmt.Printf("订单 %s 正在发货\n", payload.OrderID)
			return nil
		})

	// 定义多源状态转换
	builder.ExternalTransitions().
		FromAmong(OrderCreated, OrderPaid, OrderShipped).
		To(OrderCancelled).
		On(EventCancel).
		WhenFunc(func(payload OrderPayload) bool {
			return true
		}).
		PerformFunc(func(from, to OrderState, event OrderEvent, payload OrderPayload) error {
			fmt.Printf("订单 %s 从 %s 状态取消\n", payload.OrderID, from)
			return nil
		})

	// 构建状态机
	stateMachine, err := builder.Build("OrderStateMachine")
	if err != nil {
		log.Fatalf("构建状态机失败: %v", err)
	}

	// 创建载荷
	payload := OrderPayload{
		OrderID: "ORD-20250425-001",
		Amount:  100.0,
	}

	// 从 CREATED 转换到 PAID
	newState, err := stateMachine.FireEvent(OrderCreated, EventPay, payload)
	if err != nil {
		log.Fatalf("状态转换失败: %v", err)
	}

	fmt.Printf("新状态: %v\n", newState)
}
```

## 🧩 核心概念

| 概念 | 描述 |
|------|------|
| **状态 (State)** | 表示业务流程中的特定状态 |
| **事件 (Event)** | 触发状态转换 |
| **转换 (Transition)** | 定义状态如何响应事件而变化 |
| **条件 (Condition)** | 决定是否应该发生转换的逻辑 |
| **动作 (Action)** | 转换发生时执行的逻辑 |
| **状态机 (StateMachine)** | 管理状态和转换的核心组件 |

### 转换类型

- **外部转换 (External Transition)**: 不同状态之间的转换
- **内部转换 (Internal Transition)**: 同一状态内的动作
- **并行转换 (Parallel Transition)**: 同时转换到多个状态。需要用 `FireParallelEvent` 触发；某个动作失败时会跳过剩余的转换，并把已经到达的状态和错误一起返回

`When` 可以省略：`On(event).Perform(action)` 定义一个无条件转换。与 COLA 一致，同一状态和事件下有多个转换时，第一个满足条件的转换生效，无条件转换作为兜底。`Build` 会拒绝没有调用 `Perform` 的定义、源状态、事件和目标状态都相同的重复转换，以及同一状态和事件下的第二个兜底转换。

### 注册表

`Build` 不会把状态机注册到全局。如果之后需要按 ID 查找，请显式注册：

```go
if err := fsm.RegisterStateMachine("OrderStateMachine", stateMachine); err != nil {
	log.Fatal(err)
}

sm, err := fsm.GetStateMachine[OrderState, OrderEvent, OrderPayload]("OrderStateMachine")
```

### 传递 Context

与 COLA 一致，条件和动作只接收载荷。如果动作需要 `context.Context` 来控制超时、取消或做链路追踪，请把它放在载荷里：

```go
type OrderPayload struct {
	Ctx     context.Context
	OrderID string
}

builder.ExternalTransition().
	From(OrderCreated).
	To(OrderPaid).
	On(EventPay).
	PerformFunc(func(from, to OrderState, event OrderEvent, payload OrderPayload) error {
		_, err := db.ExecContext(payload.Ctx, "UPDATE orders SET status = ? WHERE id = ?", to, payload.OrderID)
		return err
	})

newState, err := stateMachine.FireEvent(OrderCreated, EventPay, OrderPayload{Ctx: ctx, OrderID: orderID})
```

每次调用都新建一个载荷，不要把它保存下来，这样 context 不会比请求活得更久。

## 📚 示例

查看 `examples` 目录获取更详细的示例：

- `examples/order`: 订单处理工作流
- `examples/workflow`: 审批工作流
- `examples/game`: 游戏状态管理

## ⚡ 性能

Go-FSM 设计注重高性能：

- 无状态设计最小化内存使用
- 高效的转换查找
- 线程安全且无锁：构建后的状态机不可变，并发调用 `FireEvent` 不会互相争用
- 测试套件中包含基准测试

## 📊 可视化

Go-FSM 提供一种统一的方式来可视化状态机：

```go
// 默认格式 (PlantUML)
plantUML := stateMachine.GenerateDiagram()
fmt.Println(plantUML)

// 生成特定格式
table := stateMachine.GenerateDiagram(fsm.MarkdownTable)     // Markdown 表格格式
fmt.Println(table)

flow := stateMachine.GenerateDiagram(fsm.MarkdownFlowchart)  // Markdown 流程图格式
fmt.Println(flow)

// 分别生成多种格式
diagrams := stateMachine.GenerateDiagram(fsm.PlantUML, fsm.MarkdownTable, fsm.MarkdownFlowchart, fsm.MarkdownStateDiagram)
fmt.Println(diagrams)
```

状态和转换按名称排序输出，每次生成的结果都一致。状态名和事件名可以包含空格、引号或 `|`：图中节点使用生成的 ID，标签会自动转义。

## 📄 许可证

[MIT](LICENSE) © LingCoder
