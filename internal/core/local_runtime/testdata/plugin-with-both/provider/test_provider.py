from dify_plugin import ToolProvider


class TestProvider(ToolProvider):
    def validate_credentials(self, credentials: dict) -> None:
        pass
