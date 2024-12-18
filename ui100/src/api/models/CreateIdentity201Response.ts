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
 * @interface CreateIdentity201Response
 */
export interface CreateIdentity201Response {
    /**
     * 
     * @type {string}
     * @memberof CreateIdentity201Response
     */
    identity?: string;
    /**
     * 
     * @type {string}
     * @memberof CreateIdentity201Response
     */
    cfg?: string;
}

/**
 * Check if a given object implements the CreateIdentity201Response interface.
 */
export function instanceOfCreateIdentity201Response(value: object): value is CreateIdentity201Response {
    return true;
}

export function CreateIdentity201ResponseFromJSON(json: any): CreateIdentity201Response {
    return CreateIdentity201ResponseFromJSONTyped(json, false);
}

export function CreateIdentity201ResponseFromJSONTyped(json: any, ignoreDiscriminator: boolean): CreateIdentity201Response {
    if (json == null) {
        return json;
    }
    return {
        
        'identity': json['identity'] == null ? undefined : json['identity'],
        'cfg': json['cfg'] == null ? undefined : json['cfg'],
    };
}

export function CreateIdentity201ResponseToJSON(value?: CreateIdentity201Response | null): any {
    if (value == null) {
        return value;
    }
    return {
        
        'identity': value['identity'],
        'cfg': value['cfg'],
    };
}
