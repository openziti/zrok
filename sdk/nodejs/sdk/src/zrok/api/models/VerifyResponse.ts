/* tslint:disable */
/* eslint-disable */
/**
 * zrok
 * zrok client access
 *
 * The version of the OpenAPI document: 0.3.0
 * 
 *
 * NOTE: This class is auto generated by OpenAPI Generator (https://openapi-generator.tech).
 * https://openapi-generator.tech
 * Do not edit the class manually.
 */

import { exists, mapValues } from '../runtime';
/**
 * 
 * @export
 * @interface VerifyResponse
 */
export interface VerifyResponse {
    /**
     * 
     * @type {string}
     * @memberof VerifyResponse
     */
    email?: string;
}

/**
 * Check if a given object implements the VerifyResponse interface.
 */
export function instanceOfVerifyResponse(value: object): boolean {
    let isInstance = true;

    return isInstance;
}

export function VerifyResponseFromJSON(json: any): VerifyResponse {
    return VerifyResponseFromJSONTyped(json, false);
}

export function VerifyResponseFromJSONTyped(json: any, ignoreDiscriminator: boolean): VerifyResponse {
    if ((json === undefined) || (json === null)) {
        return json;
    }
    return {
        
        'email': !exists(json, 'email') ? undefined : json['email'],
    };
}

export function VerifyResponseToJSON(value?: VerifyResponse | null): any {
    if (value === undefined) {
        return undefined;
    }
    if (value === null) {
        return null;
    }
    return {
        
        'email': value.email,
    };
}
