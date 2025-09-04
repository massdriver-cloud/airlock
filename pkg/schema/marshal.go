package schema

import "encoding/json"

func (s *Schema) MarshalJSON() ([]byte, error) {
	if s.AdditionalProperties != nil {
		addPropBytes, err := json.Marshal(s.AdditionalProperties)
		if err != nil {
			return nil, err
		}
		addPropRaw := json.RawMessage(addPropBytes)
		s.AdditionalPropertiesRaw = &addPropRaw
	}

	if s.Dependencies != nil {
		dependenciesBytes, err := json.Marshal(s.Dependencies)
		if err != nil {
			return nil, err
		}
		dependenciesRaw := json.RawMessage(dependenciesBytes)
		s.DependenciesRaw = &dependenciesRaw
	}

	type Alias Schema
	return json.Marshal(&struct {
		*Alias
	}{
		Alias: (*Alias)(s),
	})
}

// need a custom unmarshaler to deal with the ambiguity of additionalProperties
func (s *Schema) UnmarshalJSON(data []byte) error {
	// we need to redirect to another type to avoid infinite recursion loop
	type Alias Schema
	alias := &struct {
		*Alias
	}{
		Alias: (*Alias)(s),
	}

	if err := json.Unmarshal(data, &alias); err != nil {
		return err
	}

	if s.AdditionalPropertiesRaw != nil {
		var addPropBool bool
		err := json.Unmarshal(*s.AdditionalPropertiesRaw, &addPropBool)
		if err != nil {
			// additionalProperties is schema
			var addPropSchema Schema
			err = json.Unmarshal(*s.AdditionalPropertiesRaw, &addPropSchema)
			if err != nil {
				return err
			}
			s.AdditionalProperties = &addPropSchema
		} else {
			// additionalProperties is bool
			s.AdditionalProperties = addPropBool
		}
		s.AdditionalPropertiesRaw = nil
	}

	if s.DependenciesRaw != nil {
		var dependendentRequired map[string][]string
		err := json.Unmarshal(*s.DependenciesRaw, &dependendentRequired)
		if err != nil {
			// dependencies is map, try to unmarshal as map[string]*Schema
			var dependentSchema map[string]*Schema
			err = json.Unmarshal(*s.DependenciesRaw, &dependentSchema)
			if err != nil {
				return err
			}
			s.Dependencies = dependentSchema
		} else {
			// dependencies is map[string][]string
			s.Dependencies = dependendentRequired
		}
	}

	return nil
}
