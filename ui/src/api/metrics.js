/** @module metrics */
// Auto-generated, edits will be overwritten
import * as gateway from './gateway'

/**
 */
export function getAccountMetrics() {
  return gateway.request(getAccountMetricsOperation)
}

/**
 * @param {string} envId 
 * @return {Promise<module:types.metrics>} environment metrics
 */
export function getEnvironmentMetrics(envId) {
  const parameters = {
    path: {
      envId
    }
  }
  return gateway.request(getEnvironmentMetricsOperation, parameters)
}

/**
 * @param {string} shrToken 
 * @return {Promise<module:types.metrics>} share metrics
 */
export function getShareMetrics(shrToken) {
  const parameters = {
    path: {
      shrToken
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
