inspectjson.ObjectValue{
  BeginToken: inspectjson.BeginObjectToken{
    SourceOffsets: &cursorio.TextOffsetRange{
      From: cursorio.TextOffset{Byte: 0, LineColumn: cursorio.TextLineColumn{0, 0}},
      Until: cursorio.TextOffset{Byte: 1, LineColumn: cursorio.TextLineColumn{0, 1}},
    },
  },
  EndToken: inspectjson.EndObjectToken{
    SourceOffsets: &cursorio.TextOffsetRange{
      From: cursorio.TextOffset{Byte: 9, LineColumn: cursorio.TextLineColumn{0, 9}},
      Until: cursorio.TextOffset{Byte: 10, LineColumn: cursorio.TextLineColumn{0, 10}},
    },
  },
  Members: map[string]inspectjson.ObjectMember{
    "n": inspectjson.ObjectMember{
      Name: inspectjson.StringValue{
        SourceOffsets: &cursorio.TextOffsetRange{
          From: cursorio.TextOffset{Byte: 1, LineColumn: cursorio.TextLineColumn{0, 1}},
          Until: cursorio.TextOffset{Byte: 4, LineColumn: cursorio.TextLineColumn{0, 4}},
        },
        Value: "n",
      },
      Value: inspectjson.BooleanValue{
        SourceOffsets: &cursorio.TextOffsetRange{
          From: cursorio.TextOffset{Byte: 5, LineColumn: cursorio.TextLineColumn{0, 5}},
          Until: cursorio.TextOffset{Byte: 9, LineColumn: cursorio.TextLineColumn{0, 9}},
        },
        Value: true,
      },
    },
  },
  ReplacedMembers: nil,
}
