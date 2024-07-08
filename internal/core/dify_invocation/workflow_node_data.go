package dify_invocation

type WorkflowNodeData interface {
	KnowledgeRetrievalNodeData | QuestionClassifierNodeData |
		ParameterExtractorNodeData | CodeNodeData
}

type KnowledgeRetrievalNodeData struct {
}

type QuestionClassifierNodeData struct {
}

type ParameterExtractorNodeData struct {
}

type CodeNodeData struct {
}
