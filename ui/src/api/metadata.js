/** @module metadata */
// Auto-generated, edits will be overwritten
import * as gateway from './gateway'

/**
 * @param {string} envZId 
 * @return {Promise<module:types.environmentServices>} ok
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
  path: '/detail/environment/{envZId}',
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
