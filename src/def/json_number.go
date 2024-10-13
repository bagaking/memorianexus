package def

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"strconv"
)

func decodeJSONValue(data []byte) (any, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()

	var value any
	if err := decoder.Decode(&value); err != nil {
		return nil, err
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if err == nil {
			return nil, errors.New("invalid JSON value")
		}
		return nil, err
	}
	return value, nil
}

func parseUint8JSONNumber(v json.Number) (uint8, bool) {
	parsed, err := strconv.ParseUint(v.String(), 10, 8)
	if err != nil {
		return 0, false
	}
	return uint8(parsed), true
}
