package services

import "encoding/json"

func mustMarshal(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		return []byte("null")
	}
	return b
}

func jsonUnmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, v)
}

func jsonMarshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

func fromJSONMap(j []byte) map[string]interface{} {
	if len(j) == 0 || string(j) == "null" {
		return nil
	}
	var out map[string]interface{}
	_ = json.Unmarshal(j, &out)
	return out
}
