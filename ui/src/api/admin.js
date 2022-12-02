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

/**
 * @param {object} options Optional options
 * @param {module:types.updateFrontendRequest} [options.body] 
 * @return {Promise<object>} frontend updated
 */
export function updateFrontend(options) {
  if (!options) options = {}
  const parameters = {
    body: {
      body: options.body
    }
  }
  return gateway.request(updateFrontendOperation, parameters)
}

/**
 * @param {object} options Optional options
 * @param {module:types.deleteFrontendRequest} [options.body] 
 * @return {Promise<object>} frontend deleted
 */
export function deleteFrontend(options) {
  if (!options) options = {}
  const parameters = {
    body: {
      body: options.body
    }
  }
  return gateway.request(deleteFrontendOperation, parameters)
}

/**
 */
export function listFrontends() {
  return gateway.request(listFrontendsOperation)
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

const updateFrontendOperation = {
  path: '/frontend',
  contentTypes: ['application/zrok.v1+json'],
  method: 'patch',
  security: [
    {
      id: 'key'
    }
  ]
}

const deleteFrontendOperation = {
  path: '/frontend',
  contentTypes: ['application/zrok.v1+json'],
  method: 'delete',
  security: [
    {
      id: 'key'
    }
  ]
}

const listFrontendsOperation = {
  path: '/frontends',
  method: 'get',
  security: [
    {
      id: 'key'
    }
  ]
}
