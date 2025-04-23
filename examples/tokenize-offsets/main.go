package main

import (
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/dpb587/inspectjson-go/inspectjson"
)

func main() {
	tokenizer := inspectjson.NewTokenizer(
		os.Stdin,
		inspectjson.TokenizerConfig{}.
			SetSourceOffsets(true).
			SetLax(true),
	)

	for {
		token, err := tokenizer.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			panic(err)
		}

		b, err := token.AppendText([]byte{})
		if err != nil {
			panic(err)
		}

		fmt.Fprintf(os.Stdout, "%-28s\t%-18s\t%s\n", token.GetSourceOffsets(), token.GetGrammarName(), b)
	}
}
