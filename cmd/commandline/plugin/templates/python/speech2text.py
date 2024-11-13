from typing import IO, Optional

from dify_plugin.errors.model import CredentialsValidateFailedError

class {{ .PluginName | SnakeToCamel }}Speech2TextModel(Speech2TextModel):
    """
    Model class for OpenAI Speech to text model.
    """

    def _invoke(self, model: str, credentials: dict,
                file: IO[bytes], user: Optional[str] = None) \
            -> str:
        """
        Invoke speech2text model

        :param model: model name
        :param credentials: model credentials
        :param file: audio file
        :param user: unique user id
        :return: text for given audio file
        """
        pass

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
