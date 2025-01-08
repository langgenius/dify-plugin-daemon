package serverless

type RunnerInstance struct {
	ID           string `json:"ID" validate:"required"`
	Name         string `json:"Name" validate:"required"`
	Endpoint     string `json:"Endpoint" validate:"required"`
	ResourceName string `json:"ResourceName" validate:"required"`
	Status       struct {
		State string `json:"State" validate:"required"`
	} `json:"Status" validate:"required"`
}

type RunnerInstances struct {
	Error string           `json:"error"`
	Items []RunnerInstance `json:"Items"`
}

type LaunchStage string

const (
	LAUNCH_STAGE_START LaunchStage = "start"
	LAUNCH_STAGE_BUILD LaunchStage = "build"
	LAUNCH_STAGE_RUN   LaunchStage = "run"
	LAUNCH_STAGE_END   LaunchStage = "end"
)

type LaunchState string

const (
	LAUNCH_STATE_SUCCESS LaunchState = "success"
	LAUNCH_STATE_RUNNING LaunchState = "running"
	LAUNCH_STATE_FAILED  LaunchState = "failed"
)

type LaunchFunctionResponseChunk struct {
	Stage   LaunchStage `json:"Stage"`
	Obj     string      `json:"Obj"`
	State   LaunchState `json:"State"`
	Message string      `json:"Message"`
}

type LaunchFunctionFinalStageMessage struct {
	Endpoint string `comma:"endpoint"`
	Name     string `comma:"name"`
	ID       string `comma:"id"`
}
