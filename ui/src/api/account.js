/** @module account */
// Auto-generated, edits will be overwritten
import * as gateway from './gateway'

/**
 * @param {object} options Optional options
 * @param {module:types.forgotPasswordRequest} [options.body] 
 * @return {Promise<object>} forgot password request created
 */
export function forgotPassword(options) {
  if (!options) options = {}
  const parameters = {
    body: {
      body: options.body
    }
  }
  return gateway.request(forgotPasswordOperation, parameters)
}

/**
 * @param {object} options Optional options
 * @param {module:types.inviteRequest} [options.body] 
 * @return {Promise<object>} invitation created
 */
export function invite(options) {
  if (!options) options = {}
  const parameters = {
    body: {
      body: options.body
    }
  }
  return gateway.request(inviteOperation, parameters)
}

/**
 * @param {object} options Optional options
 * @param {module:types.loginRequest} [options.body] 
 * @return {Promise<module:types.loginResponse>} login successful
 */
export function login(options) {
  if (!options) options = {}
  const parameters = {
    body: {
      body: options.body
    }
  }
  return gateway.request(loginOperation, parameters)
}

/**
 * @param {object} options Optional options
 * @param {module:types.registerRequest} [options.body] 
 * @return {Promise<module:types.registerResponse>} account created
 */
export function register(options) {
  if (!options) options = {}
  const parameters = {
    body: {
      body: options.body
    }
  }
  return gateway.request(registerOperation, parameters)
}

/**
 * @param {object} options Optional options
 * @param {module:types.resetPasswordRequest} [options.body] 
 * @return {Promise<object>} password reset
 */
export function resetPassword(options) {
  if (!options) options = {}
  const parameters = {
    body: {
      body: options.body
    }
  }
  return gateway.request(resetPasswordOperation, parameters)
}

/**
 * @param {object} options Optional options
 * @param {module:types.verifyRequest} [options.body] 
 * @return {Promise<module:types.verifyResponse>} token ready
 */
export function verify(options) {
  if (!options) options = {}
  const parameters = {
    body: {
      body: options.body
    }
  }
  return gateway.request(verifyOperation, parameters)
}

const forgotPasswordOperation = {
  path: '/forgotPassword',
  contentTypes: ['application/zrok.v1+json'],
  method: 'post'
}

const inviteOperation = {
  path: '/invite',
  contentTypes: ['application/zrok.v1+json'],
  method: 'post'
}

const loginOperation = {
  path: '/login',
  contentTypes: ['application/zrok.v1+json'],
  method: 'post'
}

const registerOperation = {
  path: '/register',
  contentTypes: ['application/zrok.v1+json'],
  method: 'post'
}

const resetPasswordOperation = {
  path: '/resetPassword',
  contentTypes: ['application/zrok.v1+json'],
  method: 'post'
}

const verifyOperation = {
  path: '/verify',
  contentTypes: ['application/zrok.v1+json'],
  method: 'post'
}
