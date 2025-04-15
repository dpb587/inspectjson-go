package testsuite_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dpb587/cursorio-go/cursorio"
	"github.com/dpb587/inspectjson-go/inspectjson"
)

func Test(t *testing.T) {
	testdata := requireTestdataArchive(t)

	for {
		header, err := testdata.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			t.Fatalf("read test archive: %v", err)
		} else if ok, _ := filepath.Match("./test_parsing/*.json", header.Name); !ok {
			continue
		}

		nameBase := filepath.Base(header.Name)

		var wrapName string

		if strings.HasPrefix(nameBase, "n_") {
			wrapName = "InvalidSyntax"
		} else if strings.HasPrefix(nameBase, "y_") {
			wrapName = "ValidSyntax"
		} else if strings.HasPrefix(nameBase, "i_") {
			wrapName = "AmbiguousSyntax"
		} else {
			t.Fatalf("unknown test name: %s", nameBase)
		}

		t.Run(wrapName+"/"+nameBase, func(t *testing.T) {
			tokenizer := inspectjson.NewTokenizer(testdata)
			for {
				_, err := tokenizer.Next()
				if err != nil {
					if errors.Is(err, io.EOF) {
						break
					} else if !strings.HasPrefix(nameBase, "y_") {
						t.Logf("error: %v", err.Error())

						return
					}

					t.Fatalf("error: %v", err)
				}
			}

			if strings.HasPrefix(nameBase, "n_") {
				t.Fatal("error expected, but got nil")
			}
		})
	}
}

func TestFuzz(t *testing.T) {
	testdata := requireTestdataArchive(t)

	for {
		header, err := testdata.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			t.Fatalf("read test archive: %v", err)
		} else if ok, _ := filepath.Match("./test_parsing/*.json", header.Name); !ok {
			continue
		}

		nameBase := filepath.Base(header.Name)

		buf, err := io.ReadAll(testdata)
		if err != nil {
			t.Fatalf("read test archive: %v", err)
		}

		if strings.HasPrefix(nameBase, "n_") {
			t.Run("_FuzzByteOffset/"+nameBase, func(t *testing.T) {
				var baseOffsetErr cursorio.OffsetError

				_, baseErr := inspectjson.Parse(bytes.NewReader(buf))
				if baseErr == nil {
					t.Fatal("error expected, but got nil")
				} else if !errors.As(baseErr, &baseOffsetErr) {
					return
				}

				var textOffsetErr cursorio.OffsetError

				_, sourceErr := inspectjson.Parse(bytes.NewReader(buf), inspectjson.TokenizerOptions{}.SourceOffsets(true))
				if sourceErr == nil {
					t.Fatal("offsets error expected, but got nil")
				} else if !errors.As(sourceErr, &textOffsetErr) {
					t.Fatalf("offset error expected, but got: %v", sourceErr)
				}

				byteOffset := baseOffsetErr.Offset.(cursorio.ByteOffset)
				textOffset := textOffsetErr.Offset.(cursorio.TextOffset)

				if byteOffset != cursorio.ByteOffset(textOffset.Byte) {
					t.Fatalf("raw offset (byte=%d) does not match source offsets (byte=%d)", byteOffset, textOffset.Byte)
				}
			})
		} else if strings.HasPrefix(nameBase, "y_") {
			t.Run("_FuzzMarshal/"+nameBase, func(t *testing.T) {
				var stdValue any

				err := json.Unmarshal(buf, &stdValue)
				if err != nil {
					t.Fatalf("error: std: %v", err)
				}

				value, err := inspectjson.Parse(bytes.NewReader(buf))
				if err != nil {
					t.Fatalf("error: inspectjson: parse: %v", err)
				}

				stdBytes, err := json.Marshal(stdValue)
				if err != nil {
					t.Fatalf("error: std: marshal: %v", err)
				}

				valueBytes, err := json.Marshal(value.AsBuiltin())
				if err != nil {
					t.Fatalf("error: inspectjson: parse: %v", err)
				}

				if !bytes.Equal(stdBytes, valueBytes) {
					t.Log("error: inspectjson and std do not match")
					t.Log("")
					t.Logf("inspectjson: %s\n", string(valueBytes))
					t.Logf("    std: %s\n", string(stdBytes))

					t.FailNow()
				}
			})
		}
	}
}

func requireTestdataArchive(t *testing.T) *tar.Reader {
	osFile, err := os.Open("testdata.tar.gz")
	if err != nil {
		t.Fatalf("open test archive: %v", err)
	}

	t.Cleanup(func() {
		osFile.Close()
	})

	gzReader, err := gzip.NewReader(osFile)
	if err != nil {
		t.Fatalf("open test archive: %v", err)
	}

	t.Cleanup(func() {
		gzReader.Close()
	})

	tarReader := tar.NewReader(gzReader)
	if err != nil {
		t.Fatalf("open test archive: %v", err)
	}

	return tarReader
}
