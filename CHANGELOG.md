# Changelog

Everything worth knowing about a release of Arandu Markdown is recorded here.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and
the versions follow [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

A published module version is immutable: Go serves it from the proxy forever, so
a release is corrected by another release and never by moving a tag.

## [Unreleased]

## [0.2.4] - 2026-10-09

### Changed

- Requires `framework` v0.56.0 and `hesape` v0.54.0, up from v0.55.1 and
  v0.52.0, and `arandu.mod.toml` declares `framework = ">= 0.56"`. Nothing
  here read `config.Config.SessionTTL`, which that framework release removes.

## [0.2.3] - 2026-10-09

### Changed

- Requires `framework` v0.55.1, `hesape` v0.52.0 and `golang.org/x/net`
  v0.60.0, up from v0.50.0, v0.50.3 and v0.59.0, and `arandu.mod.toml`
  declares `framework = ">= 0.55"`. The `x/net` release fixes six advisories
  the code did not reach. Nothing here calls a symbol `hesape` v0.52.0 removed
  or deprecated.

## [0.2.2] - 2026-10-09

### Changed

- Requires `hesape` v0.50.3 and `framework` v0.50.0, up from v0.43.1 and
  v0.49.0, and `arandu.mod.toml` declares `framework = ">= 0.50"`. v0.50.0 is the
  lowest framework release that builds against that `hesape`.

### Fixed

- A body that nests quotes or lists and continues their paragraph with lazy
  lines no longer makes `Render` allocate hundreds of bytes for each byte of
  it. A paragraph under a hundred nested quotes took about 580, 23 MB for 40 KB
  of text, and takes 36 now, this package's own rewrite included: the
  `str.Markdown` it requires holds the lines of a nested level once instead of
  copying them at every level. `TestAHostileNestedBodyIsHeldOnce` fails above
  100.
- The `Render` doc comment and the README no longer call `str.Markdown`
  quadratic on a run of link openers that never close. It is linear there, and
  they still advise capping the length of a body nobody vetted.

## [0.2.1] - 2026-10-09

### Added

- The `arandu-ecosystem` skill: the shared architecture an application keeps
  before it builds a second implementation of what an Arandu module owns.

### Changed

- The `markdown-package` skill carries `audience: app` under `metadata` in its
  frontmatter, which is what `aru skills:sync` reads to copy it into an
  application whose `go.mod` requires this package. No other skill is marked.

## [0.2.0] - 2026-09-28

### Changed

- `Document.HTML` is a method, `Document.HTML()`, instead of a field. A Kyse
  view writes it as `{!! .Body.HTML() !!}`, the call shape `aru doctor` accepts
  for raw output; a field there read as a value and drew a warning in every
  application that installed the module.

### Fixed

- The release workflow no longer vets `configure.go`, which a configured copy
  does not have.

## [0.1.0] - 2026-09-28

### Added

- `New(Config)` and `(*Module).Render(source) Document`: a Markdown body
  rendered by `hesape/str.Markdown` and rewritten through a closed allowlist —
  paragraphs, headings, emphasis, strikethrough, code, lists, task lists,
  quotes, tables, links and images. Anything else is shown as the text it was.
- Links keep http(s) with no userinfo, `mailto:` with an address, same-site
  paths and in-page anchors. Images keep same-site paths and the origins in
  `Config.ImageOrigins`, and are written with `loading="lazy"` and
  `decoding="async"`. An image alone in its paragraph becomes a `figure`, with
  its title as the caption.
- Headings are lowered to `Config.TopHeading` (default 2) and carry unique ids
  that avoid `Config.Reserved`. `Document.Headings` lists the top two levels for
  a table of contents.
- `Document.Text`, `Words` and `Minutes`: the body as plain text, one line per
  block, with its word count and reading time at `Config.WordsPerMinute`
  (default 200).
- A source the parser cannot read is shown as escaped paragraphs instead of
  failing the render.
- `Module` implements `foundation.Module`, registers no route and declares no
  capability in `arandu.mod.toml`.
