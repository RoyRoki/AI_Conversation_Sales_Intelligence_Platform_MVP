package rules

import (
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"

	"ai-conversation-platform/internal/models"
)

// Violation represents a rule violation
type Violation struct {
	RuleID      string
	RuleName    string
	RuleType    string
	Action      string
	Pattern     string
	MatchedText string
	Severity    string
}

// ValidationResult represents the result of rule validation
type ValidationResult struct {
	Passed        bool
	Violations    []Violation
	CorrectedText string
	Blocked       bool
	Explanation   string
	RuleResults   []bool // Boolean results per rule for confidence scoring
}

// RuleEngine handles rule validation
type RuleEngine struct {
	correctionTemplates map[PolicyType]string
}

// NewRuleEngine creates a new rule engine
func NewRuleEngine() *RuleEngine {
	return &RuleEngine{
		correctionTemplates: CorrectionTemplates(),
	}
}

// ValidateOutput validates text output against rules
func (e *RuleEngine) ValidateOutput(input string, rules []*models.Rule) ValidationResult {
	result := ValidationResult{
		Passed:        true,
		Violations:    []Violation{},
		CorrectedText: input,
		Blocked:       false,
		Explanation:   "",
		RuleResults:   make([]bool, len(rules)),
	}

	// Filter active rules only
	activeRules := e.filterActiveRules(rules)
	if len(activeRules) == 0 {
		return result
	}

	// Sort rules by priority (block > correct > flag)
	sortedRules := e.sortRulesByPriority(activeRules)

	// Evaluate each rule
	violations := []Violation{}
	correctedText := input

	for i, rule := range sortedRules {
		matched, matchedText := e.matchPattern(correctedText, rule.Pattern)
		result.RuleResults[i] = !matched // true if rule passed (no match)

		if matched {
			violation := Violation{
				RuleID:      rule.ID,
				RuleName:    rule.Name,
				RuleType:    rule.Type,
				Action:      rule.Action,
				Pattern:     rule.Pattern,
				MatchedText: matchedText,
				Severity:    e.getSeverityForRule(rule),
			}
			violations = append(violations, violation)

			// Log rule trigger for audit
			log.Printf("[RULE] violation detected rule_id=%s rule_name=%s action=%s severity=%s matched=%s",
				violation.RuleID, violation.RuleName, violation.Action, violation.Severity, violation.MatchedText)

			// Apply action based on rule
			switch rule.Action {
			case "block":
				result.Blocked = true
				result.Passed = false
				result.Explanation = e.GenerateExplanation(violations)
				result.Violations = violations
				log.Printf("[RULE] response blocked rule_id=%s violations=%d", violation.RuleID, len(violations))
				return result
			case "auto_correct":
				correctedText = e.AutoCorrect(correctedText, violation)
				log.Printf("[RULE] auto-correction applied rule_id=%s", violation.RuleID)
			case "flag":
				// Flagged but continue evaluation
				log.Printf("[RULE] response flagged rule_id=%s", violation.RuleID)
			}
		}
	}

	// Set final result
	result.Violations = violations
	result.CorrectedText = correctedText
	result.Passed = len(violations) == 0 || !result.Blocked

	// If there were violations but not blocked, generate explanation
	if len(violations) > 0 && !result.Blocked {
		result.Explanation = fmt.Sprintf("Response was auto-corrected due to %d policy violation(s)", len(violations))
	}

	return result
}

// filterActiveRules returns only active rules
func (e *RuleEngine) filterActiveRules(rules []*models.Rule) []*models.Rule {
	active := []*models.Rule{}
	for _, rule := range rules {
		if rule.IsActive {
			active = append(active, rule)
		}
	}
	return active
}

// sortRulesByPriority sorts rules by action priority (block > correct > flag)
func (e *RuleEngine) sortRulesByPriority(rules []*models.Rule) []*models.Rule {
	sorted := make([]*models.Rule, len(rules))
	copy(sorted, rules)

	sort.Slice(sorted, func(i, j int) bool {
		priorityI := e.getActionPriority(sorted[i].Action)
		priorityJ := e.getActionPriority(sorted[j].Action)
		return priorityI < priorityJ // Lower number = higher priority
	})

	return sorted
}

// getActionPriority returns priority number for action type
func (e *RuleEngine) getActionPriority(action string) int {
	switch action {
	case "block":
		return 1
	case "auto_correct":
		return 2
	case "flag":
		return 3
	default:
		return 999
	}
}

// matchPattern matches text against a pattern (regex or keyword)
func (e *RuleEngine) matchPattern(text, pattern string) (bool, string) {
	// Try regex first
	regex, err := regexp.Compile(pattern)
	if err == nil {
		matches := regex.FindStringSubmatch(text)
		if len(matches) > 0 {
			return true, matches[0]
		}
		return false, ""
	}

	// Fallback to simple keyword matching (case-insensitive)
	lowerText := strings.ToLower(text)
	lowerPattern := strings.ToLower(pattern)
	if strings.Contains(lowerText, lowerPattern) {
		return true, pattern
	}

	return false, ""
}

// AutoCorrect applies correction template to text based on violation
func (e *RuleEngine) AutoCorrect(text string, violation Violation) string {
	// Determine policy type from rule name/type
	policyType := e.getPolicyTypeFromRule(violation.RuleType, violation.RuleName)
	template, exists := e.correctionTemplates[policyType]

	if !exists || template == "" {
		// Generic correction: remove the matched text
		return strings.ReplaceAll(text, violation.MatchedText, "")
	}

	// Replace matched text with correction template
	corrected := strings.ReplaceAll(text, violation.MatchedText, template)

	// If replacement didn't work (matched text not found), append template
	if corrected == text {
		corrected = text + " " + template
	}

	return corrected
}

// ShouldBlock determines if violations should result in blocking
func (e *RuleEngine) ShouldBlock(violations []Violation) bool {
	for _, v := range violations {
		if v.Action == "block" {
			return true
		}
	}
	return false
}

// GenerateExplanation generates human-readable explanation for blocked responses
func (e *RuleEngine) GenerateExplanation(violations []Violation) string {
	if len(violations) == 0 {
		return ""
	}

	if len(violations) == 1 {
		v := violations[0]
		return fmt.Sprintf("Response blocked due to policy violation: %s (matched: %s)", v.RuleName, v.MatchedText)
	}

	ruleNames := []string{}
	for _, v := range violations {
		ruleNames = append(ruleNames, v.RuleName)
	}

	return fmt.Sprintf("Response blocked due to %d policy violations: %s", len(violations), strings.Join(ruleNames, ", "))
}

// getSeverityForRule determines severity based on rule action and type
func (e *RuleEngine) getSeverityForRule(rule *models.Rule) string {
	switch rule.Action {
	case "block":
		return "critical"
	case "auto_correct":
		return "high"
	case "flag":
		return "medium"
	default:
		return "low"
	}
}

// getPolicyTypeFromRule maps rule type/name to PolicyType
func (e *RuleEngine) getPolicyTypeFromRule(ruleType, ruleName string) PolicyType {
	lowerType := strings.ToLower(ruleType)
	lowerName := strings.ToLower(ruleName)

	// Check by type first
	switch lowerType {
	case "no_false_claims", "false_claims":
		return NoFalseClaims
	case "no_unauthorized_discounts", "unauthorized_discounts":
		return NoUnauthorizedDiscounts
	case "no_legal_promises", "legal_promises":
		return NoLegalPromises
	case "brand_tone_compliance", "brand_tone":
		return BrandToneCompliance
	case "objection_confirmation", "objection":
		return ObjectionConfirmation
	}

	// Check by name keywords
	if strings.Contains(lowerName, "false") || strings.Contains(lowerName, "claim") {
		return NoFalseClaims
	}
	if strings.Contains(lowerName, "discount") {
		return NoUnauthorizedDiscounts
	}
	if strings.Contains(lowerName, "legal") || strings.Contains(lowerName, "promise") {
		return NoLegalPromises
	}
	if strings.Contains(lowerName, "tone") || strings.Contains(lowerName, "brand") {
		return BrandToneCompliance
	}
	if strings.Contains(lowerName, "objection") {
		return ObjectionConfirmation
	}

	return NoFalseClaims // Default
}

// ValidateObjections validates AI-detected objections against keyword patterns
func (e *RuleEngine) ValidateObjections(detectedObjections []string, conversationText string) []string {
	validated := []string{}
	patterns := DefaultPatterns()
	objectionPatterns := patterns[ObjectionConfirmation]

	// Build keyword map for objection types
	objectionMap := map[string][]string{
		"price":      {"price", "expensive", "cost", "cheaper", "afford"},
		"trust":      {"trust", "reliable", "reputation", "credible"},
		"delivery":   {"delivery", "shipping", "time", "wait"},
		"competitor": {"competitor", "alternative", "better option", "other company"},
	}

	lowerText := strings.ToLower(conversationText)

	// Validate each detected objection
	for _, obj := range detectedObjections {
		objLower := strings.ToLower(obj)
		keywords, exists := objectionMap[objLower]

		if !exists {
			// Check if objection matches any pattern
			matched := false
			for _, pattern := range objectionPatterns {
				if strings.Contains(lowerText, strings.ToLower(pattern)) {
					matched = true
					break
				}
			}
			if matched {
				validated = append(validated, obj)
			}
			continue
		}

		// Check if any keyword for this objection appears in text
		for _, keyword := range keywords {
			if strings.Contains(lowerText, keyword) {
				validated = append(validated, obj)
				break
			}
		}
	}

	return validated
}

