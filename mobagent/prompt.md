

> **让 Cursor Agent 一次性实现 Telegram Mini App，用于远程控制电脑上的 Cursor Agent / Claude Code / Codex，并且重点保证 Agent 的实时反馈体验。**

可以直接保存成 `cursor.md`，然后丢给 Cursor Agent。

````markdown
# AI Coding Agent Remote Controller

## ROLE

You are the lead Product Designer, UX Designer, Software Architect, Senior Frontend Engineer, Backend Engineer and QA Lead.

Build a production-ready Telegram Mini App.

This is NOT a demo.

This is NOT an MVP.

Everything must be production quality.

The goal is to provide the best remote coding experience on mobile.

The application should feel like Cursor + Linear + Raycast + Telegram Native.

---

# PRODUCT

Telegram Mini App

Remote control local AI Coding Agents.

Supported Agents

- Cursor Agent
- Claude Code
- OpenAI Codex
- Gemini CLI
- Aider
- OpenHands

Architecture must support adding new Agent adapters without modifying existing code.

---

# PRODUCT GOAL

Users are away from their computer.

Using Telegram Mini App they can:

- Create Task
- Continue Task
- Pause Task
- Resume Task
- Stop Task
- Retry Task
- Review Task
- Explain Current Progress
- View Live Agent Output
- View Timeline
- View Live Diff
- View Current File
- View Workspace
- View Running Command
- View Test Status
- View Git Status
- View CPU
- View Memory
- View Token Usage
- View Cost

Everything updates in realtime.

Never wait until task completes.

---

# CORE EXPERIENCE

The experience is NOT terminal streaming.

The experience is Event Streaming.

Never directly display terminal output as the primary UI.

Instead transform terminal output into structured realtime events.

Terminal output only appears inside Console page.

The Home page should only display structured events.

---

# ARCHITECTURE

Implement:

Agent

↓

Adapter

↓

Raw Output

↓

Output Parser

↓

Normalized Events

↓

Event Engine

↓

WebSocket

↓

Telegram Mini App

The frontend NEVER parses terminal output.

Frontend only consumes Events.

---

# EVENT ENGINE

Create an Event Engine.

Responsibilities

- Parse
- Normalize
- Deduplicate
- Merge
- Prioritize
- Broadcast

Example

Cursor Output

Reading package.json

Reading package.json

Reading package.json

↓

Single Event

Reading Project...

42 files scanned

Claude Code

Thinking...

Thinking...

↓

Planning Implementation...

Elapsed 18s

Never flood UI.

---

# EVENT MODEL

Every event must contain

```ts
{
    id
    taskId
    sessionId
    timestamp

    agent
    workspace

    type
    status

    payload
}
```

Supported Event Types

- TaskStarted
- Planning
- ReadingProject
- ReadingFile
- Searching
- Editing
- CreatingFile
- UpdatingFile
- RemovingFile
- RunningCommand
- InstallingDependencies
- Building
- RunningTests
- Reviewing
- GitCommit
- GitPush
- NeedUserInput
- Warning
- Error
- Finished
- Failed
- DiffUpdated
- ConsoleOutput
- CostUpdated
- ResourceUpdated
- Heartbeat

Support future extensions.

---

# WEBSOCKET

Use WebSocket only.

Never use polling.

Support

- reconnect
- heartbeat
- offline
- resume
- retry

Heartbeat every 10 seconds.

Resume task after reconnect.

No duplicated events.

No event loss.

---

# UI

Telegram Mini App.

Dark Mode.

Native feeling.

One-hand operation.

Large touch targets.

Bottom Sheet.

No desktop UI.

No admin dashboard style.

Reference

- Telegram
- Cursor
- Linear
- Raycast
- Arc
- Warp

The interface should feel premium.

---

# PAGES

## Home

Display running agents.

Each card contains

- Agent Avatar
- Agent Name
- Workspace
- Status
- Progress
- Current Stage
- Current File
- Running Time
- Token Usage
- Cost

Quick Actions

- Pause
- Resume
- Stop
- Restart
- Continue

---

## Timeline

Render Events.

Examples

Task Started

↓

Reading Project

↓

Searching

↓

Planning

↓

Editing

↓

Testing

↓

Reviewing

↓

Finished

Animated.

Realtime.

---

## Console

Raw output.

Streaming.

Support

ANSI Color

Search

Copy

Pause Scroll

Auto Scroll

Virtual Scroll

---

## Live Diff

Realtime update.

Display

Modified

Added

Deleted

Renamed

Tap to open Diff.

---

## Current Context

Always display

Workspace

Current Branch

Current File

Elapsed Time

CPU

Memory

Token Usage

Estimated Cost

---

## Task Composer

Bottom Sheet.

Support

Text

Voice

Image

Screenshot

File

Paste

---

# QUICK ACTIONS

Continue

Explain

Pause

Resume

Retry

Review

Stop

Merge

Push

Cancel

---

# ANIMATION

All animations

200~300ms

60fps

No layout shift.

Use Motion.

---

# DESIGN SYSTEM

8px spacing.

Large radius.

Glass effect.

Blur.

Premium typography.

Responsive.

Native Telegram style.

---

# PERFORMANCE

Virtual List.

Lazy Load.

Memoization.

Keep WebSocket lightweight.

Avoid unnecessary rerender.

No memory leak.

Support 100000+ events.

---

# TECH STACK

Vue 3.5

TypeScript

Composition API

Pinia

TailwindCSS



Motion

Floating UI

Vite8

Never use deprecated APIs.

---

# TESTING

Generate automatically

Vitest

Playwright

Unit Tests

Integration Tests

E2E Tests

WebSocket Tests

Reconnect Tests

Offline Tests

Performance Tests

Stress Tests

Memory Leak Tests

Animation Tests

---

# STRESS TEST

Automatically simulate

100000 Events

10000 Console Logs

1000 Diff Updates

Network Disconnect

Reconnect

Agent Restart

Large Workspace

Long Running Task

8 Hour Task

Multiple Agents Running

Verify UI remains smooth.

---

# QUALITY GATE

Must pass

No TypeScript Error

No ESLint Error

No Build Error

No Runtime Error

No Memory Leak

No WebSocket Leak

No Lost Event

No Duplicate Event

No Layout Shift

No Accessibility Error

Responsive

Dark Mode

60fps Animation

---

# SELF REVIEW

When implementation is complete:

DO NOT STOP.

Review the entire project.

Improve UX.

Improve UI consistency.

Improve animations.

Improve architecture.

Reduce duplicated code.

Optimize performance.

Run all tests again.

Fix every failed test.

Repeat until:

- All tests pass
- Build passes
- Lint passes
- TypeScript passes
- UI is polished
- UX feels production-ready

Only finish when the application is ready for real users.

---

# DELIVERABLES

Before completion automatically generate:

- Architecture documentation
- Component documentation
- Event flow diagram
- WebSocket protocol
- State management documentation
- Test report
- Performance report
- Accessibility report
- Remaining TODO list (if any)

No unfinished code.

No placeholder.

No mock implementation.

Everything must be fully functional.
````
