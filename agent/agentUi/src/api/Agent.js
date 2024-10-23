/** @module Agent */
// Auto-generated, edits will be overwritten
import * as gateway from './gateway'

/**
 * @param {object} options Optional options
 * @param {string} [options.token] 
 * @param {string} [options.bindAddress] 
 * @param {string[]} [options.responseHeaders] 
 * @return {Promise<module:types.AccessPrivateResponse>} A successful response.
 */
export function Agent_AccessPrivate(options) {
  if (!options) options = {}
  const parameters = {
    query: {
      token: options.token,
      bindAddress: options.bindAddress,
      responseHeaders: gateway.formatArrayParam(options.responseHeaders, 'multi', 'responseHeaders')
    }
  }
  return gateway.request(Agent_AccessPrivateOperation, parameters)
}

/**
 * @param {object} options Optional options
 * @param {string} [options.frontendToken] 
 * @return {Promise<module:types.ReleaseAccessResponse>} A successful response.
 */
export function Agent_ReleaseAccess(options) {
  if (!options) options = {}
  const parameters = {
    query: {
      frontendToken: options.frontendToken
    }
  }
  return gateway.request(Agent_ReleaseAccessOperation, parameters)
}

/**
 * @param {object} options Optional options
 * @param {string} [options.token] 
 * @return {Promise<module:types.ReleaseShareResponse>} A successful response.
 */
export function Agent_ReleaseShare(options) {
  if (!options) options = {}
  const parameters = {
    query: {
      token: options.token
    }
  }
  return gateway.request(Agent_ReleaseShareOperation, parameters)
}

/**
 * @param {object} options Optional options
 * @param {string} [options.target] 
 * @param {string} [options.backendMode] 
 * @param {boolean} [options.insecure] 
 * @param {boolean} [options.closed] 
 * @param {string[]} [options.accessGrants] 
 * @return {Promise<module:types.SharePrivateResponse>} A successful response.
 */
export function Agent_SharePrivate(options) {
  if (!options) options = {}
  const parameters = {
    query: {
      target: options.target,
      backendMode: options.backendMode,
      insecure: options.insecure,
      closed: options.closed,
      accessGrants: gateway.formatArrayParam(options.accessGrants, 'multi', 'accessGrants')
    }
  }
  return gateway.request(Agent_SharePrivateOperation, parameters)
}

/**
 * @param {object} options Optional options
 * @param {string} [options.target] 
 * @param {string[]} [options.basicAuth] 
 * @param {string[]} [options.frontendSelection] 
 * @param {string} [options.backendMode] 
 * @param {boolean} [options.insecure] 
 * @param {string} [options.oauthProvider] 
 * @param {string[]} [options.oauthEmailAddressPatterns] 
 * @param {string} [options.oauthCheckInterval] 
 * @param {boolean} [options.closed] 
 * @param {string[]} [options.accessGrants] 
 * @return {Promise<module:types.SharePublicResponse>} A successful response.
 */
export function Agent_SharePublic(options) {
  if (!options) options = {}
  const parameters = {
    query: {
      target: options.target,
      basicAuth: gateway.formatArrayParam(options.basicAuth, 'multi', 'basicAuth'),
      frontendSelection: gateway.formatArrayParam(options.frontendSelection, 'multi', 'frontendSelection'),
      backendMode: options.backendMode,
      insecure: options.insecure,
      oauthProvider: options.oauthProvider,
      oauthEmailAddressPatterns: gateway.formatArrayParam(options.oauthEmailAddressPatterns, 'multi', 'oauthEmailAddressPatterns'),
      oauthCheckInterval: options.oauthCheckInterval,
      closed: options.closed,
      accessGrants: gateway.formatArrayParam(options.accessGrants, 'multi', 'accessGrants')
    }
  }
  return gateway.request(Agent_SharePublicOperation, parameters)
}

/**
 */
export function Agent_Status() {
  return gateway.request(Agent_StatusOperation)
}

/**
 */
export function Agent_Version() {
  return gateway.request(Agent_VersionOperation)
}

const Agent_AccessPrivateOperation = {
  path: '/v1/agent/accessPrivate',
  method: 'post'
}

const Agent_ReleaseAccessOperation = {
  path: '/v1/agent/releaseAccess',
  method: 'post'
}

const Agent_ReleaseShareOperation = {
  path: '/v1/agent/releaseShare',
  method: 'post'
}

const Agent_SharePrivateOperation = {
  path: '/v1/agent/sharePrivate',
  method: 'post'
}

const Agent_SharePublicOperation = {
  path: '/v1/agent/sharePublic',
  method: 'post'
}

const Agent_StatusOperation = {
  path: '/v1/agent/status',
  method: 'get'
}

const Agent_VersionOperation = {
  path: '/v1/agent/version',
  method: 'get'
}
