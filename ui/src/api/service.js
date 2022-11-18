/** @module service */
// Auto-generated, edits will be overwritten
import * as gateway from './gateway'

/**
 * @param {object} options Optional options
 * @param {module:types.shareRequest} [options.body] 
 * @return {Promise<module:types.shareResponse>} service created
 */
export function share(options) {
  if (!options) options = {}
  const parameters = {
    body: {
      body: options.body
    }
  }
  return gateway.request(shareOperation, parameters)
}

/**
 * @param {object} options Optional options
 * @param {module:types.unshareRequest} [options.body] 
 * @return {Promise<object>} service removed
 */
export function unshare(options) {
  if (!options) options = {}
  const parameters = {
    body: {
      body: options.body
    }
  }
  return gateway.request(unshareOperation, parameters)
}

const shareOperation = {
  path: '/share',
  contentTypes: ['application/zrok.v1+json'],
  method: 'post',
  security: [
    {
      id: 'key'
    }
  ]
}

const unshareOperation = {
  path: '/unshare',
  contentTypes: ['application/zrok.v1+json'],
  method: 'delete',
  security: [
    {
      id: 'key'
    }
  ]
}
