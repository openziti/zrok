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


class Share(object):
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
        'share_token': 'str',
        'z_id': 'str',
        'share_mode': 'str',
        'backend_mode': 'str',
        'frontend_selection': 'str',
        'frontend_endpoint': 'str',
        'backend_proxy_endpoint': 'str',
        'reserved': 'bool',
        'activity': 'SparkData',
        'limited': 'bool',
        'created_at': 'int',
        'updated_at': 'int'
    }

    attribute_map = {
        'share_token': 'shareToken',
        'z_id': 'zId',
        'share_mode': 'shareMode',
        'backend_mode': 'backendMode',
        'frontend_selection': 'frontendSelection',
        'frontend_endpoint': 'frontendEndpoint',
        'backend_proxy_endpoint': 'backendProxyEndpoint',
        'reserved': 'reserved',
        'activity': 'activity',
        'limited': 'limited',
        'created_at': 'createdAt',
        'updated_at': 'updatedAt'
    }

    def __init__(self, share_token=None, z_id=None, share_mode=None, backend_mode=None, frontend_selection=None, frontend_endpoint=None, backend_proxy_endpoint=None, reserved=None, activity=None, limited=None, created_at=None, updated_at=None, _configuration=None):  # noqa: E501
        """Share - a model defined in Swagger"""  # noqa: E501
        if _configuration is None:
            _configuration = Configuration()
        self._configuration = _configuration

        self._share_token = None
        self._z_id = None
        self._share_mode = None
        self._backend_mode = None
        self._frontend_selection = None
        self._frontend_endpoint = None
        self._backend_proxy_endpoint = None
        self._reserved = None
        self._activity = None
        self._limited = None
        self._created_at = None
        self._updated_at = None
        self.discriminator = None

        if share_token is not None:
            self.share_token = share_token
        if z_id is not None:
            self.z_id = z_id
        if share_mode is not None:
            self.share_mode = share_mode
        if backend_mode is not None:
            self.backend_mode = backend_mode
        if frontend_selection is not None:
            self.frontend_selection = frontend_selection
        if frontend_endpoint is not None:
            self.frontend_endpoint = frontend_endpoint
        if backend_proxy_endpoint is not None:
            self.backend_proxy_endpoint = backend_proxy_endpoint
        if reserved is not None:
            self.reserved = reserved
        if activity is not None:
            self.activity = activity
        if limited is not None:
            self.limited = limited
        if created_at is not None:
            self.created_at = created_at
        if updated_at is not None:
            self.updated_at = updated_at

    @property
    def share_token(self):
        """Gets the share_token of this Share.  # noqa: E501


        :return: The share_token of this Share.  # noqa: E501
        :rtype: str
        """
        return self._share_token

    @share_token.setter
    def share_token(self, share_token):
        """Sets the share_token of this Share.


        :param share_token: The share_token of this Share.  # noqa: E501
        :type: str
        """

        self._share_token = share_token

    @property
    def z_id(self):
        """Gets the z_id of this Share.  # noqa: E501


        :return: The z_id of this Share.  # noqa: E501
        :rtype: str
        """
        return self._z_id

    @z_id.setter
    def z_id(self, z_id):
        """Sets the z_id of this Share.


        :param z_id: The z_id of this Share.  # noqa: E501
        :type: str
        """

        self._z_id = z_id

    @property
    def share_mode(self):
        """Gets the share_mode of this Share.  # noqa: E501


        :return: The share_mode of this Share.  # noqa: E501
        :rtype: str
        """
        return self._share_mode

    @share_mode.setter
    def share_mode(self, share_mode):
        """Sets the share_mode of this Share.


        :param share_mode: The share_mode of this Share.  # noqa: E501
        :type: str
        """

        self._share_mode = share_mode

    @property
    def backend_mode(self):
        """Gets the backend_mode of this Share.  # noqa: E501


        :return: The backend_mode of this Share.  # noqa: E501
        :rtype: str
        """
        return self._backend_mode

    @backend_mode.setter
    def backend_mode(self, backend_mode):
        """Sets the backend_mode of this Share.


        :param backend_mode: The backend_mode of this Share.  # noqa: E501
        :type: str
        """

        self._backend_mode = backend_mode

    @property
    def frontend_selection(self):
        """Gets the frontend_selection of this Share.  # noqa: E501


        :return: The frontend_selection of this Share.  # noqa: E501
        :rtype: str
        """
        return self._frontend_selection

    @frontend_selection.setter
    def frontend_selection(self, frontend_selection):
        """Sets the frontend_selection of this Share.


        :param frontend_selection: The frontend_selection of this Share.  # noqa: E501
        :type: str
        """

        self._frontend_selection = frontend_selection

    @property
    def frontend_endpoint(self):
        """Gets the frontend_endpoint of this Share.  # noqa: E501


        :return: The frontend_endpoint of this Share.  # noqa: E501
        :rtype: str
        """
        return self._frontend_endpoint

    @frontend_endpoint.setter
    def frontend_endpoint(self, frontend_endpoint):
        """Sets the frontend_endpoint of this Share.


        :param frontend_endpoint: The frontend_endpoint of this Share.  # noqa: E501
        :type: str
        """

        self._frontend_endpoint = frontend_endpoint

    @property
    def backend_proxy_endpoint(self):
        """Gets the backend_proxy_endpoint of this Share.  # noqa: E501


        :return: The backend_proxy_endpoint of this Share.  # noqa: E501
        :rtype: str
        """
        return self._backend_proxy_endpoint

    @backend_proxy_endpoint.setter
    def backend_proxy_endpoint(self, backend_proxy_endpoint):
        """Sets the backend_proxy_endpoint of this Share.


        :param backend_proxy_endpoint: The backend_proxy_endpoint of this Share.  # noqa: E501
        :type: str
        """

        self._backend_proxy_endpoint = backend_proxy_endpoint

    @property
    def reserved(self):
        """Gets the reserved of this Share.  # noqa: E501


        :return: The reserved of this Share.  # noqa: E501
        :rtype: bool
        """
        return self._reserved

    @reserved.setter
    def reserved(self, reserved):
        """Sets the reserved of this Share.


        :param reserved: The reserved of this Share.  # noqa: E501
        :type: bool
        """

        self._reserved = reserved

    @property
    def activity(self):
        """Gets the activity of this Share.  # noqa: E501


        :return: The activity of this Share.  # noqa: E501
        :rtype: SparkData
        """
        return self._activity

    @activity.setter
    def activity(self, activity):
        """Sets the activity of this Share.


        :param activity: The activity of this Share.  # noqa: E501
        :type: SparkData
        """

        self._activity = activity

    @property
    def limited(self):
        """Gets the limited of this Share.  # noqa: E501


        :return: The limited of this Share.  # noqa: E501
        :rtype: bool
        """
        return self._limited

    @limited.setter
    def limited(self, limited):
        """Sets the limited of this Share.


        :param limited: The limited of this Share.  # noqa: E501
        :type: bool
        """

        self._limited = limited

    @property
    def created_at(self):
        """Gets the created_at of this Share.  # noqa: E501


        :return: The created_at of this Share.  # noqa: E501
        :rtype: int
        """
        return self._created_at

    @created_at.setter
    def created_at(self, created_at):
        """Sets the created_at of this Share.


        :param created_at: The created_at of this Share.  # noqa: E501
        :type: int
        """

        self._created_at = created_at

    @property
    def updated_at(self):
        """Gets the updated_at of this Share.  # noqa: E501


        :return: The updated_at of this Share.  # noqa: E501
        :rtype: int
        """
        return self._updated_at

    @updated_at.setter
    def updated_at(self, updated_at):
        """Sets the updated_at of this Share.


        :param updated_at: The updated_at of this Share.  # noqa: E501
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
        if issubclass(Share, dict):
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
        if not isinstance(other, Share):
            return False

        return self.to_dict() == other.to_dict()

    def __ne__(self, other):
        """Returns true if both objects are not equal"""
        if not isinstance(other, Share):
            return True

        return self.to_dict() != other.to_dict()
