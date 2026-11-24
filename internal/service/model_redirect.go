package service

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel"
	"github.com/langgenius/dify-plugin-daemon/internal/core/io_tunnel/access_types"
	"github.com/langgenius/dify-plugin-daemon/internal/core/plugin_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/types/exception"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/plugin_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/plugin_packager/decoder"
	"github.com/langgenius/dify-plugin-daemon/pkg/utils/stream"
)

type modelRedirectRuntimeTypeKey struct{}

func WithModelRedirectRuntimeType(ctx context.Context, runtimeType plugin_entities.PluginRuntimeType) context.Context {
	return context.WithValue(ctx, modelRedirectRuntimeTypeKey{}, runtimeType)
}

func VerifyCurrentModelRedirectTarget(maxPackageSize int64, uid plugin_entities.PluginUniqueIdentifier, provider string) bool {
	manager := plugin_manager.Manager()
	if manager == nil {
		return false
	}
	packageBytes, err := manager.GetPackage(uid)
	if err != nil {
		return false
	}
	return verifyModelRedirectPackage(maxPackageSize, packageBytes, uid, provider)
}

func verifyModelRedirectPackage(maxPackageSize int64, packageBytes []byte, uid plugin_entities.PluginUniqueIdentifier, provider string) bool {
	pkg, err := decoder.NewZipPluginDecoderWithSizeLimit(packageBytes, maxPackageSize)
	if err != nil {
		return false
	}
	defer pkg.Close()
	verification, err := pkg.Verification()
	if err != nil {
		return false
	}
	declaration, err := pkg.Manifest()
	if err != nil {
		return false
	}
	actualUID, err := pkg.UniqueIdentity()
	return err == nil && actualUID == uid && isVerifiedModelRedirectProvider(&declaration, verification, provider)
}

func isVerifiedModelRedirectProvider(declaration *plugin_entities.PluginDeclaration, verification *decoder.Verification, provider string) bool {
	if declaration == nil || !declaration.Verified || declaration.Model == nil ||
		declaration.Model.Provider != provider {
		return false
	}
	if verification == nil {
		verification = decoder.DefaultVerification()
	}
	return !isUnauthorizedLanggenius(declaration, verification)
}

func StreamModelRedirect[R any](
	request *plugin_entities.InvokePluginRequest[io_tunnel.ModelDispatchData],
	ctx *gin.Context,
	action access_types.PluginAccessAction,
	bufferSize int,
	timeout int,
) {
	baseSSEWithSession(
		func(session *session_manager.Session) (*stream.Stream[R], error) {
			return io_tunnel.GenericInvokePlugin[io_tunnel.ModelDispatchData, R](session, &request.Data, bufferSize)
		},
		access_types.PLUGIN_ACCESS_TYPE_MODEL, action, request, ctx, timeout,
	)
}

func PollModelRedirect(
	request *plugin_entities.InvokePluginRequest[io_tunnel.ModelDispatchData],
	ctx *gin.Context,
	action access_types.PluginAccessAction,
) {
	session, err := createSession(request, access_types.PLUGIN_ACCESS_TYPE_MODEL, action, ctx.GetString("cluster_id"), ctx.Request.Context())
	if err != nil {
		ctx.JSON(500, exception.InternalServerError(err).ToResponse())
		return
	}
	defer session.Close(session_manager.CloseSessionPayload{IgnoreCache: false})

	response, err := io_tunnel.PollModelRedirect(session, &request.Data)
	if err != nil {
		ctx.JSON(500, exception.InvokePluginError(err).ToResponse())
		return
	}
	ctx.JSON(200, entities.NewSuccessResponse(response))
}
