package main

import (
	"fmt"
	"io"
	"os"

	"github.com/dpb587/inspectjson-go/inspectjson"
)

func main() {
	_, err := io.Copy(
		io.Discard,
		inspectjson.NewTokenizerReader(
			inspectjson.NewTokenizer(
				os.Stdin,
				inspectjson.TokenizerOptions{}.
					SourceOffsets(true).
					Lax(true).
					SyntaxRecoveryHook(func(e inspectjson.SyntaxRecovery) {
						fmt.Fprintf(os.Stderr, "%-28s\t%-18s\t%q -> %q\n", e.SourceOffsets, e.Behavior, string(e.SourceRunes), string(e.ReplacementRunes))
					}),
			),
		),
	)
	if err != nil {
		panic(err)
	}
}
