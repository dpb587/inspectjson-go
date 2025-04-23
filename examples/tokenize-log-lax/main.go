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
				inspectjson.TokenizerConfig{}.
					SetSourceOffsets(true).
					SetLax(true).
					SetSyntaxRecoveryHook(func(e inspectjson.SyntaxRecovery) {
						fmt.Fprintf(os.Stdout, "%-28s\t%-18s\t%q -> %q\n", e.SourceOffsets, e.Behavior, string(e.SourceRunes), string(e.ReplacementRunes))
					}),
			),
		),
	)
	if err != nil {
		panic(err)
	}
}
