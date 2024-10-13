package def

import (
	"encoding/json"
	"fmt"
	"strings"
)

type DungeonType uint8 // 0 ~ 255

const (
	DungeonTypeCampaign DungeonType = 0x1  // 战役地牢
	DungeonTypeEndless  DungeonType = 0x2  // 无尽地牢
	DungeonTypeInstance DungeonType = 0x21 // 即时副本 (随机地牢)
)

func (dt *DungeonType) String() string {
	switch {
	case *dt == DungeonTypeCampaign:
		return "campaign"
	case *dt == DungeonTypeEndless:
		return "endless"
	case *dt >= DungeonTypeInstance:
		return "instance"
	default:
		return "unknown"
	}
}

func (dt *DungeonType) Valid() bool {
	switch *dt {
	case DungeonTypeCampaign:
	case DungeonTypeEndless:
	case DungeonTypeInstance:
	default:
		return false
	}
	return true
}

// UnmarshalJSON custom unmarshaller to handle both strings and numbers
func (dt *DungeonType) UnmarshalJSON(data []byte) error {
	value, err := decodeJSONValue(data)
	if err != nil {
		return err
	}

	switch v := value.(type) {
	case json.Number:
		rawType, ok := parseUint8JSONNumber(v)
		if !ok {
			return fmt.Errorf("invalid dungeon type: %s", v.String())
		}
		dungeonType := DungeonType(rawType)
		if !dungeonType.Valid() {
			return fmt.Errorf("invalid dungeon type: %s", v.String())
		}
		*dt = dungeonType
	case string:
		switch strings.TrimSpace(strings.ToLower(v)) {
		case "campaign", "1":
			*dt = DungeonTypeCampaign
		case "endless", "2":
			*dt = DungeonTypeEndless
		case "instance", "3":
			*dt = DungeonTypeInstance
		default:
			return fmt.Errorf("invalid dungeon type: %s", v)
		}
	default:
		return fmt.Errorf("invalid dungeon type: %v", v)
	}

	return nil
}
