Purpose

This file defines non-negotiable coding standards for all AI agents and contributors working on this repository.

The goal is boring, predictable, testable software:

Clean Architecture

SOLID (especially SRP)

Readable over clever

Reusable without abstraction hell

IoC only where it pays real dividends

If something here feels “restrictive”, that’s intentional.

Core Principles (Read This Twice)
1. Single Responsibility Is LAW

One package, one reason to change.

If a file:

talks to DB and

calls AI and

applies rules

→ it is wrong. Split it.

2. Clarity Beats Cleverness

No smart tricks.
No magical helpers.
No hidden side effects.

If a junior dev can’t understand the file in 5 minutes, rewrite it.

3. Explicit > Implicit

Prefer:

explicit parameters

explicit return values

explicit errors

Avoid:

global state

hidden context passing

magic defaults

4. Composition Over Inheritance

Interfaces are for boundaries, not everything.

If an interface has only one implementation, question its existence.

Dependency Injection (Controlled IoC)
Allowed

Constructor injection at service boundaries

Interfaces only for:

external systems (DB, AI, cache)

cross-package boundaries

Forbidden

DI frameworks

Passing interfaces through 5 layers “just in case”

Mock-driven architecture for MVP

Rule of thumb:

If mocking it doesn’t reduce test complexity, don’t abstract it.

Package Structure Rules (Go)
internal/
├── api/            // HTTP layer only (no business logic)
├── services/       // Business use-cases
├── models/         // Pure data structures
├── storage/        // DB & external storage
├── rules/          // Policy & validation logic
└── ai/             // AI orchestration only

API Layer

Translates HTTP → domain

No business rules

No DB queries

No AI calls

If you see logic here, it’s a bug.

Service Layer

Orchestrates use-cases

Calls storage, AI, rules

No HTTP, no SQL strings

One service = one use-case.

Storage Layer

CRUD only

No business decisions

No conditional logic beyond queries

File Rules
Max File Size

300 lines max

If bigger → split

Function Rules

One function = one idea

40 lines → refactor

4 parameters → rethink design

Naming Rules (Strict)
Packages

lowercase

single responsibility

no utils, helpers, common

❌ utils
✅ conversationstore, confidencecalc

Functions

Verbs for actions

Nouns for data

❌ HandleStuff()
✅ AnalyzeConversation()

Boolean Names

Must answer a yes/no question.

❌ valid
✅ isValid, hasPermission, shouldEmbed

Error Handling

No panics in business logic

Errors must be:

wrapped with context

actionable

return fmt.Errorf("fetch conversation %s: %w", id, err)


If an error message doesn’t help debugging, it’s useless.

AI-Specific Rules
AI Is NOT Trusted

LLMs:

hallucinate

lie confidently

contradict themselves

Therefore:

AI outputs must pass rule engine

Confidence score must be computed outside the LLM

AI failure must never break UX

If confidence < threshold → fallback to human flow.

Context Handling

PostgreSQL = source of truth

Vector DB = retrieval only

Never embed raw chat blindly

Summaries > raw messages

If you embed everything, you don’t understand embeddings.

Testing Philosophy
Required

Unit tests for:

rules

scoring

pure logic

Optional (MVP)

Heavy mocking

Full end-to-end AI tests

Test behavior, not implementation.

What NOT To Do (Hard NOs)

God services

context.Context abuse

Over-abstracted interfaces

“Future proof” design without evidence

Framework-driven architecture

If you’re building something “just in case”, stop.

Code Review Checklist (Self-Check)

Before committing, ask:

Does this file have one reason to change?

Can I explain this to a fresher in 3 minutes?

Did I add an interface without a real need?

Is PostgreSQL still the source of truth?

Would removing this abstraction simplify things?

If any answer is “no” → fix it.

Final Warning

This repo values:

correctness over speed

clarity over cleverness

simplicity over hype

If you want to experiment, do it outside the core path.

Production code is not a playground.