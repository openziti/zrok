/** @module admin */
// Auto-generated, edits will be overwritten
import * as gateway from './gateway'

/**
 * @param {object} options Optional options
 * @param {module:types.createFrontendRequest} [options.body] 
 * @return {Promise<module:types.createFrontendResponse>} frontend created
 */
export function createFrontend(options) {
  if (!options) options = {}
  const parameters = {
    body: {
      body: options.body
    }
  }
  return gateway.request(createFrontendOperation, parameters)
}

const createFrontendOperation = {
  path: '/frontend',
  contentTypes: ['application/zrok.v1+json'],
  method: 'post',
  security: [
    {
      id: 'key'
    }
  ]
}
