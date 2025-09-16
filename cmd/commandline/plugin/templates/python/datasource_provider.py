from typing import Any, Mapping
import secrets
import urllib.parse

from dify_plugin.errors.tool import ToolProviderCredentialValidationError, DatasourceOAuthError
from dify_plugin.interfaces.datasource import DatasourceProvider, DatasourceOAuthCredentials



class {{ .PluginName | SnakeToCamel }}Provider(DatasourceProvider):

    def _validate_credentials(self, credentials: Mapping[str, Any]) -> None:
        try:
            """
            IMPLEMENT YOUR VALIDATION HERE
            """
        except Exception as e:
            raise ToolProviderCredentialValidationError(str(e))

    def _oauth_get_authorization_url(self, redirect_uri: str, system_credentials: Mapping[str, Any]) -> str:
        """
        Generate the authorization URL for Box OAuth 2.0.
        
        Args:
            redirect_uri: The redirect URI after authorization
            system_credentials: System credentials containing client_id and client_secret
            
        Returns:
            Authorization URL string
        """
        state = secrets.token_urlsafe(32)
        params = {
            "client_id": system_credentials["client_id"],
            "redirect_uri": redirect_uri,
            "response_type": "code",
            "scope": self._REQUIRED_SCOPES,
            "state": state,
        }
        return f"{self._AUTH_URL}?{urllib.parse.urlencode(params)}"


    #########################################################################################
    # If OAuth is supported, uncomment the following functions.
    # Warning: please make sure that the sdk version is 0.5.0 or higher.
    #########################################################################################
    # def _oauth_get_authorization_url(self, redirect_uri: str, system_credentials: Mapping[str, Any]) -> str:
    #     """
    #     Generate the authorization URL for {{ .PluginName }} OAuth.
    #     """
    #     try:
    #         """
    #         IMPLEMENT YOUR AUTHORIZATION URL GENERATION HERE
    #         """
    #     except Exception as e:
    #         raise DatasourceOAuthError(str(e))
    #     return ""

    # def _oauth_get_credentials(
    #     self, redirect_uri: str, system_credentials: Mapping[str, Any], request: Request
    # ) -> DatasourceOAuthCredentials:
    #     """
    #     Exchange code for access_token.
    #     """
    #     try:
    #         """
    #         IMPLEMENT YOUR CREDENTIALS EXCHANGE HERE
    #         """
    #     except Exception as e:
    #         raise DatasourceOAuthError(str(e))
    #     return DatasourceOAuthCredentials(
    #         name="",
    #         avatar_url="",
    #         expires_at=-1,
    #         credentials={},
    #     )

    # def _oauth_refresh_credentials(
    #     self, redirect_uri: str, system_credentials: Mapping[str, Any], credentials: Mapping[str, Any]
    # ) -> DatasourceOAuthCredentials:
    #     """
    #     Refresh the credentials
    #     """
    #     return DatasourceOAuthCredentials(
    #         name="",
    #         avatar_url="",
    #         expires_at=-1,
    #         credentials={},
    #     )

