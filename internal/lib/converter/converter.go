package converter

import (
	"uuid"

	"google.golang.org/protobuf/types/known/timestamppb"
)

//kanna:converter
var Defaults = []any{
	uuid.UUID.String,
	uuid.Parse,
	timestamppb.New,
	(*timestamppb.Timestamp).AsTime,
}
