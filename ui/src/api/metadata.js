/** @module metadata */
// Auto-generated, edits will be overwritten
import * as gateway from './gateway'

/**
 * @param {object} options Optional options
 * @param {object} [options.body] 
 * @return {Promise<module:types.environmentServices>} ok
 */
export function getEnvironmentDetail(options) {
  if (!options) options = {}
  const parameters = {
    body: {
      body: options.body
    }
  }
  return gateway.request(getEnvironmentDetailOperation, parameters)
}

/**
 */
export function overview() {
  return gateway.request(overviewOperation)
}

/**
 */
export function version() {
  return gateway.request(versionOperation)
}

const getEnvironmentDetailOperation = {
  path: '/detail/environment',
  contentTypes: ['application/zrok.v1+json'],
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

const versionOperation = {
  path: '/version',
  method: 'get'
}
