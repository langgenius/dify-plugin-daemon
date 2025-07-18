from typing import Any, Mapping

from dify_plugin import ToolProvider
from dify_plugin.errors.tool import ToolProviderCredentialValidationError, ToolProviderOAuthError
from dify_plugin.interfaces.tool import Request


class {{ .PluginName | SnakeToCamel }}Provider(ToolProvider):
    
    def _validate_credentials(self, credentials: dict[str, Any]) -> None:
        try:
            """
            IMPLEMENT YOUR VALIDATION HERE
            """
        except Exception as e:
            raise ToolProviderCredentialValidationError(str(e))

    def _oauth_get_authorization_url(self, redirect_uri: str, system_credentials: Mapping[str, Any]) -> str:
        """
        Generate the authorization URL for {{ .PluginName }} OAuth.
        """
        try:
            """
            IMPLEMENT YOUR AUTHORIZATION URL GENERATION HERE
            """
        except Exception as e:
            raise ToolProviderOAuthError(str(e))
        return ""
        
    def _oauth_get_credentials(
        self, redirect_uri: str, system_credentials: Mapping[str, Any], request: Request
    ) -> Mapping[str, Any]:
        """
        Exchange code for access_token.
        """
        try:
            """
            IMPLEMENT YOUR CREDENTIALS EXCHANGE HERE
            """
        except Exception as e:
            raise ToolProviderOAuthError(str(e))
        return dict()