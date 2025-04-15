package inspectjson

import "github.com/dpb587/cursorio-go/cursorio"

type Value interface {
	GetGrammarName() GrammarName
	GetSourceOffsets() *cursorio.TextOffsetRange

	AsBuiltin() any
}

//

type ObjectValue struct {
	BeginToken BeginObjectToken
	EndToken   EndObjectToken

	Members         map[string]ObjectMember
	ReplacedMembers []ObjectMember
}

var _ Value = ObjectValue{}

func (v ObjectValue) GetGrammarName() GrammarName {
	return grammarName_Object
}

func (v ObjectValue) GetSourceOffsets() *cursorio.TextOffsetRange {
	if v.BeginToken.SourceOffsets == nil || v.EndToken.SourceOffsets == nil {
		return nil
	}

	return &cursorio.TextOffsetRange{
		From:  v.BeginToken.SourceOffsets.From,
		Until: v.EndToken.SourceOffsets.Until,
	}
}

func (v ObjectValue) AsBuiltin() any {
	p := map[string]any{}

	for k, member := range v.Members {
		p[k] = member.Value.AsBuiltin()
	}

	return p
}

//

type ObjectMember struct {
	Name  StringValue
	Value Value
}

//

type ArrayValue struct {
	BeginToken BeginArrayToken
	EndToken   EndArrayToken

	Values []Value
}

var _ Value = ArrayValue{}

func (v ArrayValue) GetGrammarName() GrammarName {
	return grammarName_Array
}

func (v ArrayValue) GetSourceOffsets() *cursorio.TextOffsetRange {
	if v.BeginToken.SourceOffsets == nil || v.EndToken.SourceOffsets == nil {
		return nil
	}

	return &cursorio.TextOffsetRange{
		From:  v.BeginToken.SourceOffsets.From,
		Until: v.EndToken.SourceOffsets.Until,
	}
}

func (v ArrayValue) AsBuiltin() any {
	p := make([]any, len(v.Values))

	for i, value := range v.Values {
		p[i] = value.AsBuiltin()
	}

	return p
}

//

type BooleanValue struct {
	SourceOffsets *cursorio.TextOffsetRange
	Value         bool
}

var _ Value = BooleanValue{}

func (v BooleanValue) GetGrammarName() GrammarName {
	return grammarName_Boolean
}

func (v BooleanValue) GetSourceOffsets() *cursorio.TextOffsetRange {
	return v.SourceOffsets
}

func (v BooleanValue) AsBuiltin() any {
	return v.Value
}

//

type NullValue struct {
	SourceOffsets *cursorio.TextOffsetRange
}

var _ Value = NullValue{}

func (v NullValue) GetGrammarName() GrammarName {
	return grammarName_Null
}

func (v NullValue) GetSourceOffsets() *cursorio.TextOffsetRange {
	return v.SourceOffsets
}

func (v NullValue) AsBuiltin() any {
	return nil
}

//

type NumberValue struct {
	SourceOffsets *cursorio.TextOffsetRange
	Value         float64
}

var _ Value = NumberValue{}

func (v NumberValue) GetGrammarName() GrammarName {
	return grammarName_Number
}

func (v NumberValue) GetSourceOffsets() *cursorio.TextOffsetRange {
	return v.SourceOffsets
}

func (v NumberValue) AsBuiltin() any {
	return v.Value
}

//

type StringValue struct {
	SourceOffsets *cursorio.TextOffsetRange
	Value         string
}

var _ Value = StringValue{}

func (v StringValue) GetGrammarName() GrammarName {
	return grammarName_String
}

func (v StringValue) GetSourceOffsets() *cursorio.TextOffsetRange {
	return v.SourceOffsets
}

func (v StringValue) AsBuiltin() any {
	return v.Value
}
