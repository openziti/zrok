"""
Custom API client wrapper for zrok_api that handles zrok-specific content types.
"""

import json
import re
from typing import Dict, Optional, Any, List, Tuple, Union

from zrok_api.api_client import ApiClient
from zrok_api.api_response import ApiResponse
from zrok_api.rest import RESTResponse
from zrok_api.exceptions import ApiException

class ZrokApiClient(ApiClient):
    """
    Custom API client that extends the generated ApiClient to handle zrok-specific content types.
    """
    
    def deserialize(self, response_text: str, response_type: str, content_type: Optional[str]):
        """
        Overrides the default deserialize method to handle zrok-specific content types.
        """
        # Handle application/zrok.v1+json as if it were application/json
        if content_type and content_type.startswith("application/zrok.v1+json"):
            if response_text == "":
                data = ""
            else:
                try:
                    data = json.loads(response_text)
                except ValueError:
                    data = response_text
        elif content_type is None:
            try:
                data = json.loads(response_text)
            except ValueError:
                data = response_text
        elif content_type.startswith("application/json"):
            if response_text == "":
                data = ""
            else:
                data = json.loads(response_text)
        elif content_type.startswith("text/plain"):
            data = response_text
        else:
            raise ApiException(
                status=0,
                reason="Unsupported content type: {0}".format(content_type)
            )

        return self._ApiClient__deserialize(data, response_type)  # Access private method using name mangling
