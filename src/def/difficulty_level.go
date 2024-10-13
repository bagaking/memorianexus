package def

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/khicago/irr"
)

type DifficultyLevel uint8

const (
	NoviceNormal    DifficultyLevel = 0x01
	NoviceAdvanced  DifficultyLevel = 0x02
	NoviceChallenge DifficultyLevel = 0x03

	AmateurNormal    DifficultyLevel = 0x11
	AmateurAdvanced  DifficultyLevel = 0x12
	AmateurChallenge DifficultyLevel = 0x13

	ProfessionalNormal    DifficultyLevel = 0x21
	ProfessionalAdvanced  DifficultyLevel = 0x22
	ProfessionalChallenge DifficultyLevel = 0x23

	ExpertNormal    DifficultyLevel = 0x31
	ExpertAdvanced  DifficultyLevel = 0x32
	ExpertChallenge DifficultyLevel = 0x33

	MasterNormal    DifficultyLevel = 0x41
	MasterAdvanced  DifficultyLevel = 0x42
	MasterChallenge DifficultyLevel = 0x43
	MasterExtreme   DifficultyLevel = 0x44
)

// Normalize returns a normalized value between 0 and 1
func (d DifficultyLevel) Normalize() float64 {
	maxValue := float64(MasterExtreme)
	return float64(d) / maxValue
}

var difficultyFactors = map[DifficultyLevel]float64{
	NoviceNormal:          1.1,
	NoviceAdvanced:        1.15,
	NoviceChallenge:       1.2,
	AmateurNormal:         1.25,
	AmateurAdvanced:       1.3,
	AmateurChallenge:      1.35,
	ProfessionalNormal:    1.4,
	ProfessionalAdvanced:  1.45,
	ProfessionalChallenge: 1.5,
	ExpertNormal:          1.55,
	ExpertAdvanced:        1.6,
	ExpertChallenge:       1.65,
	MasterNormal:          1.7,
	MasterAdvanced:        1.75,
	MasterChallenge:       1.8,
	MasterExtreme:         1.85,
}

func (d DifficultyLevel) Factor() float64 {
	if factor, exists := difficultyFactors[d]; exists {
		return factor
	}
	return 1.0 // 默认值
}

var difficultyLevelNames = map[DifficultyLevel]string{
	NoviceNormal:          "novice_normal",
	NoviceAdvanced:        "novice_advanced",
	NoviceChallenge:       "novice_challenge",
	AmateurNormal:         "amateur_normal",
	AmateurAdvanced:       "amateur_advanced",
	AmateurChallenge:      "amateur_challenge",
	ProfessionalNormal:    "professional_normal",
	ProfessionalAdvanced:  "professional_advanced",
	ProfessionalChallenge: "professional_challenge",
	ExpertNormal:          "expert_normal",
	ExpertAdvanced:        "expert_advanced",
	ExpertChallenge:       "expert_challenge",
	MasterNormal:          "master_normal",
	MasterAdvanced:        "master_advanced",
	MasterChallenge:       "master_challenge",
	MasterExtreme:         "master_extreme",
}

func (d *DifficultyLevel) String() string {
	return difficultyLevelNames[*d]
}

func (d DifficultyLevel) Valid() bool {
	_, exists := difficultyLevelNames[d]
	return exists
}

// UnmarshalJSON unmarshal the enum from a json string or number
func (d *DifficultyLevel) UnmarshalJSON(data []byte) error {
	var value interface{}
	if err := jsoniter.Unmarshal(data, &value); err != nil {
		return err
	}

	switch v := value.(type) {
	case float64:
		rawLevel, ok := parseUint8JSONNumber(v)
		if !ok {
			return irr.Error("invalid DifficultyLevel value")
		}
		level := DifficultyLevel(rawLevel)
		if !level.Valid() {
			return irr.Error("invalid DifficultyLevel value")
		}
		*d = level
	case string:
		for key, name := range difficultyLevelNames {
			if name == v {
				*d = key
				return nil
			}
		}
		return irr.Error("invalid DifficultyLevel name")
	default:
		return irr.Error("invalid type for DifficultyLevel")
	}
	return nil
}
