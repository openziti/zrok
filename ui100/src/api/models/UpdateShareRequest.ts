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
 * @interface UpdateShareRequest
 */
export interface UpdateShareRequest {
    /**
     * 
     * @type {string}
     * @memberof UpdateShareRequest
     */
    shrToken?: string;
    /**
     * 
     * @type {string}
     * @memberof UpdateShareRequest
     */
    backendProxyEndpoint?: string;
    /**
     * 
     * @type {Array<string>}
     * @memberof UpdateShareRequest
     */
    addAccessGrants?: Array<string>;
    /**
     * 
     * @type {Array<string>}
     * @memberof UpdateShareRequest
     */
    removeAccessGrants?: Array<string>;
}

/**
 * Check if a given object implements the UpdateShareRequest interface.
 */
export function instanceOfUpdateShareRequest(value: object): value is UpdateShareRequest {
    return true;
}

export function UpdateShareRequestFromJSON(json: any): UpdateShareRequest {
    return UpdateShareRequestFromJSONTyped(json, false);
}

export function UpdateShareRequestFromJSONTyped(json: any, ignoreDiscriminator: boolean): UpdateShareRequest {
    if (json == null) {
        return json;
    }
    return {
        
        'shrToken': json['shrToken'] == null ? undefined : json['shrToken'],
        'backendProxyEndpoint': json['backendProxyEndpoint'] == null ? undefined : json['backendProxyEndpoint'],
        'addAccessGrants': json['addAccessGrants'] == null ? undefined : json['addAccessGrants'],
        'removeAccessGrants': json['removeAccessGrants'] == null ? undefined : json['removeAccessGrants'],
    };
}

export function UpdateShareRequestToJSON(value?: UpdateShareRequest | null): any {
    if (value == null) {
        return value;
    }
    return {
        
        'shrToken': value['shrToken'],
        'backendProxyEndpoint': value['backendProxyEndpoint'],
        'addAccessGrants': value['addAccessGrants'],
        'removeAccessGrants': value['removeAccessGrants'],
    };
}
