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
  CreateIdentity201Response,
  DisableRequest,
  EnableRequest,
} from '../models/index';
import {
    CreateIdentity201ResponseFromJSON,
    CreateIdentity201ResponseToJSON,
    DisableRequestFromJSON,
    DisableRequestToJSON,
    EnableRequestFromJSON,
    EnableRequestToJSON,
} from '../models/index';

export interface DisableOperationRequest {
    body?: DisableRequest;
}

export interface EnableOperationRequest {
    body?: EnableRequest;
}

/**
 * 
 */
export class EnvironmentApi extends runtime.BaseAPI {

    /**
     */
    async disableRaw(requestParameters: DisableOperationRequest, initOverrides?: RequestInit | runtime.InitOverrideFunction): Promise<runtime.ApiResponse<void>> {
        const queryParameters: any = {};

        const headerParameters: runtime.HTTPHeaders = {};

        headerParameters['Content-Type'] = 'application/zrok.v1+json';

        if (this.configuration && this.configuration.apiKey) {
            headerParameters["x-token"] = await this.configuration.apiKey("x-token"); // key authentication
        }


        let urlPath = `/disable`;

        const response = await this.request({
            path: urlPath,
            method: 'POST',
            headers: headerParameters,
            query: queryParameters,
            body: DisableRequestToJSON(requestParameters['body']),
        }, initOverrides);

        return new runtime.VoidApiResponse(response);
    }

    /**
     */
    async disable(requestParameters: DisableOperationRequest = {}, initOverrides?: RequestInit | runtime.InitOverrideFunction): Promise<void> {
        await this.disableRaw(requestParameters, initOverrides);
    }

    /**
     */
    async enableRaw(requestParameters: EnableOperationRequest, initOverrides?: RequestInit | runtime.InitOverrideFunction): Promise<runtime.ApiResponse<CreateIdentity201Response>> {
        const queryParameters: any = {};

        const headerParameters: runtime.HTTPHeaders = {};

        headerParameters['Content-Type'] = 'application/zrok.v1+json';

        if (this.configuration && this.configuration.apiKey) {
            headerParameters["x-token"] = await this.configuration.apiKey("x-token"); // key authentication
        }


        let urlPath = `/enable`;

        const response = await this.request({
            path: urlPath,
            method: 'POST',
            headers: headerParameters,
            query: queryParameters,
            body: EnableRequestToJSON(requestParameters['body']),
        }, initOverrides);

        return new runtime.JSONApiResponse(response, (jsonValue) => CreateIdentity201ResponseFromJSON(jsonValue));
    }

    /**
     */
    async enable(requestParameters: EnableOperationRequest = {}, initOverrides?: RequestInit | runtime.InitOverrideFunction): Promise<CreateIdentity201Response> {
        const response = await this.enableRaw(requestParameters, initOverrides);
        return await response.value();
    }

}
