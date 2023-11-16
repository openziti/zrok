/** @module environment */
// Auto-generated, edits will be overwritten
import * as gateway from './gateway'

/**
 * @param {object} options Optional options
 * @param {module:types.enableRequest} [options.body] 
 * @return {Promise<module:types.enableResponse>} environment enabled
 */
export function enable(options) {
  if (!options) options = {}
  const parameters = {
    body: {
      body: options.body
    }
  }
  return gateway.request(enableOperation, parameters)
}

/**
 * @param {object} options Optional options
 * @param {module:types.disableRequest} [options.body] 
 * @return {Promise<object>} environment disabled
 */
export function disable(options) {
  if (!options) options = {}
  const parameters = {
    body: {
      body: options.body
    }
  }
  return gateway.request(disableOperation, parameters)
}

const enableOperation = {
  path: '/enable',
  contentTypes: ['application/zrok.v1+json'],
  method: 'post',
  security: [
    {
      id: 'key'
    }
  ]
}

const disableOperation = {
  path: '/disable',
  contentTypes: ['application/zrok.v1+json'],
  method: 'post',
  security: [
    {
      id: 'key'
    }
  ]
}
