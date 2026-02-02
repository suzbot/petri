# skills/new-feature.md

  ---
  name: new-feature
  description: "Start implementing any feature either within the current phase or as a stand-alone. Enforces discussion-first, TDD, and human testing checkpoints."
  ---

  ## Starting a New Feature

  ### Step 1: Review Context (REQUIRED)
  Read these before discussing approach:
  - Reread any requirements (usually in docs/ ) for full context
  - Phase plan document (if exists)
  - Feature-specific plan (if exists)
    - Any questions deferred from phase planning

  ### Step 2: Discussion First (REQUIRED)
  **Do NOT write code yet.** First:
  - Ask any deferred clarifying questions
  - Present implementation options with trade-offs
  - Get user confirmation on approach
  - Document decisions in the feature plan

  ### Step 3: TDD Implementation
  Once approach is confirmed:
  - Write tests first
  - Implement minimum code to pass tests
  - Run tests to verify
  - Pause when broader design questions emerge for discussion before proceeding

  ### Step 4: Human Testing Checkpoint (REQUIRED)
  **Do NOT mark feature complete until user has tested.**
  - Pause for user to rebuild and manually test
  - If scenario for verification is complex, consider running skills/test-world
  - Wait for explicit confirmation from user before updating docs

  ### Step 5: Documentation
  Only after human testing confirms success:
  - Update README, CLAUDE.md, docs/game-mechanics, and docs/architecture as needed
  - Mark feature complete in phase plan
  - Consider running skills/retro

  ---

  **Collaboration Norms:**
  - TDD for bug fixes too - write regression tests
  - Small iterations - keep changes focused