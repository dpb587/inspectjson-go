package main

import (
	"fmt"
	"os"
	"regexp"

	"github.com/dpb587/inspectjson-go/inspectjson"
	"github.com/sanity-io/litter"
)

var reTextOffsetTerse = regexp.MustCompile(`(cursorio\.TextOffset\{)\s*(Byte: \d+,)\s*(Line: \d+,)\s*(LineColumn: \d+),\s*(})`)

func main() {
	value, err := inspectjson.Parse(
		os.Stdin,
		inspectjson.TokenizerOptions{}.
			SourceOffsets(true).
			Lax(true),
	)
	if err != nil {
		panic(err)
	}

	fmt.Fprintf(os.Stdout, "%s\n", reTextOffsetTerse.ReplaceAllString(litter.Sdump(value), "$1$2 $3 $4$5"))
}
