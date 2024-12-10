/** @module metadata */
// Auto-generated, edits will be overwritten
import * as gateway from './gateway'

/**
 */
export function configuration() {
  return gateway.request(configurationOperation)
}

/**
 */
export function getAccountDetail() {
  return gateway.request(getAccountDetailOperation)
}

/**
 * @param {string} envZId 
 * @return {Promise<module:types.environmentAndResources>} ok
 */
export function getEnvironmentDetail(envZId) {
  const parameters = {
    path: {
      envZId
    }
  }
  return gateway.request(getEnvironmentDetailOperation, parameters)
}

/**
 * @param {number} feId 
 * @return {Promise<module:types.frontend>} ok
 */
export function getFrontendDetail(feId) {
  const parameters = {
    path: {
      feId
    }
  }
  return gateway.request(getFrontendDetailOperation, parameters)
}

/**
 * @param {string} shrToken 
 * @return {Promise<module:types.share>} ok
 */
export function getShareDetail(shrToken) {
  const parameters = {
    path: {
      shrToken
    }
  }
  return gateway.request(getShareDetailOperation, parameters)
}

/**
 */
export function listMemberships() {
  return gateway.request(listMembershipsOperation)
}

/**
 */
export function overview() {
  return gateway.request(overviewOperation)
}

/**
 * @param {string} organizationToken 
 * @param {string} accountEmail 
 * @return {Promise<module:types.overview>} ok
 */
export function orgAccountOverview(organizationToken, accountEmail) {
  const parameters = {
    path: {
      organizationToken,
      accountEmail
    }
  }
  return gateway.request(orgAccountOverviewOperation, parameters)
}

/**
 * @param {object} options Optional options
 * @param {string} [options.duration] 
 * @return {Promise<module:types.metrics>} account metrics
 */
export function getAccountMetrics(options) {
  if (!options) options = {}
  const parameters = {
    query: {
      duration: options.duration
    }
  }
  return gateway.request(getAccountMetricsOperation, parameters)
}

/**
 * @param {string} envId 
 * @param {object} options Optional options
 * @param {string} [options.duration] 
 * @return {Promise<module:types.metrics>} environment metrics
 */
export function getEnvironmentMetrics(envId, options) {
  if (!options) options = {}
  const parameters = {
    path: {
      envId
    },
    query: {
      duration: options.duration
    }
  }
  return gateway.request(getEnvironmentMetricsOperation, parameters)
}

/**
 * @param {string} shrToken 
 * @param {object} options Optional options
 * @param {string} [options.duration] 
 * @return {Promise<module:types.metrics>} share metrics
 */
export function getShareMetrics(shrToken, options) {
  if (!options) options = {}
  const parameters = {
    path: {
      shrToken
    },
    query: {
      duration: options.duration
    }
  }
  return gateway.request(getShareMetricsOperation, parameters)
}

/**
 */
export function version() {
  return gateway.request(versionOperation)
}

const configurationOperation = {
  path: '/configuration',
  method: 'get'
}

const getAccountDetailOperation = {
  path: '/detail/account',
  method: 'get',
  security: [
    {
      id: 'key'
    }
  ]
}

const getEnvironmentDetailOperation = {
  path: '/detail/environment/{envZId}',
  method: 'get',
  security: [
    {
      id: 'key'
    }
  ]
}

const getFrontendDetailOperation = {
  path: '/detail/frontend/{feId}',
  method: 'get',
  security: [
    {
      id: 'key'
    }
  ]
}

const getShareDetailOperation = {
  path: '/detail/share/{shrToken}',
  method: 'get',
  security: [
    {
      id: 'key'
    }
  ]
}

const listMembershipsOperation = {
  path: '/memberships',
  method: 'get',
  security: [
    {
      id: 'key'
    }
  ]
}

const overviewOperation = {
  path: '/overview',
  method: 'get',
  security: [
    {
      id: 'key'
    }
  ]
}

const orgAccountOverviewOperation = {
  path: '/overview/{organizationToken}/{accountEmail}',
  method: 'get',
  security: [
    {
      id: 'key'
    }
  ]
}

const getAccountMetricsOperation = {
  path: '/metrics/account',
  method: 'get',
  security: [
    {
      id: 'key'
    }
  ]
}

const getEnvironmentMetricsOperation = {
  path: '/metrics/environment/{envId}',
  method: 'get',
  security: [
    {
      id: 'key'
    }
  ]
}

const getShareMetricsOperation = {
  path: '/metrics/share/{shrToken}',
  method: 'get',
  security: [
    {
      id: 'key'
    }
  ]
}

const versionOperation = {
  path: '/version',
  method: 'get'
}
