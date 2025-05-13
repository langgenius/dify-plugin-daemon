from typing import Mapping
from dify_plugin.interfaces.endpoint import EndpointGroup


class {{ .PluginName | SnakeToCamel }}EndpointGroup(EndpointGroup):
    def _setup(self, settings: Mapping):
        """
        Setup the endpoint group

        :param settings: the settings of the endpoint group
        """
