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
 * @interface ListFrontends200ResponseInner
 */
export interface ListFrontends200ResponseInner {
    /**
     * 
     * @type {string}
     * @memberof ListFrontends200ResponseInner
     */
    frontendToken?: string;
    /**
     * 
     * @type {string}
     * @memberof ListFrontends200ResponseInner
     */
    zId?: string;
    /**
     * 
     * @type {string}
     * @memberof ListFrontends200ResponseInner
     */
    urlTemplate?: string;
    /**
     * 
     * @type {string}
     * @memberof ListFrontends200ResponseInner
     */
    publicName?: string;
    /**
     * 
     * @type {number}
     * @memberof ListFrontends200ResponseInner
     */
    createdAt?: number;
    /**
     * 
     * @type {number}
     * @memberof ListFrontends200ResponseInner
     */
    updatedAt?: number;
}

/**
 * Check if a given object implements the ListFrontends200ResponseInner interface.
 */
export function instanceOfListFrontends200ResponseInner(value: object): value is ListFrontends200ResponseInner {
    return true;
}

export function ListFrontends200ResponseInnerFromJSON(json: any): ListFrontends200ResponseInner {
    return ListFrontends200ResponseInnerFromJSONTyped(json, false);
}

export function ListFrontends200ResponseInnerFromJSONTyped(json: any, ignoreDiscriminator: boolean): ListFrontends200ResponseInner {
    if (json == null) {
        return json;
    }
    return {
        
        'frontendToken': json['frontendToken'] == null ? undefined : json['frontendToken'],
        'zId': json['zId'] == null ? undefined : json['zId'],
        'urlTemplate': json['urlTemplate'] == null ? undefined : json['urlTemplate'],
        'publicName': json['publicName'] == null ? undefined : json['publicName'],
        'createdAt': json['createdAt'] == null ? undefined : json['createdAt'],
        'updatedAt': json['updatedAt'] == null ? undefined : json['updatedAt'],
    };
}

export function ListFrontends200ResponseInnerToJSON(json: any): ListFrontends200ResponseInner {
    return ListFrontends200ResponseInnerToJSONTyped(json, false);
}

export function ListFrontends200ResponseInnerToJSONTyped(value?: ListFrontends200ResponseInner | null, ignoreDiscriminator: boolean = false): any {
    if (value == null) {
        return value;
    }

    return {
        
        'frontendToken': value['frontendToken'],
        'zId': value['zId'],
        'urlTemplate': value['urlTemplate'],
        'publicName': value['publicName'],
        'createdAt': value['createdAt'],
        'updatedAt': value['updatedAt'],
    };
}

