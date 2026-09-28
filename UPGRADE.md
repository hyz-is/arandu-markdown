# Upgrade Guide

## Unreleased

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
