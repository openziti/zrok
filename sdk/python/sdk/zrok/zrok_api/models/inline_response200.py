# coding: utf-8

"""
    zrok

    zrok client access  # noqa: E501

    OpenAPI spec version: 1.0.0
    
    Generated by: https://github.com/swagger-api/swagger-codegen.git
"""


import pprint
import re  # noqa: F401

import six

from zrok_api.configuration import Configuration


class InlineResponse200(object):
    """NOTE: This class is auto generated by the swagger code generator program.

    Do not edit the class manually.
    """

    """
    Attributes:
      swagger_types (dict): The key is attribute name
                            and the value is attribute type.
      attribute_map (dict): The key is attribute name
                            and the value is json key in definition.
    """
    swagger_types = {
        'frontend_token': 'str',
        'z_id': 'str',
        'url_template': 'str',
        'public_name': 'str',
        'created_at': 'int',
        'updated_at': 'int'
    }

    attribute_map = {
        'frontend_token': 'frontendToken',
        'z_id': 'zId',
        'url_template': 'urlTemplate',
        'public_name': 'publicName',
        'created_at': 'createdAt',
        'updated_at': 'updatedAt'
    }

    def __init__(self, frontend_token=None, z_id=None, url_template=None, public_name=None, created_at=None, updated_at=None, _configuration=None):  # noqa: E501
        """InlineResponse200 - a model defined in Swagger"""  # noqa: E501
        if _configuration is None:
            _configuration = Configuration()
        self._configuration = _configuration

        self._frontend_token = None
        self._z_id = None
        self._url_template = None
        self._public_name = None
        self._created_at = None
        self._updated_at = None
        self.discriminator = None

        if frontend_token is not None:
            self.frontend_token = frontend_token
        if z_id is not None:
            self.z_id = z_id
        if url_template is not None:
            self.url_template = url_template
        if public_name is not None:
            self.public_name = public_name
        if created_at is not None:
            self.created_at = created_at
        if updated_at is not None:
            self.updated_at = updated_at

    @property
    def frontend_token(self):
        """Gets the frontend_token of this InlineResponse200.  # noqa: E501


        :return: The frontend_token of this InlineResponse200.  # noqa: E501
        :rtype: str
        """
        return self._frontend_token

    @frontend_token.setter
    def frontend_token(self, frontend_token):
        """Sets the frontend_token of this InlineResponse200.


        :param frontend_token: The frontend_token of this InlineResponse200.  # noqa: E501
        :type: str
        """

        self._frontend_token = frontend_token

    @property
    def z_id(self):
        """Gets the z_id of this InlineResponse200.  # noqa: E501


        :return: The z_id of this InlineResponse200.  # noqa: E501
        :rtype: str
        """
        return self._z_id

    @z_id.setter
    def z_id(self, z_id):
        """Sets the z_id of this InlineResponse200.


        :param z_id: The z_id of this InlineResponse200.  # noqa: E501
        :type: str
        """

        self._z_id = z_id

    @property
    def url_template(self):
        """Gets the url_template of this InlineResponse200.  # noqa: E501


        :return: The url_template of this InlineResponse200.  # noqa: E501
        :rtype: str
        """
        return self._url_template

    @url_template.setter
    def url_template(self, url_template):
        """Sets the url_template of this InlineResponse200.


        :param url_template: The url_template of this InlineResponse200.  # noqa: E501
        :type: str
        """

        self._url_template = url_template

    @property
    def public_name(self):
        """Gets the public_name of this InlineResponse200.  # noqa: E501


        :return: The public_name of this InlineResponse200.  # noqa: E501
        :rtype: str
        """
        return self._public_name

    @public_name.setter
    def public_name(self, public_name):
        """Sets the public_name of this InlineResponse200.


        :param public_name: The public_name of this InlineResponse200.  # noqa: E501
        :type: str
        """

        self._public_name = public_name

    @property
    def created_at(self):
        """Gets the created_at of this InlineResponse200.  # noqa: E501


        :return: The created_at of this InlineResponse200.  # noqa: E501
        :rtype: int
        """
        return self._created_at

    @created_at.setter
    def created_at(self, created_at):
        """Sets the created_at of this InlineResponse200.


        :param created_at: The created_at of this InlineResponse200.  # noqa: E501
        :type: int
        """

        self._created_at = created_at

    @property
    def updated_at(self):
        """Gets the updated_at of this InlineResponse200.  # noqa: E501


        :return: The updated_at of this InlineResponse200.  # noqa: E501
        :rtype: int
        """
        return self._updated_at

    @updated_at.setter
    def updated_at(self, updated_at):
        """Sets the updated_at of this InlineResponse200.


        :param updated_at: The updated_at of this InlineResponse200.  # noqa: E501
        :type: int
        """

        self._updated_at = updated_at

    def to_dict(self):
        """Returns the model properties as a dict"""
        result = {}

        for attr, _ in six.iteritems(self.swagger_types):
            value = getattr(self, attr)
            if isinstance(value, list):
                result[attr] = list(map(
                    lambda x: x.to_dict() if hasattr(x, "to_dict") else x,
                    value
                ))
            elif hasattr(value, "to_dict"):
                result[attr] = value.to_dict()
            elif isinstance(value, dict):
                result[attr] = dict(map(
                    lambda item: (item[0], item[1].to_dict())
                    if hasattr(item[1], "to_dict") else item,
                    value.items()
                ))
            else:
                result[attr] = value
        if issubclass(InlineResponse200, dict):
            for key, value in self.items():
                result[key] = value

        return result

    def to_str(self):
        """Returns the string representation of the model"""
        return pprint.pformat(self.to_dict())

    def __repr__(self):
        """For `print` and `pprint`"""
        return self.to_str()

    def __eq__(self, other):
        """Returns true if both objects are equal"""
        if not isinstance(other, InlineResponse200):
            return False

        return self.to_dict() == other.to_dict()

    def __ne__(self, other):
        """Returns true if both objects are not equal"""
        if not isinstance(other, InlineResponse200):
            return True

        return self.to_dict() != other.to_dict()
