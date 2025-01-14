package json

import "go.opentelemetry.io/otel/attribute"

func attrSliceToMap(attributes []attribute.KeyValue) map[string]any {
	if len(attributes) == 0 {
		return nil
	}
	attrs := make(map[string]any, len(attributes))
	for _, kv := range attributes {
		attrs[string(kv.Key)] = kv.Value.AsInterface()
	}
	return attrs
}

func attrSetToMap(attributes attribute.Set) map[string]any {
	return attrSliceToMap(attributes.ToSlice())
}
