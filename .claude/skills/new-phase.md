# skills/new-phase.md

  ---
  name: new-phase
  description: "Start planning a new development phase. Reads requirements, facilitates discussion, and creates the phase plan."
  ---

  ## Starting a New Phase

  Follow this process to plan the next development phase:

  ### Step 1: Identify the Next Phase
  Read the roadmap section of CLAUDE.md to find what's next. This skill only needs to be leveraged if the next item is a phase or large feature with a requirements doc.

  ### Step 2: Read Requirements
  - Find and read the requirements document for this phase (typically in `docs/`).
  - For context on where this phase fits within the longer term plan and how it sets up future phases, reference docs/VISION.txt.

  ### Step 3: Discuss Approach
  Before creating a plan:
  - Use skills/architecture to get a jumpstart on finding and understanding current design patterns
  - Ask clarifying questions about requirements
  - Present high-level approach options with trade-offs
  - Identify an iterative approach with frequent human testing checkpoints
  - Only ask questions that affect high-level decisions - feature-specific questions wait

  ### Step 4: Create Phase Plan
  Save the high-level phase plan to `docs/[name]-phase-plan.md` including:
  - Original requirements reference
  - User-facing outcomes for each sub-phase
  - [TEST] checkpoints where user can verify behavior
  - Suggest any opportunistic or triggered enhancements from docs/randomideas.md or triggered-enhancements.md that could be worked into the plan.

  ### Step 5: Note Feature Questions
  For clarifications that don't impact the high-level approach, save those questions in feature plan docs to ask in context during implementation.

  ---

  **Reminder:** Features within a phase are approached one at a time with fresh discussion, todos, and testing for each. Use `/new-feature` when starting each one.
