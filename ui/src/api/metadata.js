/** @module metadata */
// Auto-generated, edits will be overwritten
import * as gateway from './gateway'

/**
 * @param {string} envZId 
 * @return {Promise<module:types.environmentShares>} ok
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

const getShareDetailOperation = {
  path: '/detail/share/{shrToken}',
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
