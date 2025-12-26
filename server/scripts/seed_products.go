package main

import (
	"fmt"
	"log"
	"time"

	"ai-conversation-platform/internal/models"
	"ai-conversation-platform/internal/storage/postgres"
)

func main() {
	// Initialize database client
	dbClient, err := postgres.NewClient()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer dbClient.Close()

	productStorage := postgres.NewProductStorage(dbClient)

	// Default tenant ID (OMX26)
	tenantID := "OMX26"

	// Define the 5 demo products
	products := []*models.Product{
		{
			ID:              "prod-whatsapp-starter",
			TenantID:        tenantID,
			Name:            "WhatsApp Automation Starter",
			Description:     "A basic WhatsApp automation tool to auto-reply to customer messages and capture leads. Perfect for small businesses with 1-5 employees.",
			Category:        "Automation",
			Price:           999.0,
			PriceCurrency:   "INR",
			Features:        []string{"Automated greetings & FAQs", "Lead capture from WhatsApp", "Simple analytics dashboard"},
			Limitations:      []string{"No CRM integration", "Limited message templates"},
			TargetAudience:  "Small businesses (1–5 employees)",
			CommonQuestions: []string{"Is this enough for a small shop?", "Can I upgrade later?", "Why should I pay for this?"},
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			ID:              "prod-whatsapp-pro",
			TenantID:        tenantID,
			Name:            "WhatsApp Sales Pro",
			Description:     "Advanced WhatsApp sales automation with agent assignment and CRM sync. Designed for growing sales teams with 5-20 agents.",
			Category:        "Automation",
			Price:           3499.0,
			PriceCurrency:   "INR",
			Features:        []string{"Multi-agent inbox", "CRM integration", "Lead tagging & follow-ups", "AI-assisted replies (agent-only)"},
			Limitations:      []string{"Requires agent onboarding", "No custom AI training"},
			TargetAudience:  "Growing sales teams (5–20 agents)",
			CommonQuestions: []string{"Is this better than hiring more agents?", "Does it work with my CRM?", "Why is it expensive?"},
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			ID:              "prod-ai-lead-intelligence",
			TenantID:        tenantID,
			Name:            "AI Lead Intelligence Add-On",
			Description:     "An AI module that scores leads, predicts deal outcomes, and prioritizes follow-ups. Works seamlessly with supported channels.",
			Category:        "Intelligence",
			Price:           1999.0,
			PriceCurrency:   "INR",
			Features:        []string{"Lead scoring", "Win probability prediction", "Churn risk detection", "Sales insights dashboard"},
			Limitations:      []string{"Works only with supported channels", "AI suggestions require review"},
			TargetAudience:  "Sales managers",
			CommonQuestions: []string{"How accurate is the AI?", "Can AI really predict sales?", "Will this replace my team?"},
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			ID:              "prod-support-automation",
			TenantID:        tenantID,
			Name:            "Customer Support Automation",
			Description:     "AI-powered support automation for WhatsApp and web chat. Handles intent-based routing, sentiment detection, and escalation triggers.",
			Category:        "Support",
			Price:           2499.0,
			PriceCurrency:   "INR",
			Features:        []string{"Intent-based routing", "Sentiment detection", "Escalation triggers", "Agent assist insights"},
			Limitations:      []string{"Not designed for outbound sales", "No pricing automation"},
			TargetAudience:  "Support-heavy businesses",
			CommonQuestions: []string{"Can this handle angry customers?", "What if AI gives wrong answers?"},
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
		{
			ID:              "prod-enterprise-suite",
			TenantID:        tenantID,
			Name:            "Enterprise Automation Suite",
			Description:     "Full automation suite combining sales, support, and analytics. Includes all previous modules plus custom rules, advanced analytics, and admin-level controls.",
			Category:        "Enterprise",
			Price:           0.0, // Custom pricing
			PriceCurrency:   "INR",
			Features:        []string{"All previous modules", "Custom rules & policies", "Advanced analytics", "Admin-level controls"},
			Limitations:      []string{"Requires onboarding", "Higher setup time"},
			TargetAudience:  "Large teams / agencies",
			CommonQuestions: []string{"Why custom pricing?", "Is my data secure?", "Can we control AI behavior?"},
			CreatedAt:       time.Now(),
			UpdatedAt:       time.Now(),
		},
	}

	// Seed products
	for _, product := range products {
		// Check if product already exists
		_, err := productStorage.GetProduct(tenantID, product.ID)
		if err == nil {
			log.Printf("Product %s already exists, skipping...", product.ID)
			continue
		}

		// Create product
		if err := productStorage.CreateProduct(tenantID, product); err != nil {
			log.Printf("Failed to create product %s: %v", product.ID, err)
			continue
		}

		log.Printf("Successfully created product: %s", product.Name)
	}

	fmt.Println("Product seeding completed!")
}

