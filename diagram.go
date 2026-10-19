package fsm

import (
	"fmt"
	"strings"
)

// Escapers for labels, so state and event names with spaces, quotes, pipes or
// line breaks can't break diagram syntax. Nodes always use generated aliases.
var (
	// plantUMLEscaper keeps a label on one line and inside its double quotes
	plantUMLEscaper = strings.NewReplacer("\r", "", "\n", `\n`, `"`, `'`)

	// mermaidEscaper keeps a label on one line and replaces double quotes with a Mermaid entity code
	mermaidEscaper = strings.NewReplacer("\r", "", "\n", " ", `"`, "#quot;")

	// markdownTextEscaper keeps a value on one line
	markdownTextEscaper = strings.NewReplacer("\r", "", "\n", " ")

	// markdownCellEscaper keeps a value on one line and inside its table cell
	markdownCellEscaper = strings.NewReplacer("\r", "", "\n", " ", "|", `\|`)
)

// escapeLabel formats v and escapes it with the given escaper
func escapeLabel(escaper *strings.Replacer, v any) string {
	return escaper.Replace(fmt.Sprint(v))
}

// GenerateDiagram returns a diagram of the state machine in the specified formats
// If formats is nil or empty, defaults to PlantUML
// If multiple formats are provided, returns all requested formats concatenated
// States and transitions are sorted by name, so the output is stable
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

// stateAliases returns the states sorted by name and a diagram identifier for each
// that is valid whatever the state's name
func (sm *StateMachineImpl[S, E, P]) stateAliases(prefix string) ([]S, map[S]string) {
	states := sortedKeys(sm.stateMap)
	aliases := make(map[S]string, len(states))
	for i, stateId := range states {
		aliases[stateId] = fmt.Sprintf("%s%d", prefix, i)
	}
	return states, aliases
}

// generatePlantUML returns a PlantUML diagram of the state machine
func (sm *StateMachineImpl[S, E, P]) generatePlantUML() string {
	states, aliases := sm.stateAliases("s")

	var sb strings.Builder
	sb.WriteString("@startuml\n")
	fmt.Fprintf(&sb, "title StateMachine: %s\n", escapeLabel(plantUMLEscaper, sm.id))

	// Define states
	for _, stateId := range states {
		fmt.Fprintf(&sb, "state \"%s\" as %s\n", escapeLabel(plantUMLEscaper, stateId), aliases[stateId])
	}

	// Define transitions
	sm.forEachTransition(func(transition *Transition[S, E, P]) {
		fmt.Fprintf(&sb, "%s --> %s : %s\n",
			aliases[transition.Source.id], aliases[transition.Target.id], escapeLabel(plantUMLEscaper, transition.Event))
	})

	sb.WriteString("@enduml\n")
	return sb.String()
}

// generateMarkdownTable returns a Markdown table representation of the state machine
func (sm *StateMachineImpl[S, E, P]) generateMarkdownTable() string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "# State Machine: %s\n\n", escapeLabel(markdownTextEscaper, sm.id))

	// States section
	sb.WriteString("## States\n\n")
	for _, stateId := range sortedKeys(sm.stateMap) {
		fmt.Fprintf(&sb, "- `%s`\n", escapeLabel(markdownTextEscaper, stateId))
	}
	sb.WriteString("\n")

	// Transitions section
	sb.WriteString("## Transitions\n\n")
	sb.WriteString("| Source State | Event | Target State | Type |\n")
	sb.WriteString("|-------------|-------|--------------|------|\n")

	sm.forEachTransition(func(transition *Transition[S, E, P]) {
		transType := "External"
		if transition.TransType == Internal {
			transType = "Internal"
		}
		fmt.Fprintf(&sb, "| `%s` | `%s` | `%s` | %s |\n",
			escapeLabel(markdownCellEscaper, transition.Source.id),
			escapeLabel(markdownCellEscaper, transition.Event),
			escapeLabel(markdownCellEscaper, transition.Target.id),
			transType)
	})

	return sb.String()
}

// generateMarkdownFlow returns a Mermaid flowchart diagram in Markdown format
func (sm *StateMachineImpl[S, E, P]) generateMarkdownFlow() string {
	states, aliases := sm.stateAliases("state_")

	var sb strings.Builder
	sb.WriteString("```mermaid\nflowchart TD\n")

	// Define nodes
	for _, stateId := range states {
		fmt.Fprintf(&sb, "    %s[\"%s\"]\n", aliases[stateId], escapeLabel(mermaidEscaper, stateId))
	}

	// Define transitions; quoted labels may contain the | that delimits them
	sm.forEachTransition(func(transition *Transition[S, E, P]) {
		fmt.Fprintf(&sb, "    %s -->|\"%s\"| %s\n",
			aliases[transition.Source.id], escapeLabel(mermaidEscaper, transition.Event), aliases[transition.Target.id])
	})

	sb.WriteString("```\n")
	return sb.String()
}

// generateMarkdownStateDiagram returns a Mermaid state diagram in Markdown format
func (sm *StateMachineImpl[S, E, P]) generateMarkdownStateDiagram() string {
	states, aliases := sm.stateAliases("s")

	var sb strings.Builder
	sb.WriteString("```mermaid\nstateDiagram-v2\n")

	// Declare states, since a state ID can't contain spaces
	for _, stateId := range states {
		fmt.Fprintf(&sb, "    state \"%s\" as %s\n", escapeLabel(mermaidEscaper, stateId), aliases[stateId])
	}

	// Define transitions
	sm.forEachTransition(func(transition *Transition[S, E, P]) {
		marker := ""
		if transition.TransType == Internal {
			marker = " [internal]"
		}
		fmt.Fprintf(&sb, "    %s --> %s : %s%s\n",
			aliases[transition.Source.id], aliases[transition.Target.id], escapeLabel(mermaidEscaper, transition.Event), marker)
	})

	sb.WriteString("```\n")
	return sb.String()
}
