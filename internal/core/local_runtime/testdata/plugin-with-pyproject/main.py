from dify_plugin import Tool

class TestTool(Tool):
    def _invoke(self, tool_parameters: dict) -> str:
        return "test result"
