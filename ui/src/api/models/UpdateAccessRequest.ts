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
 * @interface UpdateAccessRequest
 */
export interface UpdateAccessRequest {
    /**
     * 
     * @type {string}
     * @memberof UpdateAccessRequest
     */
    frontendToken?: string;
    /**
     * 
     * @type {string}
     * @memberof UpdateAccessRequest
     */
    description?: string;
}

/**
 * Check if a given object implements the UpdateAccessRequest interface.
 */
export function instanceOfUpdateAccessRequest(value: object): value is UpdateAccessRequest {
    return true;
}

export function UpdateAccessRequestFromJSON(json: any): UpdateAccessRequest {
    return UpdateAccessRequestFromJSONTyped(json, false);
}

export function UpdateAccessRequestFromJSONTyped(json: any, ignoreDiscriminator: boolean): UpdateAccessRequest {
    if (json == null) {
        return json;
    }
    return {
        
        'frontendToken': json['frontendToken'] == null ? undefined : json['frontendToken'],
        'description': json['description'] == null ? undefined : json['description'],
    };
}

export function UpdateAccessRequestToJSON(value?: UpdateAccessRequest | null): any {
    if (value == null) {
        return value;
    }
    return {
        
        'frontendToken': value['frontendToken'],
        'description': value['description'],
    };
}

