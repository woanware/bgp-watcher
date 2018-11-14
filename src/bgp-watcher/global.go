package main

//
type AlertPriority int

//
const (
	Priority1 AlertPriority = 1
	Priority2 AlertPriority = 2
	Priority3 AlertPriority = 3
)

//
func (ap AlertPriority) String() string {
	switch ap {
	case Priority1:
		return "High"
	case Priority2:
		return "Medium"
	case Priority3:
		return "Low"
	default:
		return "Unknown"
	}
}
