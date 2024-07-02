from typing import Generator
from dify_plugin import ToolProvider, Plugin
from dify_plugin.tool.entities import ToolInvokeTextMessage, ToolProviderConfiguration
from dify_plugin.tool.tool import Tool
from dify_plugin.tool.entities import ToolConfiguration, ToolInvokeMessage

plugin = Plugin()

@plugin.register_tool_provider
class BasicMath(ToolProvider):
    @classmethod
    def configuration(cls) -> ToolProviderConfiguration:
        return ToolProviderConfiguration(name='basic_math')

@plugin.register_tool(BasicMath)
class Add(Tool):
    @classmethod
    def configuration(cls) -> ToolConfiguration:
        return ToolConfiguration(name='add')
    
    def _invoke(self, user_id: str, tool_parameter: dict) -> Generator[ToolInvokeMessage, None, None]:
        result = tool_parameter['a'] + tool_parameter['b']
        yield ToolInvokeTextMessage(message={'result': str(result)})

if __name__ == '__main__':
    plugin.run()
