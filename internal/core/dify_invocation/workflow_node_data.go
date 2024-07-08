package dify_invocation

type WorkflowNodeData interface {
	FromMap(map[string]any) error

	*KnowledgeRetrievalNodeData | *QuestionClassifierNodeData | *ParameterExtractorNodeData | *CodeNodeData
}

const (
	NODE_TYPE_KNOWLEDGE_RETRIEVAL = "knowledge_retrieval"
	NODE_TYPE_QUESTION_CLASSIFIER = "question_classifier"
	NODE_TYPE_PARAMETER_EXTRACTOR = "parameter_extractor"
	NODE_TYPE_CODE                = "code"
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

type CodeNodeData struct {
}

func (r *CodeNodeData) FromMap(data map[string]any) error {
	return nil
}
