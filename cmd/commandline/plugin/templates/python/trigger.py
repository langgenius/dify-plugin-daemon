from collections.abc import Mapping
from typing import Any

from werkzeug import Request

from dify_plugin.entities.trigger import Variables
from dify_plugin.errors.trigger import EventIgnoreError
from dify_plugin.interfaces.trigger import Event


class {{ .PluginName | SnakeToCamel }}TriggerEvent(Event):
    """
    Basic example trigger handler that emits the raw incoming payload.
    Replace the logic here with your integration specific processing.
    """

    def _on_event(self, request: Request, parameters: Mapping[str, Any], payload: Mapping[str, Any]) -> Variables:
        payload = request.get_json(silent=True) or {}

        sample_filter = parameters.get("sample_filter")
        if sample_filter and sample_filter not in str(payload):
            raise EventIgnoreError()

        return Variables(
            variables={
                "message": "Hello from {{ .PluginName | SnakeToCamel }}!",
                "raw_event": payload,
            }
        )
