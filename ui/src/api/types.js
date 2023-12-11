/** @module types */
// Auto-generated, edits will be overwritten

/**
 * @typedef accessRequest
 * @memberof module:types
 * 
 * @property {string} envZId 
 * @property {string} shrToken 
 */

/**
 * @typedef accessResponse
 * @memberof module:types
 * 
 * @property {string} frontendToken 
 * @property {string} backendMode 
 */

/**
 * @typedef authUser
 * @memberof module:types
 * 
 * @property {string} username 
 * @property {string} password 
 */

/**
 * @typedef configuration
 * @memberof module:types
 * 
 * @property {string} version 
 * @property {string} touLink 
 * @property {boolean} invitesOpen 
 * @property {boolean} requiresInviteToken 
 * @property {string} inviteTokenContact 
 * @property {module:types.passwordRequirements} passwordRequirements 
 */

/**
 * @typedef createFrontendRequest
 * @memberof module:types
 * 
 * @property {string} zId 
 * @property {string} url_template 
 * @property {string} public_name 
 */

/**
 * @typedef createFrontendResponse
 * @memberof module:types
 * 
 * @property {string} token 
 */

/**
 * @typedef deleteFrontendRequest
 * @memberof module:types
 * 
 * @property {string} frontendToken 
 */

/**
 * @typedef disableRequest
 * @memberof module:types
 * 
 * @property {string} identity 
 */

/**
 * @typedef enableRequest
 * @memberof module:types
 * 
 * @property {string} description 
 * @property {string} host 
 */

/**
 * @typedef enableResponse
 * @memberof module:types
 * 
 * @property {string} identity 
 * @property {string} cfg 
 */

/**
 * @typedef environment
 * @memberof module:types
 * 
 * @property {string} description 
 * @property {string} host 
 * @property {string} address 
 * @property {string} zId 
 * @property {module:types.sparkData} activity 
 * @property {boolean} limited 
 * @property {number} createdAt 
 * @property {number} updatedAt 
 */

/**
 * @typedef environmentAndResources
 * @memberof module:types
 * 
 * @property {module:types.environment} environment 
 * @property {module:types.frontends} frontends 
 * @property {module:types.shares} shares 
 */

/**
 * @typedef frontend
 * @memberof module:types
 * 
 * @property {number} id 
 * @property {string} shrToken 
 * @property {string} zId 
 * @property {number} createdAt 
 * @property {number} updatedAt 
 */

/**
 * @typedef inviteTokenGenerateRequest
 * @memberof module:types
 * 
 * @property {string[]} tokens 
 */

/**
 * @typedef inviteRequest
 * @memberof module:types
 * 
 * @property {string} email 
 * @property {string} token 
 */

/**
 * @typedef loginRequest
 * @memberof module:types
 * 
 * @property {string} email 
 * @property {string} password 
 */

/**
 * @typedef metrics
 * @memberof module:types
 * 
 * @property {string} scope 
 * @property {string} id 
 * @property {number} period 
 * @property {module:types.metricsSample[]} samples 
 */

/**
 * @typedef metricsSample
 * @memberof module:types
 * 
 * @property {number} rx 
 * @property {number} tx 
 * @property {number} timestamp 
 */

/**
 * @typedef overview
 * @memberof module:types
 * 
 * @property {boolean} accountLimited 
 * @property {module:types.environmentAndResources[]} environments 
 */

/**
 * @typedef passwordRequirements
 * @memberof module:types
 * 
 * @property {number} length 
 * @property {boolean} requireCapital 
 * @property {boolean} requireNumeric 
 * @property {boolean} requireSpecial 
 * @property {string} validSpecialCharacters 
 */

/**
 * @typedef principal
 * @memberof module:types
 * 
 * @property {number} id 
 * @property {string} email 
 * @property {string} token 
 * @property {boolean} limitless 
 * @property {boolean} admin 
 */

/**
 * @typedef publicFrontend
 * @memberof module:types
 * 
 * @property {string} token 
 * @property {string} zId 
 * @property {string} urlTemplate 
 * @property {string} publicName 
 * @property {number} createdAt 
 * @property {number} updatedAt 
 */

/**
 * @typedef registerRequest
 * @memberof module:types
 * 
 * @property {string} token 
 * @property {string} password 
 */

/**
 * @typedef registerResponse
 * @memberof module:types
 * 
 * @property {string} token 
 */

/**
 * @typedef resetPasswordRequest
 * @memberof module:types
 * 
 * @property {string} token 
 * @property {string} password 
 */

/**
 * @typedef share
 * @memberof module:types
 * 
 * @property {string} token 
 * @property {string} zId 
 * @property {string} shareMode 
 * @property {string} backendMode 
 * @property {string} frontendSelection 
 * @property {string} frontendEndpoint 
 * @property {string} backendProxyEndpoint 
 * @property {boolean} reserved 
 * @property {module:types.sparkData} activity 
 * @property {boolean} limited 
 * @property {number} createdAt 
 * @property {number} updatedAt 
 */

/**
 * @typedef shareRequest
 * @memberof module:types
 * 
 * @property {string} envZId 
 * @property {string} shareMode 
 * @property {string[]} frontendSelection 
 * @property {string} backendMode 
 * @property {string} backendProxyEndpoint 
 * @property {string} authScheme 
 * @property {module:types.authUser[]} authUsers 
 * @property {string} oauthProvider 
 * @property {string[]} oauthEmailDomains 
 * @property {string} oauthAuthorizationCheckInterval 
 * @property {boolean} reserved 
 * @property {string} uniqueName 
 */

/**
 * @typedef shareResponse
 * @memberof module:types
 * 
 * @property {string[]} frontendProxyEndpoints 
 * @property {string} shrToken 
 */

/**
 * @typedef sparkDataSample
 * @memberof module:types
 * 
 * @property {number} rx 
 * @property {number} tx 
 */

/**
 * @typedef unaccessRequest
 * @memberof module:types
 * 
 * @property {string} frontendToken 
 * @property {string} envZId 
 * @property {string} shrToken 
 */

/**
 * @typedef unshareRequest
 * @memberof module:types
 * 
 * @property {string} envZId 
 * @property {string} shrToken 
 * @property {boolean} reserved 
 */

/**
 * @typedef updateFrontendRequest
 * @memberof module:types
 * 
 * @property {string} frontendToken 
 * @property {string} publicName 
 * @property {string} urlTemplate 
 */

/**
 * @typedef updateShareRequest
 * @memberof module:types
 * 
 * @property {string} shrToken 
 * @property {string} backendProxyEndpoint 
 */

/**
 * @typedef verifyRequest
 * @memberof module:types
 * 
 * @property {string} token 
 */

/**
 * @typedef verifyResponse
 * @memberof module:types
 * 
 * @property {string} email 
 */
