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
  ChangePasswordRequest,
  InviteRequest,
  LoginRequest,
  RegenerateAccountToken200Response,
  RegenerateAccountTokenRequest,
  RegisterRequest,
  ResetPasswordRequest,
  Verify200Response,
  VerifyRequest,
} from '../models/index';
import {
    ChangePasswordRequestFromJSON,
    ChangePasswordRequestToJSON,
    InviteRequestFromJSON,
    InviteRequestToJSON,
    LoginRequestFromJSON,
    LoginRequestToJSON,
    RegenerateAccountToken200ResponseFromJSON,
    RegenerateAccountToken200ResponseToJSON,
    RegenerateAccountTokenRequestFromJSON,
    RegenerateAccountTokenRequestToJSON,
    RegisterRequestFromJSON,
    RegisterRequestToJSON,
    ResetPasswordRequestFromJSON,
    ResetPasswordRequestToJSON,
    Verify200ResponseFromJSON,
    Verify200ResponseToJSON,
    VerifyRequestFromJSON,
    VerifyRequestToJSON,
} from '../models/index';

export interface ChangePasswordOperationRequest {
    body?: ChangePasswordRequest;
}

export interface InviteOperationRequest {
    body?: InviteRequest;
}

export interface LoginOperationRequest {
    body?: LoginRequest;
}

export interface RegenerateAccountTokenOperationRequest {
    body?: RegenerateAccountTokenRequest;
}

export interface RegisterOperationRequest {
    body?: RegisterRequest;
}

export interface ResetPasswordOperationRequest {
    body?: ResetPasswordRequest;
}

export interface ResetPasswordRequestRequest {
    body?: RegenerateAccountTokenRequest;
}

export interface VerifyOperationRequest {
    body?: VerifyRequest;
}

/**
 * 
 */
export class AccountApi extends runtime.BaseAPI {

    /**
     */
    async changePasswordRaw(requestParameters: ChangePasswordOperationRequest, initOverrides?: RequestInit | runtime.InitOverrideFunction): Promise<runtime.ApiResponse<void>> {
        const queryParameters: any = {};

        const headerParameters: runtime.HTTPHeaders = {};

        headerParameters['Content-Type'] = 'application/zrok.v1+json';

        if (this.configuration && this.configuration.apiKey) {
            headerParameters["x-token"] = await this.configuration.apiKey("x-token"); // key authentication
        }

        const response = await this.request({
            path: `/changePassword`,
            method: 'POST',
            headers: headerParameters,
            query: queryParameters,
            body: ChangePasswordRequestToJSON(requestParameters['body']),
        }, initOverrides);

        return new runtime.VoidApiResponse(response);
    }

    /**
     */
    async changePassword(requestParameters: ChangePasswordOperationRequest = {}, initOverrides?: RequestInit | runtime.InitOverrideFunction): Promise<void> {
        await this.changePasswordRaw(requestParameters, initOverrides);
    }

    /**
     */
    async inviteRaw(requestParameters: InviteOperationRequest, initOverrides?: RequestInit | runtime.InitOverrideFunction): Promise<runtime.ApiResponse<void>> {
        const queryParameters: any = {};

        const headerParameters: runtime.HTTPHeaders = {};

        headerParameters['Content-Type'] = 'application/zrok.v1+json';

        const response = await this.request({
            path: `/invite`,
            method: 'POST',
            headers: headerParameters,
            query: queryParameters,
            body: InviteRequestToJSON(requestParameters['body']),
        }, initOverrides);

        return new runtime.VoidApiResponse(response);
    }

    /**
     */
    async invite(requestParameters: InviteOperationRequest = {}, initOverrides?: RequestInit | runtime.InitOverrideFunction): Promise<void> {
        await this.inviteRaw(requestParameters, initOverrides);
    }

    /**
     */
    async loginRaw(requestParameters: LoginOperationRequest, initOverrides?: RequestInit | runtime.InitOverrideFunction): Promise<runtime.ApiResponse<string>> {
        const queryParameters: any = {};

        const headerParameters: runtime.HTTPHeaders = {};

        headerParameters['Content-Type'] = 'application/zrok.v1+json';

        const response = await this.request({
            path: `/login`,
            method: 'POST',
            headers: headerParameters,
            query: queryParameters,
            body: LoginRequestToJSON(requestParameters['body']),
        }, initOverrides);

        if (this.isJsonMime(response.headers.get('content-type'))) {
            return new runtime.JSONApiResponse<string>(response);
        } else {
            return new runtime.TextApiResponse(response) as any;
        }
    }

    /**
     */
    async login(requestParameters: LoginOperationRequest = {}, initOverrides?: RequestInit | runtime.InitOverrideFunction): Promise<string> {
        const response = await this.loginRaw(requestParameters, initOverrides);
        return await response.value();
    }

    /**
     */
    async regenerateAccountTokenRaw(requestParameters: RegenerateAccountTokenOperationRequest, initOverrides?: RequestInit | runtime.InitOverrideFunction): Promise<runtime.ApiResponse<RegenerateAccountToken200Response>> {
        const queryParameters: any = {};

        const headerParameters: runtime.HTTPHeaders = {};

        headerParameters['Content-Type'] = 'application/zrok.v1+json';

        if (this.configuration && this.configuration.apiKey) {
            headerParameters["x-token"] = await this.configuration.apiKey("x-token"); // key authentication
        }

        const response = await this.request({
            path: `/regenerateAccountToken`,
            method: 'POST',
            headers: headerParameters,
            query: queryParameters,
            body: RegenerateAccountTokenRequestToJSON(requestParameters['body']),
        }, initOverrides);

        return new runtime.JSONApiResponse(response, (jsonValue) => RegenerateAccountToken200ResponseFromJSON(jsonValue));
    }

    /**
     */
    async regenerateAccountToken(requestParameters: RegenerateAccountTokenOperationRequest = {}, initOverrides?: RequestInit | runtime.InitOverrideFunction): Promise<RegenerateAccountToken200Response> {
        const response = await this.regenerateAccountTokenRaw(requestParameters, initOverrides);
        return await response.value();
    }

    /**
     */
    async registerRaw(requestParameters: RegisterOperationRequest, initOverrides?: RequestInit | runtime.InitOverrideFunction): Promise<runtime.ApiResponse<RegenerateAccountToken200Response>> {
        const queryParameters: any = {};

        const headerParameters: runtime.HTTPHeaders = {};

        headerParameters['Content-Type'] = 'application/zrok.v1+json';

        const response = await this.request({
            path: `/register`,
            method: 'POST',
            headers: headerParameters,
            query: queryParameters,
            body: RegisterRequestToJSON(requestParameters['body']),
        }, initOverrides);

        return new runtime.JSONApiResponse(response, (jsonValue) => RegenerateAccountToken200ResponseFromJSON(jsonValue));
    }

    /**
     */
    async register(requestParameters: RegisterOperationRequest = {}, initOverrides?: RequestInit | runtime.InitOverrideFunction): Promise<RegenerateAccountToken200Response> {
        const response = await this.registerRaw(requestParameters, initOverrides);
        return await response.value();
    }

    /**
     */
    async resetPasswordRaw(requestParameters: ResetPasswordOperationRequest, initOverrides?: RequestInit | runtime.InitOverrideFunction): Promise<runtime.ApiResponse<void>> {
        const queryParameters: any = {};

        const headerParameters: runtime.HTTPHeaders = {};

        headerParameters['Content-Type'] = 'application/zrok.v1+json';

        const response = await this.request({
            path: `/resetPassword`,
            method: 'POST',
            headers: headerParameters,
            query: queryParameters,
            body: ResetPasswordRequestToJSON(requestParameters['body']),
        }, initOverrides);

        return new runtime.VoidApiResponse(response);
    }

    /**
     */
    async resetPassword(requestParameters: ResetPasswordOperationRequest = {}, initOverrides?: RequestInit | runtime.InitOverrideFunction): Promise<void> {
        await this.resetPasswordRaw(requestParameters, initOverrides);
    }

    /**
     */
    async resetPasswordRequestRaw(requestParameters: ResetPasswordRequestRequest, initOverrides?: RequestInit | runtime.InitOverrideFunction): Promise<runtime.ApiResponse<void>> {
        const queryParameters: any = {};

        const headerParameters: runtime.HTTPHeaders = {};

        headerParameters['Content-Type'] = 'application/zrok.v1+json';

        const response = await this.request({
            path: `/resetPasswordRequest`,
            method: 'POST',
            headers: headerParameters,
            query: queryParameters,
            body: RegenerateAccountTokenRequestToJSON(requestParameters['body']),
        }, initOverrides);

        return new runtime.VoidApiResponse(response);
    }

    /**
     */
    async resetPasswordRequest(requestParameters: ResetPasswordRequestRequest = {}, initOverrides?: RequestInit | runtime.InitOverrideFunction): Promise<void> {
        await this.resetPasswordRequestRaw(requestParameters, initOverrides);
    }

    /**
     */
    async verifyRaw(requestParameters: VerifyOperationRequest, initOverrides?: RequestInit | runtime.InitOverrideFunction): Promise<runtime.ApiResponse<Verify200Response>> {
        const queryParameters: any = {};

        const headerParameters: runtime.HTTPHeaders = {};

        headerParameters['Content-Type'] = 'application/zrok.v1+json';

        const response = await this.request({
            path: `/verify`,
            method: 'POST',
            headers: headerParameters,
            query: queryParameters,
            body: VerifyRequestToJSON(requestParameters['body']),
        }, initOverrides);

        return new runtime.JSONApiResponse(response, (jsonValue) => Verify200ResponseFromJSON(jsonValue));
    }

    /**
     */
    async verify(requestParameters: VerifyOperationRequest = {}, initOverrides?: RequestInit | runtime.InitOverrideFunction): Promise<Verify200Response> {
        const response = await this.verifyRaw(requestParameters, initOverrides);
        return await response.value();
    }

}
