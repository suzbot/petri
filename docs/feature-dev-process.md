When the user is ready to start a new feature within a phase:

1. Review the claude.md file for current roadmap status
2. Review the requirements document, phase plan, and feature plan for that feature, ask questions, and confirm approach before starting implementation.
3. Update plan document with new clarifications/decisions/approaches before starting implementation.
4. When starting implementation, Use a TDD approach, see docs/testingProcess for details
5. When implementing reveals broader design questions, pause to review architecture before proceeding.
6. **Keep docs current**: Include README, Claude.md, and docs/game-mechanics updates as todo items when implementing features. Don't defer documentation to end of session.

---

## Collaboration Norms

**User runs builds and tests**: After making code changes, wait for user to run `go build` and `go test` rather than assuming success. The user will share output.

**TDD for bug fixes too**: When fixing bugs discovered during testing, write regression tests first (or alongside). Don't implement fixes without corresponding tests.

**Human testing checkpoints**: At [TEST] markers in phase plans, pause for user to rebuild binary and manually test. Don't proceed until user confirms behavior.

**Small iterations**: Discuss approach before implementing. Present options with trade-offs when decisions are needed. Keep changes focused and reviewable.
