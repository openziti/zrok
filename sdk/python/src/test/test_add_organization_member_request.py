# coding: utf-8

"""
    zrok

    zrok client access

    The version of the OpenAPI document: 1.0.0
    Generated by OpenAPI Generator (https://openapi-generator.tech)

    Do not edit the class manually.
"""  # noqa: E501


import unittest

from zrok_api.models.add_organization_member_request import AddOrganizationMemberRequest

class TestAddOrganizationMemberRequest(unittest.TestCase):
    """AddOrganizationMemberRequest unit test stubs"""

    def setUp(self):
        pass

    def tearDown(self):
        pass

    def make_instance(self, include_optional) -> AddOrganizationMemberRequest:
        """Test AddOrganizationMemberRequest
            include_optional is a boolean, when False only required
            params are included, when True both required and
            optional params are included """
        # uncomment below to create an instance of `AddOrganizationMemberRequest`
        """
        model = AddOrganizationMemberRequest()
        if include_optional:
            return AddOrganizationMemberRequest(
                organization_token = '',
                email = '',
                admin = True
            )
        else:
            return AddOrganizationMemberRequest(
        )
        """

    def testAddOrganizationMemberRequest(self):
        """Test AddOrganizationMemberRequest"""
        # inst_req_only = self.make_instance(include_optional=False)
        # inst_req_and_optional = self.make_instance(include_optional=True)

if __name__ == '__main__':
    unittest.main()
