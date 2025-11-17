package plugin_daemon

import (
	"github.com/langgenius/dify-plugin-daemon/internal/core/session_manager"
	"github.com/langgenius/dify-plugin-daemon/internal/utils/stream"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/oauth_entities"
	"github.com/langgenius/dify-plugin-daemon/pkg/entities/requests"
)

func OAuthGetAuthorizationURL(
	session *session_manager.Session,
	request *requests.RequestOAuthGetAuthorizationURL,
) (*stream.Stream[oauth_entities.OAuthGetAuthorizationURLResult], error) {
	return GenericInvokePlugin[requests.RequestOAuthGetAuthorizationURL, oauth_entities.OAuthGetAuthorizationURLResult](
		session,
		request,
		1,
	)
}

func OAuthGetCredentials(
	session *session_manager.Session,
	request *requests.RequestOAuthGetCredentials,
) (*stream.Stream[oauth_entities.OAuthGetCredentialsResult], error) {
	return GenericInvokePlugin[requests.RequestOAuthGetCredentials, oauth_entities.OAuthGetCredentialsResult](
		session,
		request,
		1,
	)
}
