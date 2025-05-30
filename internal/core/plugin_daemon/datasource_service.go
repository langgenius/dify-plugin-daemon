package plugin_daemon

import (
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/datasource_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/requests"
)

func ValidateDatasourceCredentials(
	session *session_manager.Session,
	request *requests.RequestValidateDatasourceCredentials,
) (
	*stream.Stream[datasource_entities.DataSourceValidateCredentialsResponse], error,
) {
	return GenericInvokePlugin[requests.RequestValidateDatasourceCredentials, datasource_entities.DataSourceValidateCredentialsResponse](
		session,
		request,
		1,
	)
}

func InvokeDatasourceFirstStep(
	session *session_manager.Session,
	request *requests.RequestInvokeDatasourceFirstStep,
) (
	*stream.Stream[datasource_entities.DataSourceInvokeFirstStepResponse], error,
) {
	return GenericInvokePlugin[requests.RequestInvokeDatasourceFirstStep, datasource_entities.DataSourceInvokeFirstStepResponse](
		session,
		request,
		1,
	)
}

func InvokeDatasourceSecondStep(
	session *session_manager.Session,
	request *requests.RequestInvokeDatasourceSecondStep,
) (
	*stream.Stream[datasource_entities.DataSourceInvokeSecondStepResponse], error,
) {
	return GenericInvokePlugin[requests.RequestInvokeDatasourceSecondStep, datasource_entities.DataSourceInvokeSecondStepResponse](
		session,
		request,
		1,
	)
}
