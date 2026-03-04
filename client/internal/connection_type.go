package internal

type ConnectionType int

const (
	ConnectionTypeNone ConnectionType = iota
	ConnectionTypeNetBird
	ConnectionTypeV2Ray
)

func (ct ConnectionType) String() string {
	switch ct {
	case ConnectionTypeNetBird:
		return "NetBird"
	case ConnectionTypeV2Ray:
		return "V2Ray"
	case ConnectionTypeNone:
		return "None"
	default:
		return "Unknown"
	}
}
