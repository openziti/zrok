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


import * as runtime from '../runtime';
import type {
  Access201Response,
  AccessRequest,
  ShareRequest,
  ShareResponse,
  UnaccessRequest,
  UnshareRequest,
  UpdateAccessRequest,
  UpdateShareRequest,
} from '../models/index';
import {
    Access201ResponseFromJSON,
    Access201ResponseToJSON,
    AccessRequestFromJSON,
    AccessRequestToJSON,
    ShareRequestFromJSON,
    ShareRequestToJSON,
    ShareResponseFromJSON,
    ShareResponseToJSON,
    UnaccessRequestFromJSON,
    UnaccessRequestToJSON,
    UnshareRequestFromJSON,
    UnshareRequestToJSON,
    UpdateAccessRequestFromJSON,
    UpdateAccessRequestToJSON,
    UpdateShareRequestFromJSON,
    UpdateShareRequestToJSON,
} from '../models/index';

export interface AccessOperationRequest {
    body?: AccessRequest;
}

export interface ShareOperationRequest {
    body?: ShareRequest;
}

export interface UnaccessOperationRequest {
    body?: UnaccessRequest;
}

export interface UnshareOperationRequest {
    body?: UnshareRequest;
}

export interface UpdateAccessOperationRequest {
    body?: UpdateAccessRequest;
}

export interface UpdateShareOperationRequest {
    body?: UpdateShareRequest;
}

/**
 * 
 */
export class ShareApi extends runtime.BaseAPI {

    /**
     */
    async accessRaw(requestParameters: AccessOperationRequest, initOverrides?: RequestInit | runtime.InitOverrideFunction): Promise<runtime.ApiResponse<Access201Response>> {
        const queryParameters: any = {};

        const headerParameters: runtime.HTTPHeaders = {};

        headerParameters['Content-Type'] = 'application/zrok.v1+json';

        if (this.configuration && this.configuration.apiKey) {
            headerParameters["x-token"] = await this.configuration.apiKey("x-token"); // key authentication
        }

        const response = await this.request({
            path: `/access`,
            method: 'POST',
            headers: headerParameters,
            query: queryParameters,
            body: AccessRequestToJSON(requestParameters['body']),
        }, initOverrides);

        return new runtime.JSONApiResponse(response, (jsonValue) => Access201ResponseFromJSON(jsonValue));
    }

    /**
     */
    async access(requestParameters: AccessOperationRequest = {}, initOverrides?: RequestInit | runtime.InitOverrideFunction): Promise<Access201Response> {
        const response = await this.accessRaw(requestParameters, initOverrides);
        return await response.value();
    }

    /**
     */
    async shareRaw(requestParameters: ShareOperationRequest, initOverrides?: RequestInit | runtime.InitOverrideFunction): Promise<runtime.ApiResponse<ShareResponse>> {
        const queryParameters: any = {};

        const headerParameters: runtime.HTTPHeaders = {};

        headerParameters['Content-Type'] = 'application/zrok.v1+json';

        if (this.configuration && this.configuration.apiKey) {
            headerParameters["x-token"] = await this.configuration.apiKey("x-token"); // key authentication
        }

        const response = await this.request({
            path: `/share`,
            method: 'POST',
            headers: headerParameters,
            query: queryParameters,
            body: ShareRequestToJSON(requestParameters['body']),
        }, initOverrides);

        return new runtime.JSONApiResponse(response, (jsonValue) => ShareResponseFromJSON(jsonValue));
    }

    /**
     */
    async share(requestParameters: ShareOperationRequest = {}, initOverrides?: RequestInit | runtime.InitOverrideFunction): Promise<ShareResponse> {
        const response = await this.shareRaw(requestParameters, initOverrides);
        return await response.value();
    }

    /**
     */
    async unaccessRaw(requestParameters: UnaccessOperationRequest, initOverrides?: RequestInit | runtime.InitOverrideFunction): Promise<runtime.ApiResponse<void>> {
        const queryParameters: any = {};

        const headerParameters: runtime.HTTPHeaders = {};

        headerParameters['Content-Type'] = 'application/zrok.v1+json';

        if (this.configuration && this.configuration.apiKey) {
            headerParameters["x-token"] = await this.configuration.apiKey("x-token"); // key authentication
        }

        const response = await this.request({
            path: `/unaccess`,
            method: 'DELETE',
            headers: headerParameters,
            query: queryParameters,
            body: UnaccessRequestToJSON(requestParameters['body']),
        }, initOverrides);

        return new runtime.VoidApiResponse(response);
    }

    /**
     */
    async unaccess(requestParameters: UnaccessOperationRequest = {}, initOverrides?: RequestInit | runtime.InitOverrideFunction): Promise<void> {
        await this.unaccessRaw(requestParameters, initOverrides);
    }

    /**
     */
    async unshareRaw(requestParameters: UnshareOperationRequest, initOverrides?: RequestInit | runtime.InitOverrideFunction): Promise<runtime.ApiResponse<void>> {
        const queryParameters: any = {};

        const headerParameters: runtime.HTTPHeaders = {};

        headerParameters['Content-Type'] = 'application/zrok.v1+json';

        if (this.configuration && this.configuration.apiKey) {
            headerParameters["x-token"] = await this.configuration.apiKey("x-token"); // key authentication
        }

        const response = await this.request({
            path: `/unshare`,
            method: 'DELETE',
            headers: headerParameters,
            query: queryParameters,
            body: UnshareRequestToJSON(requestParameters['body']),
        }, initOverrides);

        return new runtime.VoidApiResponse(response);
    }

    /**
     */
    async unshare(requestParameters: UnshareOperationRequest = {}, initOverrides?: RequestInit | runtime.InitOverrideFunction): Promise<void> {
        await this.unshareRaw(requestParameters, initOverrides);
    }

    /**
     */
    async updateAccessRaw(requestParameters: UpdateAccessOperationRequest, initOverrides?: RequestInit | runtime.InitOverrideFunction): Promise<runtime.ApiResponse<void>> {
        const queryParameters: any = {};

        const headerParameters: runtime.HTTPHeaders = {};

        headerParameters['Content-Type'] = 'application/zrok.v1+json';

        if (this.configuration && this.configuration.apiKey) {
            headerParameters["x-token"] = await this.configuration.apiKey("x-token"); // key authentication
        }

        const response = await this.request({
            path: `/access`,
            method: 'PATCH',
            headers: headerParameters,
            query: queryParameters,
            body: UpdateAccessRequestToJSON(requestParameters['body']),
        }, initOverrides);

        return new runtime.VoidApiResponse(response);
    }

    /**
     */
    async updateAccess(requestParameters: UpdateAccessOperationRequest = {}, initOverrides?: RequestInit | runtime.InitOverrideFunction): Promise<void> {
        await this.updateAccessRaw(requestParameters, initOverrides);
    }

    /**
     */
    async updateShareRaw(requestParameters: UpdateShareOperationRequest, initOverrides?: RequestInit | runtime.InitOverrideFunction): Promise<runtime.ApiResponse<void>> {
        const queryParameters: any = {};

        const headerParameters: runtime.HTTPHeaders = {};

        headerParameters['Content-Type'] = 'application/zrok.v1+json';

        if (this.configuration && this.configuration.apiKey) {
            headerParameters["x-token"] = await this.configuration.apiKey("x-token"); // key authentication
        }

        const response = await this.request({
            path: `/share`,
            method: 'PATCH',
            headers: headerParameters,
            query: queryParameters,
            body: UpdateShareRequestToJSON(requestParameters['body']),
        }, initOverrides);

        return new runtime.VoidApiResponse(response);
    }

    /**
     */
    async updateShare(requestParameters: UpdateShareOperationRequest = {}, initOverrides?: RequestInit | runtime.InitOverrideFunction): Promise<void> {
        await this.updateShareRaw(requestParameters, initOverrides);
    }

}
