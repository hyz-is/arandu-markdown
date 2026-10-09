# Skills

Procedures an assistant follows when working on this package.

They live in `.agents/skills/<name>/SKILL.md`, which is the path the coding
assistants read from — Cursor, Codex, Cline, Copilot, Gemini CLI, Amp, OpenCode,
Warp, Zed and the rest all look there. It is one directory rather than a file
per vendor, so a skill written once is read by whatever this package is being
written with.

Each file opens with frontmatter carrying a `name` and a `description`. The
`name` has to equal the directory name exactly, or the skill is not loaded. The
description is what a tool reads to decide whether the skill is relevant, so it
names the situation you are in rather than the topic it covers.

| skill | when it fires |
| --- | --- |
| `markdown-module` | changing what the renderer writes: an element, an attribute, a destination, a heading id, a config field |
| `markdown-release` | the gates, the manifest, a dependency, a version, a tag |
| `markdown-vault-notes` | writing the note, the gap or the journal entry, when this checkout sits inside the Arandu Obsidian vault |
| `markdown-package` | installing and wiring this package **into an application** |

The last one has a different audience from the other three, and that is on
purpose: it travels with the package so that an assistant working in somebody
else's project — the one running `go get` — has the wiring, the content
security policy and the caching in front of it instead of guessing.

Its frontmatter says so, under `metadata`, with `audience: app`. That line is
what `aru skills:sync` reads to copy the skill into an application whose
`go.mod` requires this package, and no other skill here carries it: an
application that received the release procedure would follow it.

`markdown-vault-notes` fires on a condition rather than on a task: it applies
only when `MOC-arandu.md`, `plans/cmd/audit-vault/` and `45-modules/` are actually
beside this checkout. Outside the vault it is inert, and it says so first, so a
package cloned somewhere else never grows a folder tree imitating one.

## Why these exist

The audience of the first three is somebody changing the package. The common
failure modes are a second Markdown parser, a sanitizer library pulled in beside
the tokenizer, an allowlist entry for something the parser never writes, a
configuration switch that widens the list, and a refused tag that vanishes
instead of showing. None belongs here, and the first four are how a sanitizer
stops being auditable.

## Adding your own

A skill in this directory is yours and travels with the repository. Keep it a
procedure rather than a description: a file that says "read the documentation"
never changes what anybody does.
