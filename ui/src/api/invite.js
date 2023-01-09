/** @module invite */
// Auto-generated, edits will be overwritten
import * as gateway from './gateway'

/**
 * @param {object} options Optional options
 * @param {module:types.inviteGenerateRequest} [options.body] 
 * @return {Promise<object>} invitation tokens created
 */
export function inviteGenerate(options) {
  if (!options) options = {}
  const parameters = {
    body: {
      body: options.body
    }
  }
  return gateway.request(inviteGenerateOperation, parameters)
}

const inviteGenerateOperation = {
  path: '/invite/generate',
  contentTypes: ['application/zrok.v1+json'],
  method: 'post'
}
