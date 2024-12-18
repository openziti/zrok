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
 * @interface RegenerateTokenRequest
 */
export interface RegenerateTokenRequest {
    /**
     * 
     * @type {string}
     * @memberof RegenerateTokenRequest
     */
    emailAddress?: string;
}

/**
 * Check if a given object implements the RegenerateTokenRequest interface.
 */
export function instanceOfRegenerateTokenRequest(value: object): value is RegenerateTokenRequest {
    return true;
}

export function RegenerateTokenRequestFromJSON(json: any): RegenerateTokenRequest {
    return RegenerateTokenRequestFromJSONTyped(json, false);
}

export function RegenerateTokenRequestFromJSONTyped(json: any, ignoreDiscriminator: boolean): RegenerateTokenRequest {
    if (json == null) {
        return json;
    }
    return {
        
        'emailAddress': json['emailAddress'] == null ? undefined : json['emailAddress'],
    };
}

export function RegenerateTokenRequestToJSON(value?: RegenerateTokenRequest | null): any {
    if (value == null) {
        return value;
    }
    return {
        
        'emailAddress': value['emailAddress'],
    };
}
