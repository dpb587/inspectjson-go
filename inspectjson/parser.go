package inspectjson

import (
	"errors"
	"fmt"
	"io"
	"strconv"
)

func Parse(r io.Reader, opts ...ParserOption) (Value, error) {
	return newParser(r, false, opts...).parse()
}

type parser struct {
	t *Tokenizer

	keepReplacedObjectMembers bool
}

func newParser(r io.Reader, multi bool, opts ...ParserOption) *parser {
	var topts []TokenizerOption
	var popts []ParserOption

	for _, opt := range opts {
		if opt == nil {
			continue
		} else if topt, ok := opt.(TokenizerOption); ok {
			topts = append(topts, topt)
		} else {
			popts = append(popts, opt)
		}
	}

	if multi {
		topts = append(topts, TokenizerConfig{}.SetMultistream(true))
	}

	p := &parser{
		t: NewTokenizer(r, topts...),
	}

	for _, opt := range popts {
		opt.applyParser(p)
	}

	return p
}

type parseStackMode int

const (
	parseStackModeRoot parseStackMode = iota
	parseStackModeArray
	parseStackModeObjectName
	parseStackModeObjectValue
)

func (p *parser) parse() (Value, error) {
	var rootValue Value
	var stackValues []Value
	var stackMode = []parseStackMode{parseStackModeRoot}

	for {
		token, err := p.t.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}

			return nil, err
		}

		var completedValue Value

		switch tokenType := token.(type) {
		case BeginArrayToken:
			stackMode = append(stackMode, parseStackModeArray)
			stackValues = append(stackValues, ArrayValue{
				BeginToken: tokenType,
			})

			continue
		case BeginObjectToken:
			stackMode = append(stackMode, parseStackModeObjectName)
			stackValues = append(stackValues, ObjectValue{
				BeginToken: tokenType,
				Members:    map[string]ObjectMember{},
			})

			continue
		case EndArrayToken:
			completedArrayValue := stackValues[len(stackValues)-1].(ArrayValue)
			completedArrayValue.EndToken = tokenType

			stackMode = stackMode[0 : len(stackMode)-1]
			stackValues = stackValues[0 : len(stackValues)-1]
			completedValue = completedArrayValue
		case EndObjectToken:
			completedObjectValue := stackValues[len(stackValues)-1].(ObjectValue)
			completedObjectValue.EndToken = tokenType

			stackMode = stackMode[0 : len(stackMode)-1]
			stackValues = stackValues[0 : len(stackValues)-1]
			completedValue = completedObjectValue
		case StringToken:
			completedValue = StringValue{
				SourceOffsets: tokenType.SourceOffsets,
				Value:         tokenType.Content,
			}
		case NumberToken:
			valueFloat64, err := strconv.ParseFloat(tokenType.Content, 64)
			if err != nil {
				return nil, fmt.Errorf("parse number (float): %v", err)
			}

			completedValue = NumberValue{
				SourceOffsets: tokenType.SourceOffsets,
				Value:         valueFloat64,
			}
		case TrueToken:
			completedValue = BooleanValue{
				SourceOffsets: tokenType.SourceOffsets,
				Value:         true,
			}
		case FalseToken:
			completedValue = BooleanValue{
				SourceOffsets: tokenType.SourceOffsets,
				Value:         false,
			}
		case NullToken:
			completedValue = NullValue{
				SourceOffsets: tokenType.SourceOffsets,
			}
		default:
			continue
		}

		switch stackMode[len(stackMode)-1] {
		case parseStackModeArray:
			parentArrayValue := stackValues[len(stackValues)-1].(ArrayValue)
			parentArrayValue.Values = append(parentArrayValue.Values, completedValue)
			stackValues[len(stackValues)-1] = parentArrayValue
		case parseStackModeObjectName:
			stackValues = append(stackValues, completedValue)
			stackMode[len(stackMode)-1] = parseStackModeObjectValue
		case parseStackModeObjectValue:
			objectMemberValue := ObjectMember{
				Name:  stackValues[len(stackValues)-1].(StringValue),
				Value: completedValue,
			}

			stackValues = stackValues[0 : len(stackValues)-1]
			stackMode[len(stackMode)-1] = parseStackModeObjectName

			parentObjectValue := stackValues[len(stackValues)-1].(ObjectValue)

			if existingMember, exists := parentObjectValue.Members[objectMemberValue.Name.Value]; exists {
				if p.keepReplacedObjectMembers {
					parentObjectValue.ReplacedMembers = append(parentObjectValue.ReplacedMembers, existingMember)
				}
			}

			parentObjectValue.Members[objectMemberValue.Name.Value] = objectMemberValue

			stackValues[len(stackValues)-1] = parentObjectValue
		case parseStackModeRoot:
			rootValue = completedValue
		}
	}

	return rootValue, nil
}
