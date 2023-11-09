package docgen

import "gopkg.in/yaml.v3"

// Custom Unmarshal for fields that can either include be a string or an array of strings
// https://github.com/go-yaml/yaml/issues/100#issuecomment-901604971

type StringArray []string

func (a *StringArray) UnmarshalYAML(value *yaml.Node) error {
	var multi []string
	err := value.Decode(&multi)
	if err != nil {
		var single string
		err := value.Decode(&single)
		if err != nil {
			return err
		}
		*a = []string{single}
	} else {
		*a = multi
	}
	return nil
}
