/**
 * zrok
 * zrok client access
 *
 * The version of the OpenAPI document: 0.3.0
 * 
 *
 * NOTE: This class is auto generated by OpenAPI Generator (https://openapi-generator.tech).
 * https://openapi-generator.tech
 * Do not edit the class manually.
 */

import { RequestFile } from './models';

export class SparkDataSample {
    'rx'?: number;
    'tx'?: number;

    static discriminator: string | undefined = undefined;

    static attributeTypeMap: Array<{name: string, baseName: string, type: string}> = [
        {
            "name": "rx",
            "baseName": "rx",
            "type": "number"
        },
        {
            "name": "tx",
            "baseName": "tx",
            "type": "number"
        }    ];

    static getAttributeTypeMap() {
        return SparkDataSample.attributeTypeMap;
    }
}
