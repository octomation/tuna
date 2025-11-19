# Candidate Issues

---

## Interactive query execution

Add feedback to `tuna exec`: progress bar, streaming output, current query/model indicator. Currently the command runs silently until all requests complete.

---

## Chocolatey distribution for Windows

Simplify Windows distribution via GoReleaser with Chocolatey (https://chocolatey.org/) package support.

---

## Interactive getting started guide

Create an interactive tutorial for new users covering configuration setup, model aliases, providers, and basic workflow (init → plan → exec).

---

## Interactive response viewer

TUI for browsing and comparing LLM responses. Navigate by query (Input) and see all model responses side-by-side for each query. Helps evaluate and compare model outputs across the entire assistant.

---

## Response annotation for prompt refinement

Extends the response viewer with ability to mark/tag responses (e.g., "good", "needs improvement", "off-topic"). Annotations help identify patterns and guide system prompt iteration.

---

## Interactive assistant picker for plan command

Allow `tuna plan` to run without arguments. Shows interactive picker with valid assistant folders (those matching expected structure: Input/, system prompt files).

---

## ULID for plans and interactive exec picker

Replace UUID with ULID for plan IDs. ULID is lexicographically sortable by timestamp. Enables `tuna exec` without arguments: discover plan folders and show interactive picker sorted from newest to oldest.

Useful resources:
- https://youtu.be/otW7nLd8P04

---

## TUI execution dashboard

Integrate Charm TUI framework (bubbletea) for `tuna exec`. Real-time dashboard showing: active parallel requests, elapsed time per request, completion progress, ETA, and overall statistics.

---

## Research: prompt evaluation platforms

Study existing solutions for prompt engineering and evaluation:
- DeepEval Arena (https://deepeval.com/docs/evaluation-prompts#arena)
- PromptLayer (https://www.promptlayer.com/)

Identify useful patterns and features that could enhance tuna.

Add useful resources
- https://www.promptingguide.ai/
- https://youtu.be/pwWBcsxEoLk
- https://youtu.be/iRTK-jsfleg
- https://youtu.be/jC4v5AS4RIM
  - https://www.youtube.com/playlist?list=PLo-kPya_Ww2zT0trbGN68Rmh_xZcq_BoR

---

## Cloud storage export

Export execution results to cloud storage: NextCloud, Google Drive, Dropbox. Useful for sharing results with team members or backing up experiments.

---

## System prompt versioning and rollback

Restore system prompt from previously executed plans. Consider: hash-based deduplication to identify unique prompt versions, plan creation prevention if identical prompt already executed, or simply allow navigation through plan history to restore earlier versions.

---

## Response metadata

Add metadata to LLM response files. Metadata should include:
- Provider base URL
- Resolved model name (full name, not alias)
- Request execution time
- Input token count
- Output token count

Format options: YAML frontmatter, separate `.meta.json` file, or embedded in response file header.
