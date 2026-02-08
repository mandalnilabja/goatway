package types

import "encoding/json"

// marshalStringArray marshals a string slice to JSON.
func marshalStringArray(values []string) ([]byte, error) {
	return json.Marshal(values)
}

// unmarshalString attempts to unmarshal data as a string.
func unmarshalString(data []byte, s *string) error {
	return json.Unmarshal(data, s)
}

// unmarshalStringArray attempts to unmarshal data as a string array.
func unmarshalStringArray(data []byte, values *[]string) error {
	return json.Unmarshal(data, values)
}
