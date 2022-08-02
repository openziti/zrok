/** @module metadata */
// Auto-generated, edits will be overwritten
import * as gateway from './gateway'

/**
 */
export function version() {
  return gateway.request(versionOperation)
}

const versionOperation = {
  path: '/version',
  method: 'get'
}
