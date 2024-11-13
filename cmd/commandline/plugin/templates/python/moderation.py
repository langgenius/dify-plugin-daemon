from typing import Optional

from dify_plugin.errors.model import CredentialsValidateFailedError
from dify_plugin import ModerationModel

class {{ .PluginName | SnakeToCamel }}ModerationModel(ModerationModel):
    """
    Model class for {{ .PluginName | CamelToTitle }} text moderation model.
    """

    def _invoke(self, model: str, credentials: dict,
                text: str, user: Optional[str] = None) \
            -> bool:
        """
        Invoke moderation model

        :param model: model name
        :param credentials: model credentials
        :param text: text to moderate
        :param user: unique user id
        :return: false if text is safe, true otherwise
        """
        # transform credentials to kwargs for model instance
        return True

    def validate_credentials(self, model: str, credentials: dict) -> None:
        """
        Validate model credentials

        :param model: model name
        :param credentials: model credentials
        :return:
        """
        try:
            pass
        except Exception as ex:
            raise CredentialsValidateFailedError(str(ex))
