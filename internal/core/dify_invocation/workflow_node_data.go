package dify_invocation

type WorkflowNodeData interface {
	KnowledgeRetrievalNodeData | QuestionClassifierNodeData | ParameterExtractorNodeData
}

type NodeType string

const (
	KNOWLEDGE_RETRIEVAL NodeType = "knowledge_retrieval"
	QUESTION_CLASSIFIER NodeType = "question_classifier"
	PARAMETER_EXTRACTOR NodeType = "parameter_extractor"
	CODE                NodeType = "code"
)

type KnowledgeRetrievalNodeData struct {
}

type QuestionClassifierNodeData struct {
}

type ParameterExtractorNodeData struct {
}
