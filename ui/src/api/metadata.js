/** @module metadata */
// Auto-generated, edits will be overwritten
import * as gateway from './gateway'

/**
 */
export function listIdentities() {
  return gateway.request(listIdentitiesOperation)
}

/**
 */
export function version() {
  return gateway.request(versionOperation)
}

const listIdentitiesOperation = {
  path: '/listIdentities',
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
