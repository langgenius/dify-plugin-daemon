package requests

type WorkflowNodeData interface {
	KnowledgeRetrievalNodeData | QuestionClassifierNodeData | ParameterExtractorNodeData
}

type NodeType string

const (
	KNOWLEDGE_RETRIEVAL NodeType = "knowledge_retrieval"
	QUESTION_CLASSIFIER NodeType = "question_classifier"
	PARAMETER_EXTRACTOR NodeType = "parameter_extractor"
)

type KnowledgeRetrievalNodeData struct {
}

type QuestionClassifierNodeData struct {
}

type ParameterExtractorNodeData struct {
}

type InvokeNodeRequest[T WorkflowNodeData] struct {
	NodeType NodeType `json:"node_type"`
	NodeData T        `json:"node_data"`
}
