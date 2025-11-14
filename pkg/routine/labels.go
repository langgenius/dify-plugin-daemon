package routine

type RoutineLabelKey string

type Labels map[RoutineLabelKey]string

const (
	RoutineLabelKeyModule    RoutineLabelKey = "module"
	RoutineLabelKeyFunction  RoutineLabelKey = "function"
	RoutineLabelKeyAction    RoutineLabelKey = "action"
	RoutineLabelKeyType      RoutineLabelKey = "type"
	RoutineLabelKeyMethod    RoutineLabelKey = "method"
	RoutineLabelKeySessionID RoutineLabelKey = "session_id"
	RoutineLabelKeyTarget    RoutineLabelKey = "target"
)
