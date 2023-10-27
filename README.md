# JSONR - Relaxed JSON with Comments
[![GoDoc](https://godoc.org/github.com/msolo/jsonr?status.svg)](https://godoc.org/github.com/msolo/jsonr)

JSONR allows parsing chunks of JSON that contain helpful comments as well as stray, but semantically unambiguous commas in the name of simpler diffs and fewer wasted brain cycles.

The original motivation was to have usable config files without having to resort to things like YAML that are staggeringly complex despite apparent simplicity.

The `jsonr` package is drop-in compatible for decodeing the `json` package in the standard library. 

The performance of this parser is perfectly adequate. Although it is 60% slower than the `json` decoder in the Go standard library, it manages to be about 10x faster than trying to use Jsonnet to handle this degenerate case of simple commented JSON.


## Sample JSONR Snippet
```java
/*
You can see that comments are safely valid in any sane place.

You can put a lengthy preamble or embed a poem if necessary.
*/
{
  // Line comment.
  "x": "a string", // Trailing comment.
  "y": 1.0,
  "z": null,
  "quoted-range": "/* this is not a comment *",
  "quoted-line": "// this is also not a comment",
  // "a": "value temporarily removed for debugging or idle curiosity",
  "array": [],
  "dict": {},  // We can now have a trailing comma here.
}
// Post-amble.
```

## Sample Usage in Go
```go
v := make(map[string]interface{})
f, _ := os.Open("sample.jsonr")
dec := jsonr.NewDecoder(f)
if err := dec.Decode(&v); err != nil {
  return err
}
```

## Command Line Tools

### `jsonr`

`jsonr` is simple tool to filter out comments and trailing commas so standard tools like `jq` are still useful. The output mimics the input so that the order of fields in an object is preserved.

```
go install github.com/msolo/jsonr/cmd/jsonr

jsonr < sample.jsonr

jsonr < sample.jsonr | jq .x
"a string"
```

### `jsonr-fmt`

`jsonr-fmt` formats JSONR in a deterministic way.

```
go install github.com/msolo/jsonr/cmd/jsonr-fmt

jsonr-fmt < sample.jsonr
```

### `jsonr-dump`

`jsonr-dump` dumps JSONR in a deterministic way using a line-oriented key-value notation that is easy to grep.

```
go install github.com/msolo/jsonr/cmd/jsonr-dump
```

```
cat sample.jsonr
/*
You can see that comments are safely valid in any sane place.

You can put a lenghty preamble or embed a poem if necessary.
*/
{
  // Line comment.
  "x": "a string", // Trailing comment.
  "y": 1.0,
  "z": null,
  "quoted-range": "/* this is not a comment *",
  "quoted-line": "// this is also not a comment",
  "value with newlines": "this is also not a comment
but contains a newline a tab (	) and should still remain on one line.",
  // "a": "value temporarily removed for debugging or idle curiosity",
  "array": [1],
  "dict": {"key": null}, // We can have a trailing comma here.
  "a/b/c": "grubby key",
}
// Postamble.
```

```
jsonr-dump sample.jsonr
/x = "a string"
/y = 1.0
/z = null
/quoted-range = "/* this is not a comment *"
/quoted-line = "// this is also not a comment"
/value with newlines = "this is also not a comment\nbut contains a newline a tab (\t) and should still remain on one line."
/array/0 = 1
/dict/key = null
/a\/b\/c = "grubby key"
```
