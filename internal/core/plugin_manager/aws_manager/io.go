package aws_manager

import "github.com/langgenius/dify-plugin-daemon/internal/types/entities"

func (r *AWSPluginRuntime) Listen(session_id string) *entities.BytesIOListener {
	l := entities.NewIOListener[[]byte]()
	return l
}

func (r *AWSPluginRuntime) Write(session_id string, data []byte) {

}

func (r *AWSPluginRuntime) Request(session_id string, data []byte) ([]byte, error) {
	return nil, nil
}
