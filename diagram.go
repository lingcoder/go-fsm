package fsm

import (
	"fmt"
	"strings"
)

// GenerateDiagram returns a diagram of the state machine in the specified formats
// If formats is nil or empty, defaults to PlantUML
// If multiple formats are provided, returns all requested formats concatenated
func (sm *StateMachineImpl[S, E, P]) GenerateDiagram(formats ...DiagramFormat) string {
	if len(formats) == 0 {
		return sm.generatePlantUML()
	}

	var result strings.Builder
	for i, format := range formats {
		if i > 0 {
			result.WriteString("\n\n")
		}

		switch format {
		case MarkdownTable:
			result.WriteString(sm.generateMarkdownTable())
		case MarkdownFlowchart:
			result.WriteString(sm.generateMarkdownFlow())
		case MarkdownStateDiagram:
			result.WriteString(sm.generateMarkdownStateDiagram())
		case PlantUML:
			result.WriteString(sm.generatePlantUML())
		default:
			result.WriteString(sm.generatePlantUML())
		}
	}

	return result.String()
}

// generatePlantUML returns a PlantUML diagram of the state machine
func (sm *StateMachineImpl[S, E, P]) generatePlantUML() string {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	var sb strings.Builder
	sb.WriteString("@startuml\n")
	sb.WriteString(fmt.Sprintf("title StateMachine: %s\n", sm.id))

	// Define states
	for stateId := range sm.stateMap {
		sb.WriteString(fmt.Sprintf("state \"%v\" as %v\n", stateId, stateId))
	}

	// Define transitions
	for _, state := range sm.stateMap {
		for _, transitions := range state.eventTransitions {
			for _, transition := range transitions {
				sb.WriteString(fmt.Sprintf("%v --> %v : %v\n", transition.Source.id, transition.Target.id, transition.Event))
			}
		}
	}

	sb.WriteString("@enduml\n")
	return sb.String()
}

// generateMarkdownTable returns a Markdown table representation of the state machine
func (sm *StateMachineImpl[S, E, P]) generateMarkdownTable() string {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("# State Machine: %s\n\n", sm.id))

	// States section
	sb.WriteString("## States\n\n")
	for stateId := range sm.stateMap {
		sb.WriteString(fmt.Sprintf("- `%v`\n", stateId))
	}
	sb.WriteString("\n")

	// Transitions section
	sb.WriteString("## Transitions\n\n")
	sb.WriteString("| Source State | Event | Target State | Type |\n")
	sb.WriteString("|-------------|-------|--------------|------|\n")

	// Sort states for consistent output
	stateIds := make([]S, 0, len(sm.stateMap))
	for stateId := range sm.stateMap {
		stateIds = append(stateIds, stateId)
	}

	// Sort events for each state
	for _, sourceId := range stateIds {
		sourceState := sm.stateMap[sourceId]

		for event, transitions := range sourceState.eventTransitions {
			for _, transition := range transitions {
				transType := "External"
				if transition.TransType == Internal {
					transType = "Internal"
				}
				sb.WriteString(fmt.Sprintf("| `%v` | `%v` | `%v` | %s |\n",
					sourceId, event, transition.Target.id, transType))
			}
		}
	}

	return sb.String()
}

// generateMarkdownFlow returns a Mermaid flowchart diagram in Markdown format
func (sm *StateMachineImpl[S, E, P]) generateMarkdownFlow() string {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	var sb strings.Builder
	sb.WriteString("```mermaid\nflowchart TD\n")

	// Define node IDs - we need to ensure they are valid Mermaid IDs
	nodeIds := make(map[S]string)
	i := 0
	for stateId := range sm.stateMap {
		// Create a valid Mermaid ID (alphanumeric and underscores only)
		nodeIds[stateId] = fmt.Sprintf("state_%d", i)
		sb.WriteString(fmt.Sprintf("    %s[\"%v\"]\n", nodeIds[stateId], stateId))
		i++
	}

	// Define transitions
	for _, state := range sm.stateMap {
		sourceNodeId := nodeIds[state.id]

		for event, transitions := range state.eventTransitions {
			for _, transition := range transitions {
				targetNodeId := nodeIds[transition.Target.id]
				sb.WriteString(fmt.Sprintf("    %s -->|%v| %s\n",
					sourceNodeId, event, targetNodeId))
			}
		}
	}

	sb.WriteString("```\n")
	return sb.String()
}

// generateMarkdownStateDiagram returns a Mermaid state diagram in Markdown format
func (sm *StateMachineImpl[S, E, P]) generateMarkdownStateDiagram() string {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	var sb strings.Builder
	sb.WriteString("```mermaid\nstateDiagram-v2\n")

	// Add transitions (states are automatically created in Mermaid)
	for _, state := range sm.stateMap {
		for event, transitions := range state.eventTransitions {
			for _, transition := range transitions {
				if transition.TransType == External {
					sb.WriteString(fmt.Sprintf("    %v --> %v : %v\n",
						transition.Source.id, transition.Target.id, event))
				} else {
					sb.WriteString(fmt.Sprintf("    %v --> %v : %v [internal]\n",
						transition.Source.id, transition.Target.id, event))
				}
			}
		}
	}

	sb.WriteString("```\n")
	return sb.String()
}
