When the user is ready to start a new feature within a phase:

## Feature Development Flow

### 1. Discussion First (REQUIRED)
- Review the requirements, phase plan, and feature plan
- **Discuss approach before writing any code**
- Ask clarifying questions
- Present options with trade-offs when decisions are needed
- Get user confirmation on approach

### 2. Update Plan
- Document clarifications/decisions/approaches in plan before implementing
- This creates a record of design decisions

### 3. Implementation (TDD)
- Write tests first (see docs/testingProcess)
- Implement the feature
- Run tests to verify
- When broader design questions emerge, pause to discuss

### 4. Human Testing (REQUIRED)
- **Do NOT mark feature complete until user has tested**
- At test checkpoints, pause for user to rebuild and manually test
- Wait for user confirmation before updating documentation as "complete"

### 5. Documentation
- Update README, CLAUDE.md, game-mechanics as needed
- Only mark tasks complete in phase plan after human testing confirms success

---

## Collaboration Norms

**Discussion before code**: Always discuss approach before implementing. Don't jump straight to writing code or tests.

**User runs builds and tests**: After making code changes, wait for user to run `go build` and `go test` rather than assuming success. The user will share output.

**TDD for bug fixes too**: When fixing bugs discovered during testing, write regression tests first (or alongside). Don't implement fixes without corresponding tests.

**Human testing before completion**: Features are not complete until the user has manually tested them. Do not mark phase plan items as done until user confirms.

**Small iterations**: Keep changes focused and reviewable. Present options with trade-offs when decisions are needed.
