export * from './accountApi';
import { AccountApi } from './accountApi';
export * from './adminApi';
import { AdminApi } from './adminApi';
export * from './environmentApi';
import { EnvironmentApi } from './environmentApi';
export * from './metadataApi';
import { MetadataApi } from './metadataApi';
export * from './shareApi';
import { ShareApi } from './shareApi';
import * as http from 'http';

export class HttpError extends Error {
    constructor (public response: http.IncomingMessage, public body: any, public statusCode?: number) {
        super('HTTP request failed');
        this.name = 'HttpError';
    }
}

export { RequestFile } from '../model/models';

export const APIS = [AccountApi, AdminApi, EnvironmentApi, MetadataApi, ShareApi];
