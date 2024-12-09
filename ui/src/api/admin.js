/** @module admin */
// Auto-generated, edits will be overwritten
import * as gateway from './gateway'

/**
 * @param {object} options Optional options
 * @param {object} [options.body] 
 * @return {Promise<object>} created
 */
export function createAccount(options) {
  if (!options) options = {}
  const parameters = {
    body: {
      body: options.body
    }
  }
  return gateway.request(createAccountOperation, parameters)
}

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

/**
 * @param {object} options Optional options
 * @param {object} [options.body] 
 * @return {Promise<object>} ok
 */
export function grants(options) {
  if (!options) options = {}
  const parameters = {
    body: {
      body: options.body
    }
  }
  return gateway.request(grantsOperation, parameters)
}

/**
 * @param {object} options Optional options
 * @param {object} [options.body] 
 * @return {Promise<object>} created
 */
export function createIdentity(options) {
  if (!options) options = {}
  const parameters = {
    body: {
      body: options.body
    }
  }
  return gateway.request(createIdentityOperation, parameters)
}

/**
 * @param {object} options Optional options
 * @param {module:types.inviteTokenGenerateRequest} [options.body] 
 * @return {Promise<object>} invitation tokens created
 */
export function inviteTokenGenerate(options) {
  if (!options) options = {}
  const parameters = {
    body: {
      body: options.body
    }
  }
  return gateway.request(inviteTokenGenerateOperation, parameters)
}

/**
 * @param {object} options Optional options
 * @param {object} [options.body] 
 * @return {Promise<object>} organization created
 */
export function createOrganization(options) {
  if (!options) options = {}
  const parameters = {
    body: {
      body: options.body
    }
  }
  return gateway.request(createOrganizationOperation, parameters)
}

/**
 * @param {object} options Optional options
 * @param {object} [options.body] 
 * @return {Promise<object>} organization deleted
 */
export function deleteOrganization(options) {
  if (!options) options = {}
  const parameters = {
    body: {
      body: options.body
    }
  }
  return gateway.request(deleteOrganizationOperation, parameters)
}

/**
 * @param {object} options Optional options
 * @param {object} [options.body] 
 * @return {Promise<object>} member added
 */
export function addOrganizationMember(options) {
  if (!options) options = {}
  const parameters = {
    body: {
      body: options.body
    }
  }
  return gateway.request(addOrganizationMemberOperation, parameters)
}

/**
 * @param {object} options Optional options
 * @param {object} [options.body] 
 * @return {Promise<object>} list organization members
 */
export function listOrganizationMembers(options) {
  if (!options) options = {}
  const parameters = {
    body: {
      body: options.body
    }
  }
  return gateway.request(listOrganizationMembersOperation, parameters)
}

/**
 * @param {object} options Optional options
 * @param {object} [options.body] 
 * @return {Promise<object>} member removed
 */
export function removeOrganizationMember(options) {
  if (!options) options = {}
  const parameters = {
    body: {
      body: options.body
    }
  }
  return gateway.request(removeOrganizationMemberOperation, parameters)
}

const createAccountOperation = {
  path: '/account',
  contentTypes: ['application/zrok.v1+json'],
  method: 'post',
  security: [
    {
      id: 'key'
    }
  ]
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

const grantsOperation = {
  path: '/grants',
  contentTypes: ['application/zrok.v1+json'],
  method: 'post',
  security: [
    {
      id: 'key'
    }
  ]
}

const createIdentityOperation = {
  path: '/identity',
  contentTypes: ['application/zrok.v1+json'],
  method: 'post',
  security: [
    {
      id: 'key'
    }
  ]
}

const inviteTokenGenerateOperation = {
  path: '/invite/token/generate',
  contentTypes: ['application/zrok.v1+json'],
  method: 'post',
  security: [
    {
      id: 'key'
    }
  ]
}

const createOrganizationOperation = {
  path: '/organization',
  contentTypes: ['application/zrok.v1+json'],
  method: 'post',
  security: [
    {
      id: 'key'
    }
  ]
}

const deleteOrganizationOperation = {
  path: '/organization',
  contentTypes: ['application/zrok.v1+json'],
  method: 'delete',
  security: [
    {
      id: 'key'
    }
  ]
}

const addOrganizationMemberOperation = {
  path: '/organization/add',
  contentTypes: ['application/zrok.v1+json'],
  method: 'post',
  security: [
    {
      id: 'key'
    }
  ]
}

const listOrganizationMembersOperation = {
  path: '/organization/members',
  contentTypes: ['application/zrok.v1+json'],
  method: 'post',
  security: [
    {
      id: 'key'
    }
  ]
}

const removeOrganizationMemberOperation = {
  path: '/organization/remove',
  contentTypes: ['application/zrok.v1+json'],
  method: 'post',
  security: [
    {
      id: 'key'
    }
  ]
}
