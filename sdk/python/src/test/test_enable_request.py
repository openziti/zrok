# coding: utf-8

"""
    zrok

    zrok client access

    The version of the OpenAPI document: 1.0.0
    Generated by OpenAPI Generator (https://openapi-generator.tech)

    Do not edit the class manually.
"""  # noqa: E501


import unittest

from zrok_api.models.enable_request import EnableRequest

class TestEnableRequest(unittest.TestCase):
    """EnableRequest unit test stubs"""

    def setUp(self):
        pass

    def tearDown(self):
        pass

    def make_instance(self, include_optional) -> EnableRequest:
        """Test EnableRequest
            include_optional is a boolean, when False only required
            params are included, when True both required and
            optional params are included """
        # uncomment below to create an instance of `EnableRequest`
        """
        model = EnableRequest()
        if include_optional:
            return EnableRequest(
                description = '',
                host = ''
            )
        else:
            return EnableRequest(
        )
        """

    def testEnableRequest(self):
        """Test EnableRequest"""
        # inst_req_only = self.make_instance(include_optional=False)
        # inst_req_and_optional = self.make_instance(include_optional=True)

if __name__ == '__main__':
    unittest.main()
