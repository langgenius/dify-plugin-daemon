package dify_invocation

type WorkflowNodeData interface {
	FromMap(map[string]any) error

	*KnowledgeRetrievalNodeData | *QuestionClassifierNodeData | *ParameterExtractorNodeData
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

func (r *KnowledgeRetrievalNodeData) FromMap(data map[string]any) error {
	return nil
}

type QuestionClassifierNodeData struct {
}

func (r *QuestionClassifierNodeData) FromMap(data map[string]any) error {
	return nil
}

type ParameterExtractorNodeData struct {
}

func (r *ParameterExtractorNodeData) FromMap(data map[string]any) error {
	return nil
}
