import localVarRequest from 'request';

export * from './access201Response';
export * from './accessRequest';
export * from './addOrganizationMemberRequest';
export * from './authUser';
export * from './changePasswordRequest';
export * from './clientVersionCheckRequest';
export * from './configuration';
export * from './createFrontend201Response';
export * from './createFrontendRequest';
export * from './createIdentity201Response';
export * from './createIdentityRequest';
export * from './createOrganization201Response';
export * from './createOrganizationRequest';
export * from './disableRequest';
export * from './enableRequest';
export * from './environment';
export * from './environmentAndResources';
export * from './frontend';
export * from './getSparklines200Response';
export * from './getSparklinesRequest';
export * from './inviteRequest';
export * from './inviteTokenGenerateRequest';
export * from './listFrontends200ResponseInner';
export * from './listMemberships200Response';
export * from './listMemberships200ResponseMembershipsInner';
export * from './listOrganizationMembers200Response';
export * from './listOrganizationMembers200ResponseMembersInner';
export * from './listOrganizations200Response';
export * from './listOrganizations200ResponseOrganizationsInner';
export * from './loginRequest';
export * from './metrics';
export * from './metricsSample';
export * from './overview';
export * from './principal';
export * from './regenerateAccountToken200Response';
export * from './regenerateAccountTokenRequest';
export * from './registerRequest';
export * from './removeOrganizationMemberRequest';
export * from './resetPasswordRequest';
export * from './share';
export * from './shareRequest';
export * from './shareResponse';
export * from './sparkDataSample';
export * from './unaccessRequest';
export * from './unshareRequest';
export * from './updateAccessRequest';
export * from './updateFrontendRequest';
export * from './updateShareRequest';
export * from './verify200Response';
export * from './verifyRequest';
export * from './versionInventory200Response';

import * as fs from 'fs';

export interface RequestDetailedFile {
    value: Buffer;
    options?: {
        filename?: string;
        contentType?: string;
    }
}

export type RequestFile = string | Buffer | fs.ReadStream | RequestDetailedFile;


import { Access201Response } from './access201Response';
import { AccessRequest } from './accessRequest';
import { AddOrganizationMemberRequest } from './addOrganizationMemberRequest';
import { AuthUser } from './authUser';
import { ChangePasswordRequest } from './changePasswordRequest';
import { ClientVersionCheckRequest } from './clientVersionCheckRequest';
import { Configuration } from './configuration';
import { CreateFrontend201Response } from './createFrontend201Response';
import { CreateFrontendRequest } from './createFrontendRequest';
import { CreateIdentity201Response } from './createIdentity201Response';
import { CreateIdentityRequest } from './createIdentityRequest';
import { CreateOrganization201Response } from './createOrganization201Response';
import { CreateOrganizationRequest } from './createOrganizationRequest';
import { DisableRequest } from './disableRequest';
import { EnableRequest } from './enableRequest';
import { Environment } from './environment';
import { EnvironmentAndResources } from './environmentAndResources';
import { Frontend } from './frontend';
import { GetSparklines200Response } from './getSparklines200Response';
import { GetSparklinesRequest } from './getSparklinesRequest';
import { InviteRequest } from './inviteRequest';
import { InviteTokenGenerateRequest } from './inviteTokenGenerateRequest';
import { ListFrontends200ResponseInner } from './listFrontends200ResponseInner';
import { ListMemberships200Response } from './listMemberships200Response';
import { ListMemberships200ResponseMembershipsInner } from './listMemberships200ResponseMembershipsInner';
import { ListOrganizationMembers200Response } from './listOrganizationMembers200Response';
import { ListOrganizationMembers200ResponseMembersInner } from './listOrganizationMembers200ResponseMembersInner';
import { ListOrganizations200Response } from './listOrganizations200Response';
import { ListOrganizations200ResponseOrganizationsInner } from './listOrganizations200ResponseOrganizationsInner';
import { LoginRequest } from './loginRequest';
import { Metrics } from './metrics';
import { MetricsSample } from './metricsSample';
import { Overview } from './overview';
import { Principal } from './principal';
import { RegenerateAccountToken200Response } from './regenerateAccountToken200Response';
import { RegenerateAccountTokenRequest } from './regenerateAccountTokenRequest';
import { RegisterRequest } from './registerRequest';
import { RemoveOrganizationMemberRequest } from './removeOrganizationMemberRequest';
import { ResetPasswordRequest } from './resetPasswordRequest';
import { Share } from './share';
import { ShareRequest } from './shareRequest';
import { ShareResponse } from './shareResponse';
import { SparkDataSample } from './sparkDataSample';
import { UnaccessRequest } from './unaccessRequest';
import { UnshareRequest } from './unshareRequest';
import { UpdateAccessRequest } from './updateAccessRequest';
import { UpdateFrontendRequest } from './updateFrontendRequest';
import { UpdateShareRequest } from './updateShareRequest';
import { Verify200Response } from './verify200Response';
import { VerifyRequest } from './verifyRequest';
import { VersionInventory200Response } from './versionInventory200Response';

/* tslint:disable:no-unused-variable */
let primitives = [
                    "string",
                    "boolean",
                    "double",
                    "integer",
                    "long",
                    "float",
                    "number",
                    "any"
                 ];

let enumsMap: {[index: string]: any} = {
        "CreateFrontendRequest.PermissionModeEnum": CreateFrontendRequest.PermissionModeEnum,
        "ShareRequest.ShareModeEnum": ShareRequest.ShareModeEnum,
        "ShareRequest.BackendModeEnum": ShareRequest.BackendModeEnum,
        "ShareRequest.OauthProviderEnum": ShareRequest.OauthProviderEnum,
        "ShareRequest.PermissionModeEnum": ShareRequest.PermissionModeEnum,
}

let typeMap: {[index: string]: any} = {
    "Access201Response": Access201Response,
    "AccessRequest": AccessRequest,
    "AddOrganizationMemberRequest": AddOrganizationMemberRequest,
    "AuthUser": AuthUser,
    "ChangePasswordRequest": ChangePasswordRequest,
    "ClientVersionCheckRequest": ClientVersionCheckRequest,
    "Configuration": Configuration,
    "CreateFrontend201Response": CreateFrontend201Response,
    "CreateFrontendRequest": CreateFrontendRequest,
    "CreateIdentity201Response": CreateIdentity201Response,
    "CreateIdentityRequest": CreateIdentityRequest,
    "CreateOrganization201Response": CreateOrganization201Response,
    "CreateOrganizationRequest": CreateOrganizationRequest,
    "DisableRequest": DisableRequest,
    "EnableRequest": EnableRequest,
    "Environment": Environment,
    "EnvironmentAndResources": EnvironmentAndResources,
    "Frontend": Frontend,
    "GetSparklines200Response": GetSparklines200Response,
    "GetSparklinesRequest": GetSparklinesRequest,
    "InviteRequest": InviteRequest,
    "InviteTokenGenerateRequest": InviteTokenGenerateRequest,
    "ListFrontends200ResponseInner": ListFrontends200ResponseInner,
    "ListMemberships200Response": ListMemberships200Response,
    "ListMemberships200ResponseMembershipsInner": ListMemberships200ResponseMembershipsInner,
    "ListOrganizationMembers200Response": ListOrganizationMembers200Response,
    "ListOrganizationMembers200ResponseMembersInner": ListOrganizationMembers200ResponseMembersInner,
    "ListOrganizations200Response": ListOrganizations200Response,
    "ListOrganizations200ResponseOrganizationsInner": ListOrganizations200ResponseOrganizationsInner,
    "LoginRequest": LoginRequest,
    "Metrics": Metrics,
    "MetricsSample": MetricsSample,
    "Overview": Overview,
    "Principal": Principal,
    "RegenerateAccountToken200Response": RegenerateAccountToken200Response,
    "RegenerateAccountTokenRequest": RegenerateAccountTokenRequest,
    "RegisterRequest": RegisterRequest,
    "RemoveOrganizationMemberRequest": RemoveOrganizationMemberRequest,
    "ResetPasswordRequest": ResetPasswordRequest,
    "Share": Share,
    "ShareRequest": ShareRequest,
    "ShareResponse": ShareResponse,
    "SparkDataSample": SparkDataSample,
    "UnaccessRequest": UnaccessRequest,
    "UnshareRequest": UnshareRequest,
    "UpdateAccessRequest": UpdateAccessRequest,
    "UpdateFrontendRequest": UpdateFrontendRequest,
    "UpdateShareRequest": UpdateShareRequest,
    "Verify200Response": Verify200Response,
    "VerifyRequest": VerifyRequest,
    "VersionInventory200Response": VersionInventory200Response,
}

export class ObjectSerializer {
    public static findCorrectType(data: any, expectedType: string) {
        if (data == undefined) {
            return expectedType;
        } else if (primitives.indexOf(expectedType.toLowerCase()) !== -1) {
            return expectedType;
        } else if (expectedType === "Date") {
            return expectedType;
        } else {
            if (enumsMap[expectedType]) {
                return expectedType;
            }

            if (!typeMap[expectedType]) {
                return expectedType; // w/e we don't know the type
            }

            // Check the discriminator
            let discriminatorProperty = typeMap[expectedType].discriminator;
            if (discriminatorProperty == null) {
                return expectedType; // the type does not have a discriminator. use it.
            } else {
                if (data[discriminatorProperty]) {
                    var discriminatorType = data[discriminatorProperty];
                    if(typeMap[discriminatorType]){
                        return discriminatorType; // use the type given in the discriminator
                    } else {
                        return expectedType; // discriminator did not map to a type
                    }
                } else {
                    return expectedType; // discriminator was not present (or an empty string)
                }
            }
        }
    }

    public static serialize(data: any, type: string) {
        if (data == undefined) {
            return data;
        } else if (primitives.indexOf(type.toLowerCase()) !== -1) {
            return data;
        } else if (type.lastIndexOf("Array<", 0) === 0) { // string.startsWith pre es6
            let subType: string = type.replace("Array<", ""); // Array<Type> => Type>
            subType = subType.substring(0, subType.length - 1); // Type> => Type
            let transformedData: any[] = [];
            for (let index = 0; index < data.length; index++) {
                let datum = data[index];
                transformedData.push(ObjectSerializer.serialize(datum, subType));
            }
            return transformedData;
        } else if (type === "Date") {
            return data.toISOString();
        } else {
            if (enumsMap[type]) {
                return data;
            }
            if (!typeMap[type]) { // in case we dont know the type
                return data;
            }

            // Get the actual type of this object
            type = this.findCorrectType(data, type);

            // get the map for the correct type.
            let attributeTypes = typeMap[type].getAttributeTypeMap();
            let instance: {[index: string]: any} = {};
            for (let index = 0; index < attributeTypes.length; index++) {
                let attributeType = attributeTypes[index];
                instance[attributeType.baseName] = ObjectSerializer.serialize(data[attributeType.name], attributeType.type);
            }
            return instance;
        }
    }

    public static deserialize(data: any, type: string) {
        // polymorphism may change the actual type.
        type = ObjectSerializer.findCorrectType(data, type);
        if (data == undefined) {
            return data;
        } else if (primitives.indexOf(type.toLowerCase()) !== -1) {
            return data;
        } else if (type.lastIndexOf("Array<", 0) === 0) { // string.startsWith pre es6
            let subType: string = type.replace("Array<", ""); // Array<Type> => Type>
            subType = subType.substring(0, subType.length - 1); // Type> => Type
            let transformedData: any[] = [];
            for (let index = 0; index < data.length; index++) {
                let datum = data[index];
                transformedData.push(ObjectSerializer.deserialize(datum, subType));
            }
            return transformedData;
        } else if (type === "Date") {
            return new Date(data);
        } else {
            if (enumsMap[type]) {// is Enum
                return data;
            }

            if (!typeMap[type]) { // dont know the type
                return data;
            }
            let instance = new typeMap[type]();
            let attributeTypes = typeMap[type].getAttributeTypeMap();
            for (let index = 0; index < attributeTypes.length; index++) {
                let attributeType = attributeTypes[index];
                instance[attributeType.name] = ObjectSerializer.deserialize(data[attributeType.baseName], attributeType.type);
            }
            return instance;
        }
    }
}

export interface Authentication {
    /**
    * Apply authentication settings to header and query params.
    */
    applyToRequest(requestOptions: localVarRequest.Options): Promise<void> | void;
}

export class HttpBasicAuth implements Authentication {
    public username: string = '';
    public password: string = '';

    applyToRequest(requestOptions: localVarRequest.Options): void {
        requestOptions.auth = {
            username: this.username, password: this.password
        }
    }
}

export class HttpBearerAuth implements Authentication {
    public accessToken: string | (() => string) = '';

    applyToRequest(requestOptions: localVarRequest.Options): void {
        if (requestOptions && requestOptions.headers) {
            const accessToken = typeof this.accessToken === 'function'
                            ? this.accessToken()
                            : this.accessToken;
            requestOptions.headers["Authorization"] = "Bearer " + accessToken;
        }
    }
}

export class ApiKeyAuth implements Authentication {
    public apiKey: string = '';

    constructor(private location: string, private paramName: string) {
    }

    applyToRequest(requestOptions: localVarRequest.Options): void {
        if (this.location == "query") {
            (<any>requestOptions.qs)[this.paramName] = this.apiKey;
        } else if (this.location == "header" && requestOptions && requestOptions.headers) {
            requestOptions.headers[this.paramName] = this.apiKey;
        } else if (this.location == 'cookie' && requestOptions && requestOptions.headers) {
            if (requestOptions.headers['Cookie']) {
                requestOptions.headers['Cookie'] += '; ' + this.paramName + '=' + encodeURIComponent(this.apiKey);
            }
            else {
                requestOptions.headers['Cookie'] = this.paramName + '=' + encodeURIComponent(this.apiKey);
            }
        }
    }
}

export class OAuth implements Authentication {
    public accessToken: string = '';

    applyToRequest(requestOptions: localVarRequest.Options): void {
        if (requestOptions && requestOptions.headers) {
            requestOptions.headers["Authorization"] = "Bearer " + this.accessToken;
        }
    }
}

export class VoidAuth implements Authentication {
    public username: string = '';
    public password: string = '';

    applyToRequest(_: localVarRequest.Options): void {
        // Do nothing
    }
}

export type Interceptor = (requestOptions: localVarRequest.Options) => (Promise<void> | void);
