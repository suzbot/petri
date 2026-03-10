# Process Flow

How skills, documents, and the plan→implement→test→iterate cycle fit together.

## Phase Lifecycle

```mermaid
flowchart TD

  REQ[/"Requirements"/]

    
    REQ ==> NP["/new-phase"]
    REFS -.->|informs| NP
    V -.->|informs| NP
    NP ==> DD[/"phase-design.md"/]

    DD ==> RF["/refine-feature"]
    REFS -.->|informs| RF
    RF -->|Design Decisions| DD
    RF ==> SS[/"step-spec.md"/]

    SS ==> IF["/implement-feature"]
    IF ==> CODE["Code + Tests"]

    CODE ==> TW["/test-world"]
    TW ==> HT{"Human Testing"}
    HT <-->|bug found| FB["/fix-bug"]

    HT ==>|passes testing| UD["/update-docs"] 
    UD -->|updates| REFS
   

    UD ==> RET["/retro"]
    RET -->|updates| SK[/skills/*.md"/]

    RET ==>|next step| RF
    RET --> V

REFS[/"architecture.md, ngame-mechanics.md, VISION.txt"/]   
V[/"Values.md"/]
```

## Step Rhythm

```
/refine-feature  →  discuss, decide, write step-spec
/implement-feature  →  TDD, build sub-steps
  [TEST]  →  human testing (maybe /test-world)
    → bug? → /fix-bug → retest
  [DOCS]  →  /update-docs
  [RETRO]  →  /retro
→ next step
```

[TEST] → [DOCS] → [RETRO] always appear as a unit at every testable milestone.