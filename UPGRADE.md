# Upgrade Guide

## Unreleased

## v0.2.2

Nothing to change in code that calls this package. The release raises its
floors, so `go get github.com/hyz-is/arandu-markdown@v0.2.2` brings `framework`
v0.50.0 and `hesape` v0.50.3 into an application that required less. One still
on `framework` v0.49 follows the framework's upgrade guide for v0.50.0 first:
`middleware.KeyBySession` takes the session store, and `CSRFProtect` issues a
guest's CSRF token.

## v0.2.1

Nothing to change in an application: this release touches skills, not Go code.
`aru skills:sync` now offers `markdown-package` to a project that requires this
version.

## v0.2.0

### `Document.HTML` is a method

`Document.HTML` was a field and is now `Document.HTML()`. Add the parentheses
wherever it is read:

```go
doc.HTML   // v0.1.0
doc.HTML() // v0.2.0
```

and in a Kyse view, `{!! .Body.HTML !!}` becomes `{!! .Body.HTML() !!}`. The
markup is the same; the call is the shape `aru doctor` accepts around raw
output.

## v0.1.0

The first release. Nothing to upgrade from.
