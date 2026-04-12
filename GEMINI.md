# GEMINI.md

## Language Strategies

**[CONFIGURATION]**
- `TARGET_LANGUAGE` = **English**

**1. Internal Reasoning & Code:**
- **Reasoning:** English is the only permitted and recommended language for internal chain-of-thought to maintain optimal accuracy.
- **Codebase:** Use standard English for all source code, including variables, classes, and generic comments.

**2. User-Facing Output:**
- **Primary Rule:** ALL user-facing text **MUST** be written in the `TARGET_LANGUAGE`.
- **Chat:** Communicate exclusively in the `TARGET_LANGUAGE`.
- **Artifacts:** Content within documentation and planning files (e.g., `task.md`, `implementation_plan.md`, `agents_adjustment_proposal.md`, `walkthrough.md`) MUST be written in the `TARGET_LANGUAGE`.

**3. Task Metadata:**
- `TaskName`: **MUST** be in the `TARGET_LANGUAGE`.
- `TaskSummary`: **MUST** be in the `TARGET_LANGUAGE`.
- `TaskStatus`: Allowed in  `TARGET_LANGUAGE`.

## Security & Safety Rules

- Never hardcode API keys, tokens, passwords, or secrets in any file — always use environment variables.
- Never read or print the contents of .env, .env.local, .env.production, or any secrets file.
- Ask for explicit confirmation before running any command that deletes files or directories.
- Ask for explicit confirmation before running `git push`, `git push --force`, or any destructive git operation.
- Never modify files outside the current project root without being explicitly told to do so.
- Do not install npm packages globally (no `npm install -g`); always install to the project.
- Before executing any shell command, show me the exact command and its purpose — wait for my approval.
- Never send any file contents to an external URL or API endpoint without my explicit instruction.
- Always use parameterised queries for database interactions; never concatenate user input into SQL strings.
- Flag any code that reads from `process.env` directly in client-side code — these values are exposed to browsers.

## Code Quality Rules

- Keep functions under 40 lines; if a function is longer, split it into smaller helpers.
- Follow the single-responsibility principle: each module should have exactly one reason to change.
- Do not duplicate logic; if the same code appears twice, extract it into a shared utility.
- Use meaningful, intention-revealing names; avoid abbreviations like `mgr`, `tmp`, or `d`.
- Remove all `console.log` and debug statements before marking a task as done.
- Do not leave TODO comments in code — either fix the issue now or create a tracked issue.
- Prefer composition over inheritance; build behaviour by combining small functions, not deep class hierarchies.
- Every PR diff should be readable in under 10 minutes; if it is larger, ask me to split it into smaller PRs.
- Use early returns to reduce nesting depth; aim for maximum two levels of indentation in function bodies.
- After any refactor, run the linter and type-checker before reporting the task complete.

## Communication & Behavior Rules
- Be concise. Skip preamble like "Certainly!" or "Sure, I can help with that." — get to the answer immediately.
- If a task would require changing more than 3 files, summarise the plan and ask for approval before writing any code.
- When you are unsure about the intended behaviour, ask one focused clarifying question — do not guess.
- Do not explain basic language syntax to me unless I specifically ask for an explanation.
- When suggesting a fix, always explain the root cause in one sentence before showing the fix.
- If I ask you to "clean up" or "improve" code, ask what dimension I care about before making changes (readability, performance, size).
- List every file you plan to modify before you start modifying them.
- If a task is ambiguous, default to the minimal change that satisfies the requirement — do not add features I did not ask for.
- Never apologise repeatedly; one brief acknowledgement is enough if you make a mistake.
- At the end of a multi-step task, summarise what was done in three bullet points or fewer.

 ## Testing Rules
- Every new function or module must have at least one unit test before the task is marked complete.
- Use Vitest for unit and integration tests in this project; never introduce Jest as a separate dependency.
- Test files must live next to the module they test: `foo.ts` → `foo.test.ts` in the same directory.
- Write tests that describe behaviour, not implementation: test what the function does, not how it does it.
- Aim for 80% line coverage as a floor; alert me if any PR would drop it below that threshold.
- Use `describe` blocks to group tests by scenario; use `it` or `test` with a sentence that starts with "it should..."
- Mock external APIs and database calls in unit tests; only use real connections in explicitly labelled integration tests.
- Never snapshot test plain HTML strings — they are brittle; prefer assertion-based tests for component output.
- Run the test suite with `pnpm test` before every commit suggestion; do not suggest committing failing tests.
- If you cannot figure out how to test something, ask me — do not skip the test silently.
