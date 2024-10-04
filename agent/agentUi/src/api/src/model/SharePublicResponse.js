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
 * The SharePublicResponse model module.
 * @module model/SharePublicResponse
 * @version version not set
 */
class SharePublicResponse {
    /**
     * Constructs a new <code>SharePublicResponse</code>.
     * @alias module:model/SharePublicResponse
     */
    constructor() { 
        
        SharePublicResponse.initialize(this);
    }

    /**
     * Initializes the fields of this object.
     * This method is used by the constructors of any subclasses, in order to implement multiple inheritance (mix-ins).
     * Only for internal use.
     */
    static initialize(obj) { 
    }

    /**
     * Constructs a <code>SharePublicResponse</code> from a plain JavaScript object, optionally creating a new instance.
     * Copies all relevant properties from <code>data</code> to <code>obj</code> if supplied or a new instance if not.
     * @param {Object} data The plain JavaScript object bearing properties of interest.
     * @param {module:model/SharePublicResponse} obj Optional instance to populate.
     * @return {module:model/SharePublicResponse} The populated <code>SharePublicResponse</code> instance.
     */
    static constructFromObject(data, obj) {
        if (data) {
            obj = obj || new SharePublicResponse();

            if (data.hasOwnProperty('token')) {
                obj['token'] = ApiClient.convertToType(data['token'], 'String');
            }
            if (data.hasOwnProperty('frontendEndpoints')) {
                obj['frontendEndpoints'] = ApiClient.convertToType(data['frontendEndpoints'], ['String']);
            }
        }
        return obj;
    }

    /**
     * Validates the JSON data with respect to <code>SharePublicResponse</code>.
     * @param {Object} data The plain JavaScript object bearing properties of interest.
     * @return {boolean} to indicate whether the JSON data is valid with respect to <code>SharePublicResponse</code>.
     */
    static validateJSON(data) {
        // ensure the json data is a string
        if (data['token'] && !(typeof data['token'] === 'string' || data['token'] instanceof String)) {
            throw new Error("Expected the field `token` to be a primitive type in the JSON string but got " + data['token']);
        }
        // ensure the json data is an array
        if (!Array.isArray(data['frontendEndpoints'])) {
            throw new Error("Expected the field `frontendEndpoints` to be an array in the JSON data but got " + data['frontendEndpoints']);
        }

        return true;
    }


}



/**
 * @member {String} token
 */
SharePublicResponse.prototype['token'] = undefined;

/**
 * @member {Array.<String>} frontendEndpoints
 */
SharePublicResponse.prototype['frontendEndpoints'] = undefined;






export default SharePublicResponse;
