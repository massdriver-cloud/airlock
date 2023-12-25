package schema

import (
	"slices"

	orderedmap "github.com/wk8/go-ordered-map/v2"
)

// Walk an object schema and flatten all the potential properties from oneOf/anyOf/allOf, dependencies
func ExpandProperties(schema *Schema) *orderedmap.OrderedMap[string, *Schema] {
	result := schema.Properties

	if result == nil {
		result = orderedmap.New[string, *Schema]()
	}

	for _, item := range schema.OneOf {
		oneOf := ExpandProperties(item)
		result = mergeProperties(result, oneOf)
	}
	for _, item := range schema.AnyOf {
		anyOf := ExpandProperties(item)
		result = mergeProperties(result, anyOf)
	}
	for _, item := range schema.AllOf {
		allOf := ExpandProperties(item)
		result = mergeProperties(result, allOf)
	}

	// dependencies
	for _, item := range schema.Dependencies {
		dep := ExpandProperties(item)
		result = mergeProperties(result, dep)
	}

	return result
}

// Merge two set of properties together. If the same properties exist in both base and combine, base is preserved and combine is ignored
func mergeProperties(base, combine *orderedmap.OrderedMap[string, *Schema]) *orderedmap.OrderedMap[string, *Schema] {
	merged := orderedmap.New[string, *Schema]()
	existing := []string{}

	for pair := base.Oldest(); pair != nil; pair = pair.Next() {
		existing = append(existing, pair.Key)
		merged.AddPairs(*pair)
	}

	for pair := combine.Oldest(); pair != nil; pair = pair.Next() {
		if !slices.Contains(existing, pair.Key) {
			merged.AddPairs(*pair)
		}
	}

	return merged
}
