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
 * @interface InviteRequest
 */
export interface InviteRequest {
    /**
     * 
     * @type {string}
     * @memberof InviteRequest
     */
    email?: string;
    /**
     * 
     * @type {string}
     * @memberof InviteRequest
     */
    token?: string;
}

/**
 * Check if a given object implements the InviteRequest interface.
 */
export function instanceOfInviteRequest(value: object): value is InviteRequest {
    return true;
}

export function InviteRequestFromJSON(json: any): InviteRequest {
    return InviteRequestFromJSONTyped(json, false);
}

export function InviteRequestFromJSONTyped(json: any, ignoreDiscriminator: boolean): InviteRequest {
    if (json == null) {
        return json;
    }
    return {
        
        'email': json['email'] == null ? undefined : json['email'],
        'token': json['token'] == null ? undefined : json['token'],
    };
}

export function InviteRequestToJSON(value?: InviteRequest | null): any {
    if (value == null) {
        return value;
    }
    return {
        
        'email': value['email'],
        'token': value['token'],
    };
}
