package def

const maxUint8JSONNumber = float64(^uint8(0))

func parseUint8JSONNumber(v float64) (uint8, bool) {
	if v < 0 || v > maxUint8JSONNumber {
		return 0, false
	}
	parsed := uint8(v)
	if float64(parsed) != v {
		return 0, false
	}
	return parsed, true
}
