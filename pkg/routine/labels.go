package routine

type RoutineLabelKey string

type Labels map[RoutineLabelKey]string

const (
	RoutineLabelKeyModule      RoutineLabelKey = "module"
	RoutineLabelRuntimeKeyType RoutineLabelKey = "type"
	RoutineLabelKeyMethod      RoutineLabelKey = "method"
	RoutineLabelKeySessionID   RoutineLabelKey = "session_id"
	RoutineLabelKeyLambdaURL   RoutineLabelKey = "target"
)
