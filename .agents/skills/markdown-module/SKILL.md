---
name: markdown-module
description: Change what this Arandu package's renderer writes. Use when the request is to "allow an element", "keep an attribute", "support iframes", "allow images from another host", "add a class to code blocks", "change heading ids", "list h4 in the table of contents", "add a config option", "make it configurable", "the renderer drops my tag", "a link disappeared", or when a change touches render.go, config.go or module.go. Covers the allowlist and why it only follows str.Markdown, destinations, heading ids, the four settings, and why there is no route.
license: MIT
---

# Changing the renderer

`render.go` reads what `hesape/str.Markdown` returns with the HTML5 tokenizer
from `golang.org/x/net/html` and writes it again. The allowlist is the
`elements` map at the top of that file, and the attributes each element keeps
are written out in `start` — there is no attribute map to widen, on purpose.

## Adding an element or an attribute

Only when `str.Markdown` writes it. The list mirrors the parser; something the
parser never emits can only arrive as raw HTML somebody typed, and raw HTML is
shown as text. Check first:

```sh
export GOWORK=off
grep -n '<[a-z]' "$(go env GOMODCACHE)/github.com/arandu-io/hesape@$(go list -m -f '{{.Version}}' github.com/arandu-io/hesape)/str/markdown"*.go
```

Then, in one commit:

1. the entry in `elements`, and the attributes it keeps in `start`, each value
   validated or rebuilt rather than copied;
2. the same element and attributes in `allowed`, in `tests/Unit/render_test.go`;
3. a test showing the element coming out of Markdown and a hostile variant of it
   coming out as text;
4. a changelog line — an element that starts being written is markup somebody's
   stylesheet and content security policy never saw.

A request to allow `iframe`, `video`, `script`, `style`, `form`, `object`,
`embed` or an event attribute is refused: none of them is Markdown, and each is
a way to run or load something the page did not choose.

## Destinations

`href` and `src` are the two ways a body reaches something outside itself.
`(*Module).href` keeps http(s) with a host and no userinfo, `mailto:` with an
address, same-site paths with no `..` segment, and `#` anchors shaped like the
ids headings get. `(*Module).src` keeps same-site paths and https URLs on an
origin `Config.ImageOrigins` names. A new scheme needs the argument in the pull
request; a new origin belongs in the installer's configuration, not here.

## Heading ids

`str.Slug(text, "-")`, cut at a word past 80 characters, `section` when nothing
is left, and `-2`, `-3` when taken or reserved. Changing how an id is made
changes addresses people have shared: it is a changelog line, and a minor
version while the major is 0.

## Configuration

`Config` has four fields, validated by `New` and defaulted by `withDefaults`
after validation, never before. A new field is a new thing every installer has
to understand, and a switch that widens the allowlist is refused: a list each
installer can widen is a list nobody can audit.

## No route, no table

`Routes` registers nothing. A body reaches a page through the controller that
authorized its entry; a preview endpoint here would render text for anybody,
with no entry and no Grant behind it. A route or a table would also change
`arandu.mod.toml`, and the audit suite says so.
