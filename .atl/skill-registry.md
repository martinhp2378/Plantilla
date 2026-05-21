# Skill Registry

**Delegator use only.** Any agent that launches sub-agents reads this registry to resolve compact rules, then injects them directly into sub-agent prompts. Sub-agents do NOT read this registry or individual SKILL.md files.

## User Skills

| Trigger | Skill | Path |
| :--- | :--- | :--- |
| implementation, commit splitting, chained PRs, or keeping tests and docs with code | work-unit-commits | /Users/jesusrosendomercadoporcayo/.gemini/skills/work-unit-commits/SKILL.md |
| Go tests, go test coverage, Bubbletea teatest, golden files | go-testing | /Users/jesusrosendomercadoporcayo/.gemini/skills/go-testing/SKILL.md |
| new skills, agent instructions, documenting AI usage patterns | skill-creator | /Users/jesusrosendomercadoporcayo/.gemini/skills/skill-creator/SKILL.md |
| judgment day, dual review, adversarial review, juzgar | judgment-day | /Users/jesusrosendomercadoporcayo/.gemini/skills/judgment-day/SKILL.md |
| creating GitHub issues, bug reports, or feature requests | issue-creation | /Users/jesusrosendomercadoporcayo/.gemini/skills/issue-creation/SKILL.md |
| PR feedback, issue replies, reviews, Slack messages, or GitHub comments | comment-writer | /Users/jesusrosendomercadoporcayo/.gemini/skills/comment-writer/SKILL.md |
| writing guides, READMEs, RFCs, onboarding, architecture, or review-facing docs | cognitive-doc-design | /Users/jesusrosendomercadoporcayo/.gemini/skills/cognitive-doc-design/SKILL.md |
| PRs over 400 lines, stacked PRs, review slices | chained-pr | /Users/jesusrosendomercadoporcayo/.gemini/skills/chained-pr/SKILL.md |
| creating, opening, or preparing PRs for review | branch-pr | /Users/jesusrosendomercadoporcayo/.gemini/skills/branch-pr/SKILL.md |

## Compact Rules

### work-unit-commits

- Commit by work unit (behavior, fix, migration, or docs).
- Keep tests and docs with the code they verify/explain.
- Commit messages should explain the outcome, not just a list of files.
- Group commits into chained PR slices if changes exceed 400 lines.
- Each commit must leave the repo in a working state.

### go-testing

- Use table-driven tests with `t.Run(tt.name, ...)`.
- Test behavior and state transitions, not implementation details.
- Use `t.TempDir()` for filesystem tests.
- Skip slow integration tests with `testing.Short()`.
- For Bubbletea, test `Model.Update()` directly for state; use `teatest` for interactive flows.
- Golden files must be deterministic; update with `-update` flag.
- Use small mocks/interfaces at system boundaries.

### skill-creator

- Create LLM-first skills with valid frontmatter (name, description/trigger).
- Focus on actionable patterns and constraints.
- Document AI usage patterns clearly.

### judgment-day

- Run blind dual reviews to identify issues.
- Fix confirmed issues and re-judge until resolved.
- Use adversarial review to ensure robustness.

### issue-creation

- Use issue-first checks before creating GitHub issues.
- Ensure clear descriptions for bug reports and feature requests.

### comment-writer

- Write warm, direct, and helpful collaboration comments.
- Use for PR feedback, issue replies, and general reviews.

### cognitive-doc-design

- Design documentation to minimize cognitive load.
- Focus on clarity in guides, READMEs, and RFCs.
- Use structured layouts for onboarding and architecture docs.

### chained-pr

- Split changes exceeding 400 lines into smaller, reviewable PRs.
- Use stacked PRs to maintain momentum without overwhelming reviewers.

### branch-pr

- Use issue-first checks when opening pull requests.
- Prepare PRs for review with clear context and intent.

## Project Conventions

| File | Path | Notes |
| :--- | :--- | :--- |
| GEMINI.md | /Users/jesusrosendomercadoporcayo/go/src/PlantillaInstitucional/GEMINI.md | Foundational mandates and conventions. |

Read the convention files listed above for project-specific patterns and rules.
