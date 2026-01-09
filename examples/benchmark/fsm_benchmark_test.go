package benchmark

import (
	"fmt"
	"testing"
	"time"

	"github.com/lingcoder/go-fsm"
)

// createComplexStateMachine creates a complex state machine for benchmarking
func createComplexStateMachine(b *testing.B) fsm.StateMachine[string, string, interface{}] {
	// Create a state machine
	builder := fsm.NewStateMachineBuilder[string, string, interface{}]()

	// Define states
	states := []string{"A", "B", "C", "D", "E", "F", "G", "H"}

	// Define events
	events := []string{"EVENT1", "EVENT2", "EVENT3", "EVENT4"}

	// Create a complex network of transitions
	for i, fromState := range states {
		for j, event := range events {
			// Create transitions to different states based on event
			toState := states[(i+j+1)%len(states)]

			builder.ExternalTransition().
				From(fromState).
				To(toState).
				On(event).
				WhenFunc(func(payload interface{}) bool {
					return true
				}).
				PerformFunc(func(from, to string, event string, payload interface{}) error {
					return nil
				})
		}
	}

	// Add some internal transitions
	for i, state := range states {
		if i%2 == 0 { // Only add for some states
			builder.InternalTransition().
				Within(state).
				On("INTERNAL").
				WhenFunc(func(payload interface{}) bool {
					return true
				}).
				PerformFunc(func(from, to string, event string, payload interface{}) error {
					return nil
				})
		}
	}

	// Add some parallel transitions
	builder.ExternalParallelTransition().
		From("A").
		ToAmong("B", "C", "D").
		On("PARALLEL").
		WhenFunc(func(payload interface{}) bool {
			return true
		}).
		PerformFunc(func(from, to string, event string, payload interface{}) error {
			return nil
		})

	// Add multiple source transitions
	builder.ExternalTransitions().
		FromAmong("E", "F", "G").
		To("H").
		On("MULTIPLE").
		WhenFunc(func(payload interface{}) bool {
			return true
		}).
		PerformFunc(func(from, to string, event string, payload interface{}) error {
			return nil
		})

	stateMachine, err := builder.Build(fmt.Sprintf("BenchmarkMachine-%d", time.Now().UnixNano()))
	if err != nil {
		b.Fatalf("Failed to build state machine: %v", err)
	}

	return stateMachine
}

// BenchmarkDiagramGeneration tests the performance of diagram generation
func BenchmarkDiagramGeneration(b *testing.B) {
	b.ReportAllocs()

	// Create a simple state machine for diagram generation
	builder := fsm.NewStateMachineBuilder[string, string, interface{}]()

	// Define multiple transitions
	builder.ExternalTransition().
		From("A").
		To("B").
		On("EVENT1").
		WhenFunc(func(payload interface{}) bool {
			return true
		}).
		PerformFunc(func(from, to string, event string, payload interface{}) error {
			return nil
		})

	builder.ExternalTransition().
		From("B").
		To("D").
		On("EVENT2").
		WhenFunc(func(payload interface{}) bool {
			return true
		}).
		PerformFunc(func(from, to string, event string, payload interface{}) error {
			return nil
		})

	builder.ExternalTransition().
		From("D").
		To("D").
		On("EVENT3").
		WhenFunc(func(payload interface{}) bool {
			return true
		}).
		PerformFunc(func(from, to string, event string, payload interface{}) error {
			return nil
		})

	builder.ExternalTransition().
		From("D").
		To("A").
		On("EVENT1").
		WhenFunc(func(payload interface{}) bool {
			return true
		}).
		PerformFunc(func(from, to string, event string, payload interface{}) error {
			return nil
		})

	stateMachine, err := builder.Build(fmt.Sprintf("DiagramMachine-%d", time.Now().UnixNano()))
	if err != nil {
		b.Fatalf("Failed to build state machine: %v", err)
	}

	// Run sub-benchmarks for different formats
	b.Run("PlantUML", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = stateMachine.GenerateDiagram(fsm.PlantUML)
		}
	})

	b.Run("MarkdownTable", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = stateMachine.GenerateDiagram(fsm.MarkdownTable)
		}
	})

	b.Run("MarkdownFlowchart", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = stateMachine.GenerateDiagram(fsm.MarkdownFlowchart)
		}
	})

	b.Run("MarkdownStateDiagram", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = stateMachine.GenerateDiagram(fsm.MarkdownStateDiagram)
		}
	})

	// Test combined formats
	b.Run("AllFormats", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = stateMachine.GenerateDiagram(fsm.PlantUML, fsm.MarkdownTable, fsm.MarkdownFlowchart, fsm.MarkdownStateDiagram)
		}
	})
}

// BenchmarkStateTransition tests the performance of state transitions
func BenchmarkStateTransition(b *testing.B) {
	stateMachine := createComplexStateMachine(b)

	// Define test cases with different complexity
	benchCases := []struct {
		name       string
		fromState  string
		event      string
		iterations int
	}{
		{"SimpleTransition", "A", "EVENT1", 1},
		{"MultipleTransitions", "A", "EVENT1", 10},
		{"ManyTransitions", "A", "EVENT1", 100},
	}

	for _, bc := range benchCases {
		b.Run(bc.name, func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				currentState := bc.fromState
				var err error

				// Perform multiple transitions in sequence
				for j := 0; j < bc.iterations; j++ {
					currentState, err = stateMachine.FireEvent(currentState, bc.event, nil)
					if err != nil {
						b.Fatalf("Failed to transition: %v", err)
					}
				}
			}
		})
	}
}

// BenchmarkParallelTransition tests the performance of parallel transitions
func BenchmarkParallelTransition(b *testing.B) {
	stateMachine := createComplexStateMachine(b)

	b.Run("ParallelTransition", func(b *testing.B) {
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_, err := stateMachine.FireEvent("A", "PARALLEL", nil)
			if err != nil {
				b.Fatalf("Failed to execute parallel transition: %v", err)
			}
		}
	})
}

// BenchmarkMultipleSourceTransition tests the performance of transitions with multiple source states
func BenchmarkMultipleSourceTransition(b *testing.B) {
	stateMachine := createComplexStateMachine(b)

	sources := []string{"E", "F", "G"}

	for _, source := range sources {
		b.Run(fmt.Sprintf("MultipleSourceFrom_%s", source), func(b *testing.B) {
			b.ResetTimer()

			for i := 0; i < b.N; i++ {
				_, err := stateMachine.FireEvent(source, "MULTIPLE", nil)
				if err != nil {
					b.Fatalf("Failed to execute multiple source transition: %v", err)
				}
			}
		})
	}
}

// BenchmarkStateMachineCreation tests the performance of creating state machines
func BenchmarkStateMachineCreation(b *testing.B) {
	b.Run("SimpleStateMachine", func(b *testing.B) {
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			builder := fsm.NewStateMachineBuilder[string, string, interface{}]()

			builder.ExternalTransition().
				From("A").
				To("B").
				On("EVENT").
				WhenFunc(func(payload interface{}) bool {
					return true
				}).
				PerformFunc(func(from, to string, event string, payload interface{}) error {
					return nil
				})

			_, err := builder.Build(fmt.Sprintf("SimpleMachine-%d", i))
			if err != nil {
				b.Fatalf("Failed to build state machine: %v", err)
			}
		}
	})

	b.Run("ComplexStateMachine", func(b *testing.B) {
		b.ResetTimer()

		for i := 0; i < b.N; i++ {
			_ = createComplexStateMachine(b)
		}
	})
}
