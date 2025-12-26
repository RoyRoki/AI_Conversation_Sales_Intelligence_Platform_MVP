package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"ai-conversation-platform/internal/storage/postgres"
)

func main() {
	var direction string
	flag.StringVar(&direction, "direction", "up", "Migration direction: up or down")
	flag.Parse()

	client, err := postgres.NewClient()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to connect to database: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	if direction == "up" {
		if err := runMigrations(client.DB); err != nil {
			fmt.Fprintf(os.Stderr, "Migration failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Migrations completed successfully")
	} else {
		fmt.Println("Down migrations not implemented in MVP")
	}
}

func runMigrations(db *sql.DB) error {
	migrations := []string{
		createUsersTable,
		createConversationsTable,
		createMessagesTable,
		createConversationMetadataTable,
		createCustomerMemoryTable,
		createRulesTable,
		createBrandToneTable,
		createProductsTable,
		createAutoReplyGlobalTable,
		createAutoReplyConversationsTable,
		createSuggestionsTable,
	}

	for i, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("migration %d failed: %w", i+1, err)
		}
		fmt.Printf("Migration %d completed\n", i+1)
	}

	// Handle product_id column addition separately (SQLite compatibility)
	if err := addProductIdColumn(db); err != nil {
		return fmt.Errorf("failed to add product_id column: %w", err)
	}

	// Handle customer_id column addition separately (SQLite compatibility)
	if err := addCustomerIdColumn(db); err != nil {
		return fmt.Errorf("failed to add customer_id column: %w", err)
	}

	// Seed demo products
	if err := seedDemoProducts(db); err != nil {
		return fmt.Errorf("failed to seed products: %w", err)
	}

	return nil
}

// addProductIdColumn adds product_id column to conversations table
// Handles both SQLite and PostgreSQL by attempting to add and ignoring if already exists
func addProductIdColumn(db *sql.DB) error {
	// Try to add the column - ignore error if it already exists
	_, err := db.Exec("ALTER TABLE conversations ADD COLUMN product_id TEXT")
	if err != nil {
		// Check if error is because column already exists
		errStr := strings.ToLower(err.Error())
		if !contains(errStr, "duplicate column") && !contains(errStr, "already exists") && !contains(errStr, "duplicate column name") {
			// Real error, return it
			return err
		}
		// Column exists, that's fine - continue
	}
	// Create index (IF NOT EXISTS handles if index already exists)
	_, err = db.Exec("CREATE INDEX IF NOT EXISTS idx_conversations_product_id ON conversations(product_id)")
	return err
}

func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// addCustomerIdColumn adds customer_id column to conversations table
// Handles both SQLite and PostgreSQL by attempting to add and ignoring if already exists
func addCustomerIdColumn(db *sql.DB) error {
	// Try to add the column - ignore error if it already exists
	_, err := db.Exec("ALTER TABLE conversations ADD COLUMN customer_id TEXT")
	if err != nil {
		// Check if error is because column already exists
		errStr := strings.ToLower(err.Error())
		if !contains(errStr, "duplicate column") && !contains(errStr, "already exists") && !contains(errStr, "duplicate column name") {
			// Real error, return it
			return err
		}
		// Column exists, that's fine - continue
	}
	// Create index (IF NOT EXISTS handles if index already exists)
	_, err = db.Exec("CREATE INDEX IF NOT EXISTS idx_conversations_customer_id ON conversations(customer_id)")
	if err != nil {
		return err
	}
	// Create composite index for finding active conversations by customer
	_, err = db.Exec("CREATE INDEX IF NOT EXISTS idx_conversations_customer_status ON conversations(customer_id, status)")
	return err
}

// seedDemoProducts seeds the 5 demo products
func seedDemoProducts(db *sql.DB) error {
	tenantID := "OMX26"
	now := time.Now().Format("2006-01-02 15:04:05")

	products := []struct {
		id              string
		name            string
		description     string
		category        string
		price           float64
		features        string
		limitations     string
		targetAudience  string
		commonQuestions string
	}{
		{
			id:              "prod-whatsapp-starter",
			name:            "WhatsApp Automation Starter",
			description:     "A basic WhatsApp automation tool to auto-reply to customer messages and capture leads. Perfect for small businesses with 1-5 employees.",
			category:        "Automation",
			price:           999.0,
			features:        `["Automated greetings & FAQs", "Lead capture from WhatsApp", "Simple analytics dashboard"]`,
			limitations:     `["No CRM integration", "Limited message templates"]`,
			targetAudience:  "Small businesses (1–5 employees)",
			commonQuestions: `["Is this enough for a small shop?", "Can I upgrade later?", "Why should I pay for this?"]`,
		},
		{
			id:              "prod-whatsapp-pro",
			name:            "WhatsApp Sales Pro",
			description:     "Advanced WhatsApp sales automation with agent assignment and CRM sync. Designed for growing sales teams with 5-20 agents.",
			category:        "Automation",
			price:           3499.0,
			features:        `["Multi-agent inbox", "CRM integration", "Lead tagging & follow-ups", "AI-assisted replies (agent-only)"]`,
			limitations:     `["Requires agent onboarding", "No custom AI training"]`,
			targetAudience:  "Growing sales teams (5–20 agents)",
			commonQuestions: `["Is this better than hiring more agents?", "Does it work with my CRM?", "Why is it expensive?"]`,
		},
		{
			id:              "prod-ai-lead-intelligence",
			name:            "AI Lead Intelligence Add-On",
			description:     "An AI module that scores leads, predicts deal outcomes, and prioritizes follow-ups. Works seamlessly with supported channels.",
			category:        "Intelligence",
			price:           1999.0,
			features:        `["Lead scoring", "Win probability prediction", "Churn risk detection", "Sales insights dashboard"]`,
			limitations:     `["Works only with supported channels", "AI suggestions require review"]`,
			targetAudience:  "Sales managers",
			commonQuestions: `["How accurate is the AI?", "Can AI really predict sales?", "Will this replace my team?"]`,
		},
		{
			id:              "prod-support-automation",
			name:            "Customer Support Automation",
			description:     "AI-powered support automation for WhatsApp and web chat. Handles intent-based routing, sentiment detection, and escalation triggers.",
			category:        "Support",
			price:           2499.0,
			features:        `["Intent-based routing", "Sentiment detection", "Escalation triggers", "Agent assist insights"]`,
			limitations:     `["Not designed for outbound sales", "No pricing automation"]`,
			targetAudience:  "Support-heavy businesses",
			commonQuestions: `["Can this handle angry customers?", "What if AI gives wrong answers?"]`,
		},
		{
			id:              "prod-enterprise-suite",
			name:            "Enterprise Automation Suite",
			description:     "Full automation suite combining sales, support, and analytics. Includes all previous modules plus custom rules, advanced analytics, and admin-level controls.",
			category:        "Enterprise",
			price:           0.0,
			features:        `["All previous modules", "Custom rules & policies", "Advanced analytics", "Admin-level controls"]`,
			limitations:     `["Requires onboarding", "Higher setup time"]`,
			targetAudience:  "Large teams / agencies",
			commonQuestions: `["Why custom pricing?", "Is my data secure?", "Can we control AI behavior?"]`,
		},
	}

	for _, p := range products {
		// Check if product exists
		var count int
		err := db.QueryRow("SELECT COUNT(*) FROM products WHERE id = $1 AND tenant_id = $2", p.id, tenantID).Scan(&count)
		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("failed to check product %s: %w", p.id, err)
		}

		if count > 0 {
			continue // Product already exists
		}

		// Insert product
		query := `
			INSERT INTO products (id, tenant_id, name, description, category, price, price_currency, features, limitations, target_audience, common_questions, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		`
		_, err = db.Exec(query, p.id, tenantID, p.name, p.description, p.category, p.price, "INR", p.features, p.limitations, p.targetAudience, p.commonQuestions, now, now)
		if err != nil {
			return fmt.Errorf("failed to insert product %s: %w", p.id, err)
		}
	}

	return nil
}

const createUsersTable = `
CREATE TABLE IF NOT EXISTS users (
	id TEXT PRIMARY KEY,
	tenant_id TEXT NOT NULL,
	email TEXT UNIQUE NOT NULL,
	password_hash TEXT NOT NULL,
	role TEXT NOT NULL CHECK(role IN ('customer', 'agent', 'admin')),
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_users_tenant_id ON users(tenant_id);
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
`

const createConversationsTable = `
CREATE TABLE IF NOT EXISTS conversations (
	id TEXT PRIMARY KEY,
	tenant_id TEXT NOT NULL,
	status TEXT NOT NULL DEFAULT 'active' CHECK(status IN ('active', 'closed', 'archived')),
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_conversations_tenant_id ON conversations(tenant_id);
CREATE INDEX IF NOT EXISTS idx_conversations_status ON conversations(status);
`

const createMessagesTable = `
CREATE TABLE IF NOT EXISTS messages (
	id TEXT PRIMARY KEY,
	conversation_id TEXT NOT NULL,
	sender TEXT NOT NULL CHECK(sender IN ('customer', 'agent')),
	content TEXT NOT NULL,
	channel TEXT NOT NULL DEFAULT 'web',
	language TEXT,
	timestamp TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_messages_conversation_id ON messages(conversation_id);
CREATE INDEX IF NOT EXISTS idx_messages_timestamp ON messages(timestamp);
`

const createConversationMetadataTable = `
CREATE TABLE IF NOT EXISTS conversation_metadata (
	id TEXT PRIMARY KEY,
	conversation_id TEXT NOT NULL UNIQUE,
	intent TEXT,
	intent_score REAL,
	sentiment TEXT,
	sentiment_score REAL,
	emotions TEXT, -- JSON array stored as text
	objections TEXT, -- JSON array stored as text
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_metadata_conversation_id ON conversation_metadata(conversation_id);
`

const createCustomerMemoryTable = `
CREATE TABLE IF NOT EXISTS customer_memory (
	id TEXT PRIMARY KEY,
	tenant_id TEXT NOT NULL,
	customer_id TEXT NOT NULL,
	preferred_language TEXT,
	pricing_sensitivity TEXT,
	product_interests TEXT, -- JSON array stored as text
	past_objections TEXT, -- JSON array stored as text
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_memory_tenant_id ON customer_memory(tenant_id);
CREATE INDEX IF NOT EXISTS idx_memory_customer_id ON customer_memory(customer_id);
`

const createRulesTable = `
CREATE TABLE IF NOT EXISTS rules (
	id TEXT PRIMARY KEY,
	tenant_id TEXT NOT NULL,
	name TEXT NOT NULL,
	description TEXT,
	type TEXT NOT NULL,
	pattern TEXT NOT NULL,
	action TEXT NOT NULL CHECK(action IN ('block', 'auto_correct', 'flag')),
	is_active BOOLEAN NOT NULL DEFAULT true,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_rules_tenant_id ON rules(tenant_id);
CREATE INDEX IF NOT EXISTS idx_rules_is_active ON rules(is_active);
`

const createBrandToneTable = `
CREATE TABLE IF NOT EXISTS brand_tone (
	tenant_id TEXT PRIMARY KEY,
	tone TEXT NOT NULL CHECK(tone IN ('Professional', 'Friendly', 'Sales-focused')),
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_brand_tone_tenant_id ON brand_tone(tenant_id);
`

const createProductsTable = `
CREATE TABLE IF NOT EXISTS products (
	id TEXT PRIMARY KEY,
	tenant_id TEXT NOT NULL,
	name TEXT NOT NULL,
	description TEXT NOT NULL,
	category TEXT,
	price REAL NOT NULL,
	price_currency TEXT NOT NULL DEFAULT 'INR',
	features TEXT, -- JSON array stored as text
	limitations TEXT, -- JSON array stored as text
	target_audience TEXT,
	common_questions TEXT, -- JSON array stored as text
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_products_tenant_id ON products(tenant_id);
CREATE INDEX IF NOT EXISTS idx_products_category ON products(category);
`

const createAutoReplyGlobalTable = `
CREATE TABLE IF NOT EXISTS auto_reply_global (
	tenant_id TEXT PRIMARY KEY,
	enabled BOOLEAN NOT NULL DEFAULT false,
	confidence_threshold REAL NOT NULL DEFAULT 0.8,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_autoreply_global_tenant_id ON auto_reply_global(tenant_id);
`

const createAutoReplyConversationsTable = `
CREATE TABLE IF NOT EXISTS auto_reply_conversations (
	conversation_id TEXT PRIMARY KEY,
	enabled BOOLEAN NOT NULL DEFAULT false,
	confidence_threshold REAL,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_autoreply_conversations_id ON auto_reply_conversations(conversation_id);
`

const createSuggestionsTable = `
CREATE TABLE IF NOT EXISTS suggestions (
	id TEXT PRIMARY KEY,
	conversation_id TEXT NOT NULL,
	last_customer_message_id TEXT NOT NULL,
	suggestions_data TEXT NOT NULL,
	context_used BOOLEAN NOT NULL DEFAULT false,
	created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
	FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_suggestions_conversation_message ON suggestions(conversation_id, last_customer_message_id);
CREATE INDEX IF NOT EXISTS idx_suggestions_conversation_id ON suggestions(conversation_id);
`

