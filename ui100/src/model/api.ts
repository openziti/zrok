import {AccountApi, Configuration, EnvironmentApi, MetadataApi, ShareApi} from "../api";
import {User} from "./user.ts";

export const getApiConfig = (user: User): Configuration => { return new Configuration({headers: {"X-TOKEN": user.token}}); }
export const getAccountApi = (user: User): AccountApi => { return new AccountApi(getApiConfig(user)); }
export const getEnvironmentApi = (user: User): EnvironmentApi => { return new EnvironmentApi(getApiConfig(user)); }
export const getMetadataApi = (user: User): MetadataApi => { return new MetadataApi(getApiConfig(user)); }
export const getShareApi = (user: User): ShareApi => { return new ShareApi(getApiConfig(user)); }