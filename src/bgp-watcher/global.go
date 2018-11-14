package main

//
type AlertPriority int

//
const (
	PriorityHigh   AlertPriority = 1
	PriorityMedium AlertPriority = 2
	PriorityLow    AlertPriority = 3
)

//
func (ap AlertPriority) String() string {

	switch ap {
	case PriorityHigh:
		return "High"
	case PriorityMedium:
		return "Medium"
	case PriorityLow:
		return "Low"
	default:
		return "Unknown"
	}
}
