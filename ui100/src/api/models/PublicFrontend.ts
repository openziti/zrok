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
 * @interface PublicFrontend
 */
export interface PublicFrontend {
    /**
     * 
     * @type {string}
     * @memberof PublicFrontend
     */
    token?: string;
    /**
     * 
     * @type {string}
     * @memberof PublicFrontend
     */
    zId?: string;
    /**
     * 
     * @type {string}
     * @memberof PublicFrontend
     */
    urlTemplate?: string;
    /**
     * 
     * @type {string}
     * @memberof PublicFrontend
     */
    publicName?: string;
    /**
     * 
     * @type {number}
     * @memberof PublicFrontend
     */
    createdAt?: number;
    /**
     * 
     * @type {number}
     * @memberof PublicFrontend
     */
    updatedAt?: number;
}

/**
 * Check if a given object implements the PublicFrontend interface.
 */
export function instanceOfPublicFrontend(value: object): value is PublicFrontend {
    return true;
}

export function PublicFrontendFromJSON(json: any): PublicFrontend {
    return PublicFrontendFromJSONTyped(json, false);
}

export function PublicFrontendFromJSONTyped(json: any, ignoreDiscriminator: boolean): PublicFrontend {
    if (json == null) {
        return json;
    }
    return {
        
        'token': json['token'] == null ? undefined : json['token'],
        'zId': json['zId'] == null ? undefined : json['zId'],
        'urlTemplate': json['urlTemplate'] == null ? undefined : json['urlTemplate'],
        'publicName': json['publicName'] == null ? undefined : json['publicName'],
        'createdAt': json['createdAt'] == null ? undefined : json['createdAt'],
        'updatedAt': json['updatedAt'] == null ? undefined : json['updatedAt'],
    };
}

export function PublicFrontendToJSON(value?: PublicFrontend | null): any {
    if (value == null) {
        return value;
    }
    return {
        
        'token': value['token'],
        'zId': value['zId'],
        'urlTemplate': value['urlTemplate'],
        'publicName': value['publicName'],
        'createdAt': value['createdAt'],
        'updatedAt': value['updatedAt'],
    };
}
