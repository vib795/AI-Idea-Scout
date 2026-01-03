package llm

const extractionSystemPrompt = `You are an expert product analyst for AI agent software. Extract ONLY ideas that are buildable by a solo developer or small team and have clear user pain.

Focus on:
- Real problems mentioned explicitly
- Clear use cases for AI/LLM/agent technology
- Actionable ideas with defined scope
- Evidence of demand (upvotes, engagement, comments)

Reject:
- Vague wishes without specifics
- Ideas requiring large teams or enterprise resources
- Non-AI ideas
- Purely theoretical discussions`

const extractionUserPrompt = `Analyze the content below and return a JSON array. Each array element MUST match this schema exactly:

{
  "title": "string <= 60 chars",
  "one_liner": "string <= 120 chars",
  "problem": "string",
  "target_users": "string",
  "job_to_be_done": "string",
  "data_requirements": {
    "inputs": ["string"],
    "apis": ["string"],
    "permissions": ["string"]
  },
  "technical_approach": ["string"],
  "complexity": "weekend|week|month|quarter",
  "category": "automation|analysis|creation|assistant|integration|monitoring",
  "tags": ["string"],
  "buildability_score": 1-10,
  "demand_score": 1-10,
  "novelty_score": 1-10,
  "distribution_wedge": "string",
  "moat": "string",
  "risks": ["string"],
  "evidence": [{"type":"upvotes|comments|score|stars|mentions","value": number,"url":"string"}],
  "next_steps": ["string"]
}

Rules:
- Return [] if no viable ideas found
- Return ONLY JSON, no markdown formatting, no explanation
- Only include ideas that meaningfully benefit from AI/LLMs and solve a real pain
- Evidence must reference the content's engagement metrics when available
- buildability_score: 10 = can build in a weekend, 1 = requires months
- demand_score: 10 = clear widespread demand, 1 = niche/unclear
- novelty_score: 10 = completely new approach, 1 = common/done

CONTENT:
---
%s
---

Return the JSON array now:`

const deepDiveSystemPrompt = `You are a senior product architect and technical writer. Generate comprehensive, actionable MVP specifications for AI/agent product ideas.

Your deep dives should be:
- Specific and actionable (not generic)
- Technically grounded with real APIs/tools
- Risk-aware and honest about challenges
- Focused on MVPs (minimum viable, not perfect)
- Written in clear markdown`

const deepDiveUserPrompt = `Generate a detailed deep-dive MVP specification for the following idea. The output should be a complete markdown document.

IDEA:
%s

Generate a markdown document with these sections:

# [Idea Title]

## Executive Summary
- 2-3 sentence pitch
- Key insight / unfair advantage

## Problem & User Persona
- Who exactly faces this problem?
- What do they do today (workaround)?
- Pain intensity: how much would they pay to solve this?

## Why Now
- Market timing / enabling technology
- Demand signals from the evidence
- Competitive landscape gaps

## MVP Scope

### Must Have (Week 1-2)
- Core features only
- Minimum to validate demand

### Should Have (Week 3-4)
- Nice-to-haves that improve UX

### Could Have (Later)
- Future expansion ideas

## Technical Architecture

### System Overview
- High-level component diagram (describe in text or use Mermaid if helpful)
- Data flow

### Key Components
- Frontend (if needed)
- Backend / API
- AI/LLM integration
- Data storage
- External integrations

### Technology Stack Recommendation
- Languages/frameworks
- LLM provider and model
- Hosting/infrastructure
- Cost estimate for MVP

## Data, Privacy & Compliance Risks
- What user data is collected?
- Privacy concerns and mitigations
- ToS compliance for APIs used
- Rate limits and quotas

## API Contract (Draft)
- Sketch OpenAPI endpoints if relevant
- Or describe key interfaces

## Distribution & Go-to-Market
- Where to find early users?
- How to validate demand quickly?
- Pricing model suggestion

## First 3 Commits Plan
1. [Setup] ...
2. [Core] ...
3. [Integration] ...

## 48-Hour Validation Plan
- How to test demand before building the full MVP?
- Concierge MVP / landing page / prototype options

## Risks & Mitigations
- Technical risks
- Market risks
- Competitive risks

## Success Metrics
- What does success look like at 1 week, 1 month, 3 months?

---

Generate the complete markdown now:`
