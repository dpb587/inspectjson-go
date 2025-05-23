package main

import (
	"io"
	"os"

	"github.com/dpb587/inspectjson-go/inspectjson"
)

func main() {
	_, err := io.Copy(
		os.Stdout,
		inspectjson.NewTokenizerReader(
			inspectjson.NewTokenizer(
				os.Stdin,
				inspectjson.TokenizerConfig{}.
					SetLax(true),
			),
		),
	)
	if err != nil {
		panic(err)
	}
}
