# Go-FSM: A Lightweight Finite State Machine for Go

[![Go Reference](https://pkg.go.dev/badge/github.com/lingcoder/go-fsm.svg)](https://pkg.go.dev/github.com/lingcoder/go-fsm)
[![Go Report Card](https://goreportcard.com/badge/github.com/lingcoder/go-fsm)](https://goreportcard.com/report/github.com/lingcoder/go-fsm)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Go-FSM is a lightweight, high-performance, stateless finite state machine implementation in Go, inspired by Alibaba's COLA state machine component.

<p align="center">
  <a href="README-zh.md">中文文档</a>
</p>

## ✨ Features

- 🪶 **Lightweight** - Minimal, stateless design for high performance
- 🔒 **Type-safe** - Built with Go generics for compile-time type checking
- 🔄 **Fluent API** - Intuitive builder pattern for defining state machines
- 🔀 **Versatile Transitions** - Support for external, internal, and parallel transitions
- 🧪 **Conditional Logic** - Flexible conditions to control when transitions occur
- 🎬 **Action Execution** - Custom actions that execute during transitions
- 🔄 **Thread-safe** - Designed for concurrent use in multi-threaded environments
- 📊 **Visualization** - Built-in support for generating state machine diagrams

## 📦 Installation

```bash
go get github.com/lingcoder/go-fsm
```

## 🚀 Usage

```go
package main

import (
	"fmt"
	"log"

	"github.com/lingcoder/go-fsm"
)

// Define states
type OrderState string

const (
	OrderCreated   OrderState = "CREATED"
	OrderPaid      OrderState = "PAID"
	OrderShipped   OrderState = "SHIPPED"
	OrderDelivered OrderState = "DELIVERED"
	OrderCancelled OrderState = "CANCELLED"
)

// Define events
type OrderEvent string

const (
	EventPay     OrderEvent = "PAY"
	EventShip    OrderEvent = "SHIP"
	EventDeliver OrderEvent = "DELIVER"
	EventCancel  OrderEvent = "CANCEL"
)

// Define payload
type OrderPayload struct {
	OrderID string
	Amount  float64
}

func main() {
	// Create a builder
	builder := fsm.NewStateMachineBuilder[OrderState, OrderEvent, OrderPayload]()

	// Define the state machine
	builder.ExternalTransition().
		From(OrderCreated).
		To(OrderPaid).
		On(EventPay).
		WhenFunc(func(payload OrderPayload) bool {
			// Check if amount is valid
			return payload.Amount > 0
		}).
		PerformFunc(func(from, to OrderState, event OrderEvent, payload OrderPayload) error {
			fmt.Printf("Order %s transitioning from %s to %s on event %s\n",
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
			fmt.Printf("Order %s is being shipped\n", payload.OrderID)
			return nil
		})

	// Define multiple source transitions
	builder.ExternalTransitions().
		FromAmong(OrderCreated, OrderPaid, OrderShipped).
		To(OrderCancelled).
		On(EventCancel).
		WhenFunc(func(payload OrderPayload) bool {
			return true
		}).
		PerformFunc(func(from, to OrderState, event OrderEvent, payload OrderPayload) error {
			fmt.Printf("Order %s cancelled from %s state\n", payload.OrderID, from)
			return nil
		})

	// Build the state machine
	stateMachine, err := builder.Build("OrderStateMachine")
	if err != nil {
		log.Fatalf("Failed to build state machine: %v", err)
	}

	// Create payload
	payload := OrderPayload{
		OrderID: "ORD-20250425-001",
		Amount:  100.0,
	}

	// Transition from CREATED to PAID
	newState, err := stateMachine.FireEvent(OrderCreated, EventPay, payload)
	if err != nil {
		log.Fatalf("Transition failed: %v", err)
	}

	fmt.Printf("New state: %v\n", newState)
}

```

## 🧩 Core Concepts

| Concept | Description |
|---------|-------------|
| **State** | Represents a specific state in your business process |
| **Event** | Triggers state transitions |
| **Transition** | Defines how states change in response to events |
| **Condition** | Logic that determines if a transition should occur |
| **Action** | Logic executed when a transition occurs |
| **StateMachine** | The core component that manages states and transitions |

### Types of Transitions

- **External Transition**: Transition between different states
- **Internal Transition**: Actions within the same state
- **Parallel Transition**: Transition to multiple states simultaneously. Fire it with `FireParallelEvent`; if an action fails, the remaining transitions are skipped and the states already reached are returned with the error

`When` is optional: `On(event).Perform(action)` defines a transition without condition. `Build` rejects definitions that never reach `Perform`, duplicate transitions, and transitions that can never fire.

### Registry

`Build` does not register state machines globally. Register one explicitly to look it up by ID later:

```go
if err := fsm.RegisterStateMachine("OrderStateMachine", stateMachine); err != nil {
	log.Fatal(err)
}

sm, err := fsm.GetStateMachine[OrderState, OrderEvent, OrderPayload]("OrderStateMachine")
```

## 📚 Examples

Check the `examples` directory for more detailed examples:

- `examples/order`: Order processing workflow
- `examples/workflow`: Approval workflow
- `examples/game`: Game state management

## ⚡ Performance

Go-FSM is designed for high performance:

- Stateless design minimizes memory usage
- Efficient transition lookup
- Thread-safe and lock-free: a built state machine is immutable, so concurrent `FireEvent` calls never contend
- Benchmarks included in the test suite

## 📊 Visualization

Go-FSM provides a unified way to visualize your state machine with different formats:

```go
// Default format (PlantUML)
plantUML := stateMachine.GenerateDiagram()
fmt.Println(plantUML)

// Generate specific format
table := stateMachine.GenerateDiagram(fsm.MarkdownTable)     // Markdown table format
fmt.Println(table)

flow := stateMachine.GenerateDiagram(fsm.MarkdownFlowchart)  // Markdown flowchart format
fmt.Println(flow)

// Generate multiple formats separately
diagrams := stateMachine.GenerateDiagram(fsm.PlantUML, fsm.MarkdownTable, fsm.MarkdownFlowchart, fsm.MarkdownStateDiagram)
fmt.Println(diagrams)
```

## 📄 License

[MIT](LICENSE) © LingCoder
