package transaction

import "github.com/gin-gonic/gin"

type AWSEventHandler struct {
}

func NewAWSEventHandler() *AWSEventHandler {
	return &AWSEventHandler{}
}

func (h *AWSEventHandler) Handle(ctx *gin.Context) {

}
