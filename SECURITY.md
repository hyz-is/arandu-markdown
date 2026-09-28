# Security policy

## Supported versions

The latest minor release of Arandu Markdown receives security fixes. Older minor
releases do not: a published Go module version is immutable, so a fix is a new
release rather than a moved tag.

## Reporting a vulnerability

Report it privately, through GitHub's advisory form:

<https://github.com/hyz-is/arandu-markdown/security/advisories/new>

Do not open a public issue and do not describe the problem in a pull request.
A report that arrives in public is a report every reader of this repository can
act on before there is a release to upgrade to.

Include what you need to reproduce it: the version, the configuration, and the
source that triggers it.

Expect an acknowledgement within a few days. If the report is confirmed, the
fix, the release and the advisory are published together, and you are credited
unless you ask not to be.

## What is in scope

Anything that lets a body put markup on a page that the allowlist did not write.
In particular:

- an element or an attribute in `Document.HTML` that `render.go` did not write;
- a link or image destination with a scheme other than the ones it keeps, or an
  image from an origin `Config.ImageOrigins` does not name;
- a way out of an escaped text node or attribute value;
- a source that makes `Render` fail, or run for time out of proportion to its
  length.

## What is not

- A vulnerability in `hesape/str.Markdown` itself that this package's rewrite
  already neutralizes. Report it to hesape all the same.
- What an application does with `Document.Text`: it is plain text, and a view
  that writes it unescaped has chosen to.
