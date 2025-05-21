package service

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_daemon/access_types"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/datasource_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/requests"
)

func DatasourceValidateCredentials(
	r *plugin_entities.InvokePluginRequest[requests.RequestValidateDatasourceCredentials],
	ctx *gin.Context,
	maxExecutionTimeout time.Duration,
) {
	baseSSEWithSession(
		func(session *session_manager.Session) (*stream.Stream[datasource_entities.DataSourceValidateCredentialsResponse], error) {
			return plugin_daemon.ValidateDatasourceCredentials(session, &r.Data)
		},
		access_types.PLUGIN_ACCESS_TYPE_DATASOURCE,
		access_types.PLUGIN_ACCESS_ACTION_VALIDATE_CREDENTIALS,
		r,
		ctx,
		int(maxExecutionTimeout.Seconds()),
	)
}

func DatasourceInvokeFirstStep(
	r *plugin_entities.InvokePluginRequest[requests.RequestInvokeDatasourceFirstStep],
	ctx *gin.Context,
	maxExecutionTimeout time.Duration,
) {
	baseSSEWithSession(
		func(session *session_manager.Session) (*stream.Stream[datasource_entities.DataSourceInvokeFirstStepResponse], error) {
			return plugin_daemon.InvokeDatasourceFirstStep(session, &r.Data)
		},
		access_types.PLUGIN_ACCESS_TYPE_DATASOURCE,
		access_types.PLUGIN_ACCESS_ACTION_INVOKE_DATASOURCE_FIRST_STEP,
		r,
		ctx,
		int(maxExecutionTimeout.Seconds()),
	)
}

func DatasourceInvokeSecondStep(
	r *plugin_entities.InvokePluginRequest[requests.RequestInvokeDatasourceSecondStep],
	ctx *gin.Context,
	maxExecutionTimeout time.Duration,
) {
	baseSSEWithSession(
		func(session *session_manager.Session) (*stream.Stream[datasource_entities.DataSourceInvokeSecondStepResponse], error) {
			return plugin_daemon.InvokeDatasourceSecondStep(session, &r.Data)
		},
		access_types.PLUGIN_ACCESS_TYPE_DATASOURCE,
		access_types.PLUGIN_ACCESS_ACTION_INVOKE_DATASOURCE_SECOND_STEP,
		r,
		ctx,
		int(maxExecutionTimeout.Seconds()),
	)
}
