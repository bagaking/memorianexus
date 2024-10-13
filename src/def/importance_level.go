package def

import (
	"encoding/json"
	"errors"
)

type ImportanceLevel uint8

const (
	DomainGeneral   ImportanceLevel = 0x01
	DomainKey       ImportanceLevel = 0x02
	DomainEssential ImportanceLevel = 0x03

	AreaGeneral     ImportanceLevel = 0x11
	AreaKey         ImportanceLevel = 0x12
	AreaEssential   ImportanceLevel = 0x13
	AreaMasterPiece ImportanceLevel = 0x14

	GlobalGeneral     ImportanceLevel = 0x21
	GlobalKey         ImportanceLevel = 0x22
	GlobalEssential   ImportanceLevel = 0x23
	GlobalMasterPiece ImportanceLevel = 0x24
)

var importanceLevelNames = map[ImportanceLevel]string{
	DomainGeneral:     "domain_general",
	DomainKey:         "domain_key",
	DomainEssential:   "domain_essential",
	AreaGeneral:       "area_general",
	AreaKey:           "area_key",
	AreaEssential:     "area_essential",
	AreaMasterPiece:   "area_master_piece",
	GlobalGeneral:     "global_general",
	GlobalKey:         "global_key",
	GlobalEssential:   "global_essential",
	GlobalMasterPiece: "global_master_piece",
}

func (i *ImportanceLevel) String() string {
	return importanceLevelNames[*i]
}

func (i ImportanceLevel) Valid() bool {
	_, exists := importanceLevelNames[i]
	return exists
}

// UnmarshalJSON unmarshal the enum from a json string or number
func (i *ImportanceLevel) UnmarshalJSON(data []byte) error {
	var value interface{}
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}

	switch v := value.(type) {
	case float64:
		rawLevel, ok := parseUint8JSONNumber(v)
		if !ok {
			return errors.New("invalid ImportanceLevel value")
		}
		level := ImportanceLevel(rawLevel)
		if !level.Valid() {
			return errors.New("invalid ImportanceLevel value")
		}
		*i = level
	case string:
		for key, name := range importanceLevelNames {
			if name == v {
				*i = key
				return nil
			}
		}
		return errors.New("invalid ImportanceLevel name")
	default:
		return errors.New("invalid type for ImportanceLevel")
	}
	return nil
}

// Normalize returns a normalized value between 0 and 1
func (i ImportanceLevel) Normalize() float64 {
	maxValue := float64(GlobalMasterPiece)
	return float64(i) / maxValue
}
