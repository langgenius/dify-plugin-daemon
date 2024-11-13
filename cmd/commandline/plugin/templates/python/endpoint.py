import time
from typing import Mapping
from werkzeug import Request, Response
from dify_plugin import Endpoint


class {{ .PluginName | SnakeToCamel }}Endpoint(Endpoint):
    def _invoke(self, r: Request, values: Mapping, settings: Mapping) -> Response:
        """
        Invokes the endpoint with the given request.
        """
        def generator():
            for i in range(10):
                time.sleep(1)
                yield f"{i} <br>"

        return Response(generator(), status=200, content_type="text/html")
