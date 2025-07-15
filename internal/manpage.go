package internal

import (
	"bufio"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
)

// WidgetDescription represents a widget's description from the manual
type WidgetDescription struct {
	WidgetName       string
	ShortDescription string
	FullDescription  string
}

// getWidgetDescriptions extracts widget descriptions from man zshzle
func getWidgetDescriptions() (map[string]WidgetDescription, error) {
	// Use "man zshzle | col -b" to get clean text without formatting
	cmd := exec.Command("sh", "-c", "man zshzle | col -b")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute man zshzle command: %w", err)
	}

	return ParseManPageDescriptions(string(output))
}

// parseManPageDescriptions parses the man zshzle output for widget descriptions
func ParseManPageDescriptions(content string) (map[string]WidgetDescription, error) {
	descriptions := make(map[string]WidgetDescription)
	scanner := bufio.NewScanner(strings.NewReader(content))

	// Regex to match widget entries: "       widget-name (keys) (keys) (keys)"
	// Widget headers are indented with spaces
	widgetHeaderRegex := regexp.MustCompile(`^\s+([a-zA-Z0-9_-]+)(\s*\([^)]*\))+\s*$`)

	var currentWidget string
	var descriptionLines []string
	var hasContent bool // Track if we have non-empty content in current paragraph

	for scanner.Scan() {
		line := scanner.Text()
		trimmedLine := strings.TrimSpace(line)

		// Check if this line is a widget header
		if matches := widgetHeaderRegex.FindStringSubmatch(line); len(matches) > 1 {
			// Save previous widget if we have one
			if currentWidget != "" && len(descriptionLines) > 0 {
				fullDescription := joinDescriptionLines(descriptionLines)
				shortDescription := extractFirstSentence(fullDescription)
				if shortDescription != "" {
					descriptions[currentWidget] = WidgetDescription{
						WidgetName:       currentWidget,
						ShortDescription: shortDescription,
						FullDescription:  strings.TrimSpace(fullDescription),
					}
				}
			}

			// Start new widget
			currentWidget = matches[1]
			descriptionLines = nil
			hasContent = false
			continue
		}

		// If we're tracking a widget and this looks like a description line
		if currentWidget != "" {

			// Handle empty lines - preserve paragraph breaks
			if trimmedLine == "" {
				// Add paragraph break if we have content
				if hasContent {
					descriptionLines = append(descriptionLines, "")
					hasContent = false
				}
				continue
			}

			// Stop if we hit another widget header or section
			if isNewSection(line) || isAnotherWidget(line) {
				// Save current widget
				if len(descriptionLines) > 0 {
					fullDescription := joinDescriptionLines(descriptionLines)
					shortDescription := extractFirstSentence(fullDescription)
					if shortDescription != "" {
						descriptions[currentWidget] = WidgetDescription{
							WidgetName:       currentWidget,
							ShortDescription: shortDescription,
							FullDescription:  strings.TrimSpace(fullDescription),
						}
					}
				}
				currentWidget = ""
				descriptionLines = nil
				hasContent = false
				continue
			}

			// Add this line to the description if it looks like description text
			if isDescriptionLine(line) {
				descriptionLines = append(descriptionLines, trimmedLine)
				hasContent = true
			}
		}
	}

	// Handle the last widget if we were processing one
	if currentWidget != "" && len(descriptionLines) > 0 {
		fullDescription := joinDescriptionLines(descriptionLines)
		shortDescription := extractFirstSentence(fullDescription)
		if shortDescription != "" {
			descriptions[currentWidget] = WidgetDescription{
				WidgetName:       currentWidget,
				ShortDescription: shortDescription,
				FullDescription:  strings.TrimSpace(fullDescription),
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading man page content: %w", err)
	}

	return descriptions, nil
}

// isNewSection returns true if the line indicates a new section starting
func isNewSection(line string) bool {
	// Check for section headers (usually all caps or specific patterns)
	if strings.HasPrefix(line, "ZSHZLE(1)") ||
		strings.HasPrefix(line, "NAME") ||
		strings.HasPrefix(line, "SYNOPSIS") ||
		strings.HasPrefix(line, "DESCRIPTION") ||
		strings.HasPrefix(line, "OPTIONS") ||
		strings.HasPrefix(line, "BUILTIN WIDGETS") ||
		strings.HasPrefix(line, "USER-DEFINED WIDGETS") ||
		strings.HasPrefix(line, "SPECIAL WIDGETS") ||
		strings.HasPrefix(line, "STANDARD WIDGETS") ||
		strings.HasPrefix(line, "Text Objects") ||
		strings.HasPrefix(line, "SEE ALSO") ||
		strings.HasPrefix(line, "AUTHOR") {
		return true
	}

	// Check for subsection headers (indented and title-case)
	if strings.HasPrefix(line, "   ") &&
		(strings.Contains(line, "Movement") ||
		 strings.Contains(line, "History") ||
		 strings.Contains(line, "Modifying") ||
		 strings.Contains(line, "Arguments") ||
		 strings.Contains(line, "Completion") ||
		 strings.Contains(line, "Miscellaneous")) {
		return true
	}

	return false
}

// isAnotherWidget returns true if the line looks like another widget definition
func isAnotherWidget(line string) bool {
	trimmed := strings.TrimSpace(line)

	// Check if line starts with whitespace and contains a widget-like name
	if strings.HasPrefix(line, "       ") && len(trimmed) > 0 {
		// Widget names are typically lowercase with hyphens and contain no spaces
		if matched, _ := regexp.MatchString(`^[a-z][a-z0-9-]*[a-z0-9]$`, trimmed); matched {
			return true
		}
	}

	return false
}

// isDescriptionLine returns true if the line looks like part of a widget description
func isDescriptionLine(line string) bool {
	// Skip lines that are obviously not descriptions
	if strings.HasPrefix(line, "ZSHZLE") ||
		strings.HasPrefix(line, "zsh:") ||
		strings.HasPrefix(line, "Page ") ||
		len(line) < 10 { // Very short lines are probably not descriptions
		return false
	}

	// Description lines usually start with whitespace and contain real text
	return strings.HasPrefix(line, " ") || strings.HasPrefix(line, "\t")
}

// joinDescriptionLines joins description lines preserving paragraph breaks
func joinDescriptionLines(lines []string) string {
	var result strings.Builder
	var currentParagraph strings.Builder
	
	for _, line := range lines {
		if line == "" {
			// Empty line indicates paragraph break
			if currentParagraph.Len() > 0 {
				result.WriteString(strings.TrimSpace(currentParagraph.String()))
				result.WriteString("\n\n")
				currentParagraph.Reset()
			}
		} else {
			// Add line to current paragraph
			if currentParagraph.Len() > 0 {
				currentParagraph.WriteString(" ")
			}
			currentParagraph.WriteString(line)
		}
	}
	
	// Add final paragraph if exists
	if currentParagraph.Len() > 0 {
		result.WriteString(strings.TrimSpace(currentParagraph.String()))
	}
	
	return strings.TrimSpace(result.String())
}

// extractFirstSentence extracts the first sentence from a description
func extractFirstSentence(text string) string {
	if text == "" {
		return ""
	}

	// Clean up the text
	text = strings.TrimSpace(text)
	text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")

	// Find the first sentence ending - look for punctuation followed by space or end of string
	sentenceEndRegex := regexp.MustCompile(`([.!?])(\s|$)`)
	match := sentenceEndRegex.FindStringIndex(text)

	if match != nil {
		// Found a sentence ending, return up to and including the punctuation
		endIndex := match[0] + 1 // Include the punctuation mark
		firstSentence := strings.TrimSpace(text[:endIndex])
		return firstSentence
	}

	// If no sentence break found, return the whole text (truncated if too long)
	if len(text) > 100 {
		text = text[:97] + "..."
	} else {
		// Add a period if it doesn't end with punctuation
		if text != "" && !strings.HasSuffix(text, ".") &&
		   !strings.HasSuffix(text, "!") && !strings.HasSuffix(text, "?") {
			text += "."
		}
	}
	return text
}

// getWidgetDescription gets a description for a specific widget
func getWidgetDescription(widgetName string, descriptions map[string]WidgetDescription) string {
	if desc, exists := descriptions[widgetName]; exists {
		return desc.ShortDescription
	}

	// Fallback: return the widget name itself
	return widgetName
}

// getWidgetFullDescription gets the full description for a specific widget
func getWidgetFullDescription(widgetName string, descriptions map[string]WidgetDescription) string {
	if desc, exists := descriptions[widgetName]; exists {
		return desc.FullDescription
	}

	// Fallback: return the widget name itself
	return widgetName
}

// GetWidgetDescriptionsTest is a test function to check our parsing
func GetWidgetDescriptionsTest() (map[string]WidgetDescription, error) {
	return getWidgetDescriptions()
}
