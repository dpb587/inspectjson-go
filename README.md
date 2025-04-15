# inspectjson-go

Parse JSON with imperfect syntax and capture metadata about byte offsets.

* Decode human-crafted, error-prone JSON often found on web pages.
* Reference byte and line+column offsets of JSON structures, keys, and values.
* Sanitize the syntax of JSON for any strict decoder implementation.
* Describe invalid syntax and suggest replacements.

This is implemented as a custom tokenizer based on the official JSON standards, but with configurable behaviors for edge cases of non-compliant syntax.

## Usage

Import the module and refer to the code's documentation ([pkg.go.dev](https://pkg.go.dev/github.com/dpb587/inspectjson-go/inspectjson)).

```go
import "github.com/dpb587/inspectjson-go/inspectjson"
```

Some sample use cases and starter snippets can be found in the [`examples` directory](examples).

<details><summary><code>examples$ go run ./<strong>parse-value</strong> <<<'<strong>{"n":true}</strong>'</code></summary>

```go
inspectjson.ObjectValue{
  BeginToken: inspectjson.BeginObjectToken{
    SourceOffsets: &cursorio.TextOffsetRange{
      From: cursorio.TextOffset{Byte: 0, Line: 0, LineColumn: 0},
      Until: cursorio.TextOffset{Byte: 1, Line: 0, LineColumn: 1},
    },
  },
  EndToken: inspectjson.EndObjectToken{
    SourceOffsets: &cursorio.TextOffsetRange{
      From: cursorio.TextOffset{Byte: 9, Line: 0, LineColumn: 9},
      Until: cursorio.TextOffset{Byte: 10, Line: 0, LineColumn: 10},
    },
  },
  Members: map[string]inspectjson.ObjectMember{
    "n": inspectjson.ObjectMember{
      Name: inspectjson.StringValue{
        SourceOffsets: &cursorio.TextOffsetRange{
          From: cursorio.TextOffset{Byte: 1, Line: 0, LineColumn: 1},
          Until: cursorio.TextOffset{Byte: 4, Line: 0, LineColumn: 4},
        },
        Value: "n",
      },
      Value: inspectjson.BooleanValue{
        SourceOffsets: &cursorio.TextOffsetRange{
          From: cursorio.TextOffset{Byte: 5, Line: 0, LineColumn: 5},
          Until: cursorio.TextOffset{Byte: 9, Line: 0, LineColumn: 9},
        },
        Value: true,
      },
    },
  },
  ReplacedMembers: nil,
}
```

</details>

<details><summary><code>examples$ go run ./<strong>tokenize-offsets</strong> <<<'<strong>{"n":true}</strong>'</code></summary>

```
0x0:0x1;L1C1:L1C2               begin-object            {
0x1:0x4;L1C2:L1C5               string                  "n"
0x4:0x5;L1C5:L1C6               name-separator          :
0x5:0x9;L1C6:L1C10              true                    true
0x9:0xa;L1C10:L1C11             end-object              }
```

</details>

<details><summary><code>examples$ go run ./<strong>tokenize-log-lax</strong> <<<'<strong>[01,TRUE,"hello	world",]//test</strong>'</code></summary>

```
0x1:0x2;L1C2:L1C3               LaxNumberTrimLeadingZero        "0" -> ""
0x4:0x8;L1C5:L1C9               LaxLiteralCaseInsensitive       "TRUE" -> "true"
0xf:0x10;L1C16:L1C17            LaxStringEscapeMissingEscape    "\t" -> "\\t"
0x16:0x17;L1C23:L1C24           LaxIgnoreExtraComma     "," -> ""
0x18:0x1e;L1C25:L1C31           LaxIgnoreLineComment    "//test" -> ""
```

</details>

<details><summary><code>examples$ go run ./<strong>tokenize-sanitize</strong> <<<'<strong>[01,TRUE,"hello	world",]//test</strong>'</code></summary>

```json
[1,true,"hello\tworld"]
```

</details>

More complex usage can be seen from importers like [rdfkit-go](https://github.com/dpb587/rdfkit-go).

## Parser

Given an `io.Reader`, parse and return a `Value`. The `Value` interface is implemented by the grammar value types (e.g. `BooleanValue`, `ObjectValue`), and they include fields for source offsets, scalar values, and other tokenization metadata, such as start/end delimiters.

```go
value, err := inspectjson.Parse(
  os.Stdin,
  inspectjson.TokenizerOptions{}.
    Lax(true).
    SourceOffsets(true),
)
```

### Parser Options

Use `ParserOptions` to chain any of the following customizations and use it as an extra argument. The [tokenizer options](#tokenizer-options) may also be used. Snippets in bold are a default behavior.

* `KeepReplacedObjectMembers(bool)`
  * **`KeepReplacedObjectMembers(false)`** - a previously-encountered member will be dropped (i.e. last member wins).
  * `KeepReplacedObjectMembers(true)` - replaced members will be moved and appended to the `ReplacedMembers` field.

## Tokenizer

Given an `io.Reader`, iterate over each `Token`. The `Token` interface is implemented by the grammar syntax types (e.g. `LiteralTrueToken`, `BeginObjectToken`) and include a field for source offsets and, if arbitrary, its content.

```go
tokenizer := inspectjson.NewTokenizer(
  os.Stdin,
  inspectjson.TokenizerOptions{}.
    Lax(true).
    SourceOffsets(true),
)

for {
  token, err := tokenizer.Next()
  if err != nil {
    if errors.Is(err, io.EOF) {
      break
    }

    panic(err)
  }

  switch tt := token.(type) {
  case inspectjson.BeginArrayToken:
  // ...
  }
}
```

The contents of a token will be the decoded string representation for its type (including the effects of any syntax recovery). For example, the contents of a `StringToken` may include literal new lines and UTF-16 code points.

### Tokenizer Options

Use `TokenizerOptions` to chain any of the following customizations and use it as an extra argument. Snippets in bold are a default behavior.

* `EmitWhitespace(bool)`
  * **`EmitWhitespace(false)`** - no whitespace tokens will be returned.
  * `EmitWhitespace(true)` - whitespace tokens will be returned.
* `Lax(bool)`
  * **`Lax(false)`** - requires adherence to JSON syntax.
  * `Lax(true)` - allow all of the recoverable syntax errors.
* `Multistream(bool)`
  * **`Multistream(false)`** - once a value has been completed, `EOF` is expected.
  * `Multistream(true)` - values will continue to be tokenized until `EOF`.
* `SourceOffsets(bool)`
  * **`SourceOffsets(false)`** - no offset data is included in tokens.
  * `SourceOffsets(true)` - capture byte and text line+column offsets for each token.
* `SourceInitialOffset(TextOffset)` - use a non-zero, initial offset (and enable capture of offset data).
* `SyntaxBehavior(SyntaxBehavior, bool)` - allow or disallow a specific behavior.
* `SyntaxRecoveryHook(SyntaxRecoveryHookFunc)`
  * **`SyntaxRecoveryHook(nil)`** - syntax recovery will be handled silently.
  * `SyntaxRecoveryHook(f)` - for each recovered syntax occurrence, `f` will be invoked.

### Tokenizer Reader

To use the tokenizer as a sanitization pipeline for a generic JSON decoder, create an `io.Reader` from it.

```go
decoder := json.NewDecoder(inspectjson.NewTokenizerReader(tokenizer))
```

## Syntax

Several `SyntaxBehavior` constants describe optional tokenization behaviors which may be configured via [tokenizer options](#tokenizer-options). The following describe behaviors for common human mistakes and non-standard encoders.

* `LaxIgnoreBlockComment` - ignore `/* block */` comments.
* `LaxIgnoreLineComment` - ignore `// line` comments (which continues until end of line).
* `LaxStringEscapeInvalidEscape` - convert, for example, `\z` (invalid) into `\\z`.
* `LaxStringEscapeMissingEscape` - convert, for example, `	` (tab, U+0009) into `\t`.
* `LaxNumberTrimLeadingZero` - trim invalid, leading zeros of a number.
* `LaxLiteralCaseInsensitive` - allow case-insensitive literals, such as `True`.
* `LaxIgnoreExtraComma` - ignore any repetitive or trailing commas within arrays or objects.
* `LaxIgnoreTrailingSemicolon` - ignore any semicolon after a value.

Additionally, the following warning may be observed if the recovery hook is used.

* `WarnStringUnicodeReplacementChar` - invalid Unicode sequence was replaced with U+FFFD.

### Recovery Hook

When `SyntaxRecoveryHook` is used, each recovered syntax occurrence will result in a `SyntaxRecovery` being emitted which includes metadata about the source offsets, source runes, value start, and replacement runes, as applicable.

## Resources

* [RFC 8259](https://datatracker.ietf.org/doc/html/rfc8259) - The JavaScript Object Notation (JSON) Data Interchange Format
* [Parsing JSON is a Minefield](https://seriot.ch/projects/parsing_json.html)

## License

[MIT License](LICENSE)
