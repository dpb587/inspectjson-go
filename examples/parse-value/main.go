package main

import (
	"fmt"
	"os"
	"regexp"

	"github.com/dpb587/inspectjson-go/inspectjson"
	"github.com/sanity-io/litter"
)

var reTextOffsetTerse = regexp.MustCompile(`(cursorio\.TextOffset\{)\s*(Byte: \d+,)\s*(LineColumn: cursorio.TextLineColumn{)\s*(\d+),\s*(\d+),\s*(\}),\s*(})`)

func main() {
	value, err := inspectjson.Parse(
		os.Stdin,
		inspectjson.TokenizerConfig{}.
			SetSourceOffsets(true).
			SetLax(true),
	)
	if err != nil {
		panic(err)
	}

	fmt.Fprintf(os.Stdout, "%s\n", reTextOffsetTerse.ReplaceAllString(litter.Sdump(value), "$1$2 $3$4, $5$6$7"))
}
