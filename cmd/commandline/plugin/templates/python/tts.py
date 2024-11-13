from collections.abc import Generator
from typing import Optional

from dify_plugin import TTSModel
from dify_plugin.errors.model import (
    CredentialsValidateFailedError,
)


class {{ .PluginName | SnakeToCamel }}Text2SpeechModel(TTSModel):
    """
    Model class for OpenAI Speech to text model.
    """

    def _invoke(
        self,
        model: str,
        credentials: dict,
        content_text: str,
        voice: str,
        user: Optional[str] = None,
    ) -> bytes | Generator[bytes, None, None]:
        """
        _invoke text2speech model

        :param model: model name
        :param tenant_id: user tenant id
        :param credentials: model credentials
        :param content_text: text content to be translated
        :param voice: model timbre
        :param user: unique user id
        :return: text translated to audio file
        """
        pass

    def validate_credentials(
        self, model: str, credentials: dict, user: Optional[str] = None
    ) -> None:
        """
        validate credentials text2speech model

        :param model: model name
        :param credentials: model credentials
        :param user: unique user id
        :return: text translated to audio file
        """
        try:
            pass
        except Exception as ex:
            raise CredentialsValidateFailedError(str(ex))
