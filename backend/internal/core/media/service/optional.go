package service

import (
	"bytes"
	"encoding/json"
)

// Optional distinguishes an omitted JSON field (Set=false) from an explicit
// null (Set=true, Value=nil), so PATCH requests can clear nullable fields.
type Optional[T any] struct {
	Set   bool
	Value *T
}

func Some[T any](v T) Optional[T] { return Optional[T]{Set: true, Value: &v} }

func Null[T any]() Optional[T] { return Optional[T]{Set: true} }

func (o *Optional[T]) UnmarshalJSON(data []byte) error {
	o.Set = true
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		o.Value = nil
		return nil
	}
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	o.Value = &v
	return nil
}
