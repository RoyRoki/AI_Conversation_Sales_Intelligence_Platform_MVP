package rules

// PolicyType represents the type of policy rule
type PolicyType string

const (
	// NoFalseClaims prevents false or unsubstantiated claims
	NoFalseClaims PolicyType = "no_false_claims"
	// NoUnauthorizedDiscounts prevents unauthorized discount offers
	NoUnauthorizedDiscounts PolicyType = "no_unauthorized_discounts"
	// NoLegalPromises prevents legal or financial promises
	NoLegalPromises PolicyType = "no_legal_promises"
	// BrandToneCompliance ensures brand tone consistency
	BrandToneCompliance PolicyType = "brand_tone_compliance"
	// ObjectionConfirmation validates AI-detected objections against keyword patterns
	ObjectionConfirmation PolicyType = "objection_confirmation"
)

// Severity represents the severity level of a policy violation
type Severity string

const (
	SeverityLow    Severity = "low"
	SeverityMedium Severity = "medium"
	SeverityHigh   Severity = "high"
	SeverityCritical Severity = "critical"
)

// PolicyMetadata contains metadata about a policy
type PolicyMetadata struct {
	Type        PolicyType
	Name        string
	Description string
	Severity    Severity
	Priority    int // Lower number = higher priority
}

// DefaultPolicies returns default policy configurations
func DefaultPolicies() []PolicyMetadata {
	return []PolicyMetadata{
		{
			Type:        NoFalseClaims,
			Name:        "No False Claims",
			Description: "Prevents false or unsubstantiated claims about products or services",
			Severity:    SeverityHigh,
			Priority:    1,
		},
		{
			Type:        NoUnauthorizedDiscounts,
			Name:        "No Unauthorized Discounts",
			Description: "Prevents offering discounts without authorization",
			Severity:    SeverityCritical,
			Priority:    2,
		},
		{
			Type:        NoLegalPromises,
			Name:        "No Legal/Financial Promises",
			Description: "Prevents making legal or financial promises",
			Severity:    SeverityCritical,
			Priority:    3,
		},
		{
			Type:        BrandToneCompliance,
			Name:        "Brand Tone Compliance",
			Description: "Ensures responses match brand tone guidelines",
			Severity:    SeverityMedium,
			Priority:    4,
		},
		{
			Type:        ObjectionConfirmation,
			Name:        "Objection Detection Confirmation",
			Description: "Validates AI-detected objections against keyword patterns",
			Severity:    SeverityLow,
			Priority:    5,
		},
	}
}

// DefaultPatterns returns default regex/keyword patterns for each policy type
func DefaultPatterns() map[PolicyType][]string {
	return map[PolicyType][]string{
		NoFalseClaims: {
			`(?i)\b(guaranteed|100%|always|never fails|best ever)\b`,
			`(?i)\b(will definitely|absolutely|without doubt)\b`,
		},
		NoUnauthorizedDiscounts: {
			`(?i)\b(discount|% off|save \$\d+|special price|reduced price)\b`,
			`(?i)\b(50%|75%|90% off|free shipping|no cost)\b`,
		},
		NoLegalPromises: {
			`(?i)\b(guarantee|warranty|refund policy|money back)\b`,
			`(?i)\b(legal|lawsuit|contract|agreement|binding)\b`,
			`(?i)\b(we promise|we guarantee|we will pay)\b`,
		},
		BrandToneCompliance: {
			`(?i)\b(crazy|insane|ridiculous|stupid|dumb)\b`,
			`(?i)\b(swear words|profanity)\b`,
		},
		ObjectionConfirmation: {
			`(?i)\b(price|expensive|cost|cheaper|afford)\b`,
			`(?i)\b(trust|reliable|reputation|credible)\b`,
			`(?i)\b(delivery|shipping|time|wait)\b`,
			`(?i)\b(competitor|alternative|better option|other company)\b`,
		},
	}
}

// CorrectionTemplates returns predefined correction templates per rule type
func CorrectionTemplates() map[PolicyType]string {
	return map[PolicyType]string{
		NoFalseClaims: "I'd be happy to share more information about our product features and benefits.",
		NoUnauthorizedDiscounts: "Let me check with our team about current pricing and promotions.",
		NoLegalPromises: "I can provide information about our standard policies. For specific legal matters, please consult with our legal team.",
		BrandToneCompliance: "Let me rephrase that in a more professional manner.",
		ObjectionConfirmation: "", // Objection confirmation doesn't need correction, just validation
	}
}

// GetPolicyPriority returns the priority order for rule evaluation
// Lower number = higher priority (evaluated first)
func GetPolicyPriority(policyType PolicyType) int {
	policies := DefaultPolicies()
	for _, p := range policies {
		if p.Type == policyType {
			return p.Priority
		}
	}
	return 999 // Default low priority if not found
}


