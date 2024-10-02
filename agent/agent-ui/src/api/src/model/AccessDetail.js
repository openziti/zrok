/**
 * agent/agentGrpc/agent.proto
 * No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)
 *
 * The version of the OpenAPI document: version not set
 * 
 *
 * NOTE: This class is auto generated by OpenAPI Generator (https://openapi-generator.tech).
 * https://openapi-generator.tech
 * Do not edit the class manually.
 *
 */

import ApiClient from '../ApiClient';

/**
 * The AccessDetail model module.
 * @module model/AccessDetail
 * @version version not set
 */
class AccessDetail {
    /**
     * Constructs a new <code>AccessDetail</code>.
     * @alias module:model/AccessDetail
     */
    constructor() { 
        
        AccessDetail.initialize(this);
    }

    /**
     * Initializes the fields of this object.
     * This method is used by the constructors of any subclasses, in order to implement multiple inheritance (mix-ins).
     * Only for internal use.
     */
    static initialize(obj) { 
    }

    /**
     * Constructs a <code>AccessDetail</code> from a plain JavaScript object, optionally creating a new instance.
     * Copies all relevant properties from <code>data</code> to <code>obj</code> if supplied or a new instance if not.
     * @param {Object} data The plain JavaScript object bearing properties of interest.
     * @param {module:model/AccessDetail} obj Optional instance to populate.
     * @return {module:model/AccessDetail} The populated <code>AccessDetail</code> instance.
     */
    static constructFromObject(data, obj) {
        if (data) {
            obj = obj || new AccessDetail();

            if (data.hasOwnProperty('frontendToken')) {
                obj['frontendToken'] = ApiClient.convertToType(data['frontendToken'], 'String');
            }
            if (data.hasOwnProperty('token')) {
                obj['token'] = ApiClient.convertToType(data['token'], 'String');
            }
            if (data.hasOwnProperty('bindAddress')) {
                obj['bindAddress'] = ApiClient.convertToType(data['bindAddress'], 'String');
            }
            if (data.hasOwnProperty('responseHeaders')) {
                obj['responseHeaders'] = ApiClient.convertToType(data['responseHeaders'], ['String']);
            }
        }
        return obj;
    }

    /**
     * Validates the JSON data with respect to <code>AccessDetail</code>.
     * @param {Object} data The plain JavaScript object bearing properties of interest.
     * @return {boolean} to indicate whether the JSON data is valid with respect to <code>AccessDetail</code>.
     */
    static validateJSON(data) {
        // ensure the json data is a string
        if (data['frontendToken'] && !(typeof data['frontendToken'] === 'string' || data['frontendToken'] instanceof String)) {
            throw new Error("Expected the field `frontendToken` to be a primitive type in the JSON string but got " + data['frontendToken']);
        }
        // ensure the json data is a string
        if (data['token'] && !(typeof data['token'] === 'string' || data['token'] instanceof String)) {
            throw new Error("Expected the field `token` to be a primitive type in the JSON string but got " + data['token']);
        }
        // ensure the json data is a string
        if (data['bindAddress'] && !(typeof data['bindAddress'] === 'string' || data['bindAddress'] instanceof String)) {
            throw new Error("Expected the field `bindAddress` to be a primitive type in the JSON string but got " + data['bindAddress']);
        }
        // ensure the json data is an array
        if (!Array.isArray(data['responseHeaders'])) {
            throw new Error("Expected the field `responseHeaders` to be an array in the JSON data but got " + data['responseHeaders']);
        }

        return true;
    }


}



/**
 * @member {String} frontendToken
 */
AccessDetail.prototype['frontendToken'] = undefined;

/**
 * @member {String} token
 */
AccessDetail.prototype['token'] = undefined;

/**
 * @member {String} bindAddress
 */
AccessDetail.prototype['bindAddress'] = undefined;

/**
 * @member {Array.<String>} responseHeaders
 */
AccessDetail.prototype['responseHeaders'] = undefined;






export default AccessDetail;

