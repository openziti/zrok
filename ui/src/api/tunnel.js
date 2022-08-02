/** @module tunnel */
// Auto-generated, edits will be overwritten
import * as gateway from './gateway'

/**
 * @param {object} options Optional options
 * @param {module:types.tunnelRequest} [options.body] 
 * @return {Promise<module:types.tunnelResponse>} tunnel created
 */
export function tunnel(options) {
  if (!options) options = {}
  const parameters = {
    body: {
      body: options.body
    }
  }
  return gateway.request(tunnelOperation, parameters)
}

/**
 * @param {object} options Optional options
 * @param {module:types.untunnelRequest} [options.body] 
 * @return {Promise<object>} tunnel removed
 */
export function untunnel(options) {
  if (!options) options = {}
  const parameters = {
    body: {
      body: options.body
    }
  }
  return gateway.request(untunnelOperation, parameters)
}

const tunnelOperation = {
  path: '/tunnel',
  contentTypes: ['application/zrok.v1+json'],
  method: 'post',
  security: [
    {
      id: 'key'
    }
  ]
}

const untunnelOperation = {
  path: '/untunnel',
  contentTypes: ['application/zrok.v1+json'],
  method: 'delete',
  security: [
    {
      id: 'key'
    }
  ]
}
