package inspectjson

type GrammarName string

const (
	grammarName_BeginArray     GrammarName = "begin-array"
	grammarName_BeginObject    GrammarName = "begin-object"
	grammarName_EndArray       GrammarName = "end-array"
	grammarName_EndObject      GrammarName = "end-object"
	grammarName_NameSeparator  GrammarName = "name-separator"
	grammarName_ValueSeparator GrammarName = "value-separator"
	grammarName_False          GrammarName = "false"
	grammarName_Null           GrammarName = "null"
	grammarName_True           GrammarName = "true"
	grammarName_String         GrammarName = "string"
	grammarName_Number         GrammarName = "number"
	grammarName_Ws             GrammarName = "ws"

	grammarName_Boolean GrammarName = "boolean"
	grammarName_Object  GrammarName = "object"
	grammarName_Array   GrammarName = "array"
)
