# Changelog

Everything worth knowing about a release of Arandu Markdown is recorded here.

The format follows [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) and
the versions follow [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

A published module version is immutable: Go serves it from the proxy forever, so
a release is corrected by another release and never by moving a tag.

## [Unreleased]

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
