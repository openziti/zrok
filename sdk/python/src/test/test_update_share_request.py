# coding: utf-8

"""
    zrok

    zrok client access

    The version of the OpenAPI document: 1.0.0
    Generated by OpenAPI Generator (https://openapi-generator.tech)

    Do not edit the class manually.
"""  # noqa: E501


import unittest

from zrok_api.models.update_share_request import UpdateShareRequest

class TestUpdateShareRequest(unittest.TestCase):
    """UpdateShareRequest unit test stubs"""

    def setUp(self):
        pass

    def tearDown(self):
        pass

    def make_instance(self, include_optional) -> UpdateShareRequest:
        """Test UpdateShareRequest
            include_optional is a boolean, when False only required
            params are included, when True both required and
            optional params are included """
        # uncomment below to create an instance of `UpdateShareRequest`
        """
        model = UpdateShareRequest()
        if include_optional:
            return UpdateShareRequest(
                share_token = '',
                backend_proxy_endpoint = '',
                add_access_grants = [
                    ''
                    ],
                remove_access_grants = [
                    ''
                    ]
            )
        else:
            return UpdateShareRequest(
        )
        """

    def testUpdateShareRequest(self):
        """Test UpdateShareRequest"""
        # inst_req_only = self.make_instance(include_optional=False)
        # inst_req_and_optional = self.make_instance(include_optional=True)

if __name__ == '__main__':
    unittest.main()
