/** @module metrics */
// Auto-generated, edits will be overwritten
import * as gateway from './gateway'

/**
 * @param {object} options Optional options
 * @param {number} [options.duration] 
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
 * @param {number} [options.duration] 
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
 * @param {number} [options.duration] 
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
