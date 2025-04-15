package inspectjson

import "github.com/dpb587/cursorio-go/cursorio"

type TokenizerOptionsApplier interface {
	applyTokenizer(t *tokenizerOptions)
}

type TokenizerOptions struct {
	sourceOffsets       *bool
	sourceInitialOffset *cursorio.TextOffset
	emitWhitespace      *bool
	laxBehaviors        map[SyntaxBehavior]bool
	syntaxRecoveryHook  SyntaxRecoveryHookFunc
	multistream         *bool
}

func (o TokenizerOptions) applyParser(p *parser) {
	panic("should not be called directly")
}

func (o TokenizerOptions) applyTokenizer(t *tokenizerOptions) {
	if o.sourceOffsets != nil {
		t.sourceOffsets = *o.sourceOffsets
	}

	if o.sourceInitialOffset != nil {
		t.sourceInitialOffset = *o.sourceInitialOffset
	}

	if o.emitWhitespace != nil {
		t.emitWhitespace = *o.emitWhitespace
	}

	for behavior, enable := range o.laxBehaviors {
		if enable {
			t.laxBehaviors = t.laxBehaviors | behavior
		} else {
			t.laxBehaviors = t.laxBehaviors &^ behavior
		}
	}

	if o.syntaxRecoveryHook != nil {
		t.laxListener = o.syntaxRecoveryHook
	}

	if o.multistream != nil {
		t.multistream = *o.multistream
	}
}

func (o TokenizerOptions) EmitWhitespace(enable bool) TokenizerOptions {
	o.emitWhitespace = &enable

	return o
}

func (o TokenizerOptions) Lax(enable bool) TokenizerOptions {
	o.laxBehaviors = map[SyntaxBehavior]bool{
		LaxIgnoreBlockComment:        enable,
		LaxIgnoreLineComment:         enable,
		LaxStringEscapeInvalidEscape: enable,
		LaxStringEscapeMissingEscape: enable,
		LaxNumberTrimLeadingZero:     enable,
		LaxLiteralCaseInsensitive:    enable,
		LaxIgnoreExtraComma:          enable,
		LaxIgnoreTrailingSemicolon:   enable,
	}

	return o
}

func (o TokenizerOptions) LaxBehavior(behavior SyntaxBehavior, enable bool) TokenizerOptions {
	if o.laxBehaviors == nil {
		o.laxBehaviors = map[SyntaxBehavior]bool{}
	}

	o.laxBehaviors[behavior] = enable

	return o
}

func (o TokenizerOptions) SyntaxRecoveryHook(hook SyntaxRecoveryHookFunc) TokenizerOptions {
	o.syntaxRecoveryHook = hook

	return o
}

func (o TokenizerOptions) Multistream(enable bool) TokenizerOptions {
	o.multistream = &enable

	return o
}

func (o TokenizerOptions) SourceOffsets(enable bool) TokenizerOptions {
	o.sourceOffsets = &enable

	return o
}

func (o TokenizerOptions) SourceInitialOffset(offset cursorio.TextOffset) TokenizerOptions {
	v := true

	o.sourceOffsets = &v
	o.sourceInitialOffset = &offset

	return o
}
