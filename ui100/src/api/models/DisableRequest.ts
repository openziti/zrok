/* tslint:disable */
/* eslint-disable */
/**
 * zrok
 * zrok client access
 *
 * The version of the OpenAPI document: 1.0.0
 * 
 *
 * NOTE: This class is auto generated by OpenAPI Generator (https://openapi-generator.tech).
 * https://openapi-generator.tech
 * Do not edit the class manually.
 */

import { mapValues } from '../runtime';
/**
 * 
 * @export
 * @interface DisableRequest
 */
export interface DisableRequest {
    /**
     * 
     * @type {string}
     * @memberof DisableRequest
     */
    identity?: string;
}

/**
 * Check if a given object implements the DisableRequest interface.
 */
export function instanceOfDisableRequest(value: object): value is DisableRequest {
    return true;
}

export function DisableRequestFromJSON(json: any): DisableRequest {
    return DisableRequestFromJSONTyped(json, false);
}

export function DisableRequestFromJSONTyped(json: any, ignoreDiscriminator: boolean): DisableRequest {
    if (json == null) {
        return json;
    }
    return {
        
        'identity': json['identity'] == null ? undefined : json['identity'],
    };
}

export function DisableRequestToJSON(value?: DisableRequest | null): any {
    if (value == null) {
        return value;
    }
    return {
        
        'identity': value['identity'],
    };
}
