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


class SparkDataSample(object):
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
        'rx': 'float',
        'tx': 'float'
    }

    attribute_map = {
        'rx': 'rx',
        'tx': 'tx'
    }

    def __init__(self, rx=None, tx=None, _configuration=None):  # noqa: E501
        """SparkDataSample - a model defined in Swagger"""  # noqa: E501
        if _configuration is None:
            _configuration = Configuration()
        self._configuration = _configuration

        self._rx = None
        self._tx = None
        self.discriminator = None

        if rx is not None:
            self.rx = rx
        if tx is not None:
            self.tx = tx

    @property
    def rx(self):
        """Gets the rx of this SparkDataSample.  # noqa: E501


        :return: The rx of this SparkDataSample.  # noqa: E501
        :rtype: float
        """
        return self._rx

    @rx.setter
    def rx(self, rx):
        """Sets the rx of this SparkDataSample.


        :param rx: The rx of this SparkDataSample.  # noqa: E501
        :type: float
        """

        self._rx = rx

    @property
    def tx(self):
        """Gets the tx of this SparkDataSample.  # noqa: E501


        :return: The tx of this SparkDataSample.  # noqa: E501
        :rtype: float
        """
        return self._tx

    @tx.setter
    def tx(self, tx):
        """Sets the tx of this SparkDataSample.


        :param tx: The tx of this SparkDataSample.  # noqa: E501
        :type: float
        """

        self._tx = tx

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
        if issubclass(SparkDataSample, dict):
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
        if not isinstance(other, SparkDataSample):
            return False

        return self.to_dict() == other.to_dict()

    def __ne__(self, other):
        """Returns true if both objects are not equal"""
        if not isinstance(other, SparkDataSample):
            return True

        return self.to_dict() != other.to_dict()
