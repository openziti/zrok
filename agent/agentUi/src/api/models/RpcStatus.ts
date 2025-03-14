/* tslint:disable */
/* eslint-disable */
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
 */

import { mapValues } from '../runtime';
import type { ProtobufAny } from './ProtobufAny';
import {
    ProtobufAnyFromJSON,
    ProtobufAnyFromJSONTyped,
    ProtobufAnyToJSON,
    ProtobufAnyToJSONTyped,
} from './ProtobufAny';

/**
 * 
 * @export
 * @interface RpcStatus
 */
export interface RpcStatus {
    /**
     * 
     * @type {number}
     * @memberof RpcStatus
     */
    code?: number;
    /**
     * 
     * @type {string}
     * @memberof RpcStatus
     */
    message?: string;
    /**
     * 
     * @type {Array<ProtobufAny>}
     * @memberof RpcStatus
     */
    details?: Array<ProtobufAny>;
}

/**
 * Check if a given object implements the RpcStatus interface.
 */
export function instanceOfRpcStatus(value: object): value is RpcStatus {
    return true;
}

export function RpcStatusFromJSON(json: any): RpcStatus {
    return RpcStatusFromJSONTyped(json, false);
}

export function RpcStatusFromJSONTyped(json: any, ignoreDiscriminator: boolean): RpcStatus {
    if (json == null) {
        return json;
    }
    return {
        
        'code': json['code'] == null ? undefined : json['code'],
        'message': json['message'] == null ? undefined : json['message'],
        'details': json['details'] == null ? undefined : ((json['details'] as Array<any>).map(ProtobufAnyFromJSON)),
    };
}

export function RpcStatusToJSON(json: any): RpcStatus {
    return RpcStatusToJSONTyped(json, false);
}

export function RpcStatusToJSONTyped(value?: RpcStatus | null, ignoreDiscriminator: boolean = false): any {
    if (value == null) {
        return value;
    }

    return {
        
        'code': value['code'],
        'message': value['message'],
        'details': value['details'] == null ? undefined : ((value['details'] as Array<any>).map(ProtobufAnyToJSON)),
    };
}

