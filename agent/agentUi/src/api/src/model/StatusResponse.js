/*
 * agent/agentGrpc/agent.proto
 * No description provided (generated by Swagger Codegen https://github.com/swagger-api/swagger-codegen)
 *
 * OpenAPI spec version: version not set
 *
 * NOTE: This class is auto generated by the swagger code generator program.
 * https://github.com/swagger-api/swagger-codegen.git
 *
 * Swagger Codegen version: 3.0.51
 *
 * Do not edit the class manually.
 *
 */
import {ApiClient} from '../ApiClient';
import {AccessDetail} from './AccessDetail';
import {ShareDetail} from './ShareDetail';

/**
 * The StatusResponse model module.
 * @module model/StatusResponse
 * @version version not set
 */
export class StatusResponse {
  /**
   * Constructs a new <code>StatusResponse</code>.
   * @alias module:model/StatusResponse
   * @class
   */
  constructor() {
  }

  /**
   * Constructs a <code>StatusResponse</code> from a plain JavaScript object, optionally creating a new instance.
   * Copies all relevant properties from <code>data</code> to <code>obj</code> if supplied or a new instance if not.
   * @param {Object} data The plain JavaScript object bearing properties of interest.
   * @param {module:model/StatusResponse} obj Optional instance to populate.
   * @return {module:model/StatusResponse} The populated <code>StatusResponse</code> instance.
   */
  static constructFromObject(data, obj) {
    if (data) {
      obj = obj || new StatusResponse();
      if (data.hasOwnProperty('accesses'))
        obj.accesses = ApiClient.convertToType(data['accesses'], [AccessDetail]);
      if (data.hasOwnProperty('shares'))
        obj.shares = ApiClient.convertToType(data['shares'], [ShareDetail]);
    }
    return obj;
  }
}

/**
 * @member {Array.<module:model/AccessDetail>} accesses
 */
StatusResponse.prototype.accesses = undefined;

/**
 * @member {Array.<module:model/ShareDetail>} shares
 */
StatusResponse.prototype.shares = undefined;
