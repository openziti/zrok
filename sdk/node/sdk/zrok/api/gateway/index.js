// Auto-generated, edits will be overwritten
import spec from './spec'

export class ServiceError extends Error {}

let options = {}

export function init(serviceOptions) {
	options = serviceOptions
}

export function request(op, parameters, attempt) {
	if (!attempt) attempt = 1;
	return acquireRights(op, spec, options)
		.then(rights => {
			parameters = parameters || {}
			const baseUrl = getBaseUrl(spec)
			let reqInfo = { parameters, baseUrl }
			if (options.processRequest) {
				reqInfo = options.processRequest(op, reqInfo)
			}
			const req = buildRequest(op, reqInfo.baseUrl, reqInfo.parameters, rights)
			return makeFetchRequest(req)
				.then(res => processResponse(req, res, attempt, options), e => processError(req, e))
				.then(outcome => outcome.retry ? request(op, parameters, attempt + 1) : outcome.res)
		})
}

function acquireRights(op, spec, options) {
	if (op.security && options.getAuthorization) {
		return op.security.reduce((promise, security) => {
			return promise.then(rights => {
				const securityDefinition = spec.securityDefinitions[security.id]
				return options.getAuthorization(security, securityDefinition, op)
							.then(auth => {
								rights[security.id] = auth
								return rights
							})
			})
		}, Promise.resolve({}))
	}
	return Promise.resolve({})
}

function makeFetchRequest(req) {
	let fetchOptions = {
		compress: true,
		method: (req.method || 'get').toUpperCase(),
		headers: req.headers,
		body: req.body ? JSON.stringify(req.body) : undefined
	}

	if (options.fetchOptions) {
		const opts = options.fetchOptions
		const headers = opts.headers
			? Object.assign(fetchOptions.headers, opts.headers)
			: fetchOptions.headers

		fetchOptions = Object.assign({}, fetchOptions, opts)
		fetchOptions.headers = headers
	}

	let promise = fetch(req.url, fetchOptions)
	return promise
}

function buildRequest(op, baseUrl, parameters, rights) {
	let paramGroups = groupParams(op, parameters)
	paramGroups = applyAuthorization(paramGroups, rights, spec)
	const url = buildUrl(op, baseUrl, paramGroups, spec)
	const headers = buildHeaders(op, paramGroups)
	const body = buildBody(parameters.body)
	return {
		method: op.method,
		url,
		headers,
		body
	}
}

function groupParams(op, parameters) {
	const groups = ['header', 'path', 'query', 'formData'].reduce((groups, name) => {
		groups[name] = formatParamsGroup(groups[name])
		return groups
	}, parameters)
	if (!groups.header) groups.header = {}
	return groups
}

function formatParamsGroup(groups) {
	return Object.keys(groups || {}).reduce((g, name) => {
		const param = groups[name]
		if (param !== undefined) {
			g[name] = formatParam(param)
		}
		return g
	}, {})
}

function formatParam(param) {
	if (param === undefined || param === null) return ''
	else if (param instanceof Date) return param.toJSON()
	else if (Array.isArray(param)) return param
	else return param.toString()
}

function buildUrl(op, baseUrl, parameters, spec) {
	let url = `${baseUrl}${op.path}`
	if (parameters.path) {
		url = Object.keys(parameters.path)
			.reduce((url, name) => url.replace(`{${name}}`, parameters.path[name]), url)
	}
	const query = createQueryString(parameters.query)
	return url + query
}

function getBaseUrl(spec) {
  return options.url || `${spec.schemes[0] || 'https'}://${spec.host}${spec.basePath}`
}

function createQueryParam(name, value) {
	const v = formatParam(value)
	if (v && typeof v === 'string') return `${name}=${encodeURIComponent(v)}`
	return name;
}

function createQueryString(query) {
	const names = Object.keys(query || {})
	if (!names.length) return ''
	const params = names.map(name => ({name, value: query[name]}))
		.reduce((acc, value) => {
			if (Array.isArray(value.value)) {
				return acc.concat(value.value)
			} else {
				acc.push(createQueryParam(value.name, value.value))
				return acc
			}
		}, [])
	return '?' + params.sort().join('&')
}

function buildHeaders(op, parameters) {
	const headers = {}

	let accepts
	if (op.accepts && op.accepts.length) accepts = op.accepts
	else if (spec.accepts && spec.accepts.length) accepts = spec.accepts
	else accepts = [ 'application/json' ]

	headers.Accept = accepts.join(', ')

	let contentType
	if (op.contentTypes && op.contentTypes[0]) contentType = op.contentTypes[0]
	else if (spec.contentTypes && spec.contentTypes[0]) contentType = spec.contentTypes[0]
	if (contentType) headers['Content-Type'] = contentType

	return Object.assign(headers, parameters.header)
}

function buildBody(bodyParams) {
	if (bodyParams) {
		if (bodyParams.body) return bodyParams.body
		const key = Object.keys(bodyParams)[0]
		if (key) return bodyParams[key]
	}
	return undefined
}

function resolveAuthHeaderName(headerName){
	if (options.authorizationHeader && headerName.toLowerCase() === 'authorization') {
		return options.authorizationHeader
	} else {
		return headerName
	}
}

function applyAuthorization(req, rights, spec) {
	Object.keys(rights).forEach(name => {
		const rightsInfo = rights[name]
		const definition = spec.securityDefinitions[name]
		switch (definition.type) {
			case 'basic':
				const creds = `${rightsInfo.username}:${rightsInfo.password}`
				const token = (typeof window !== 'undefined' && window.btoa)
					? window.btoa(creds)
					: new Buffer(creds).toString('base64')
				req.header[resolveAuthHeaderName('Authorization')] = `Basic ${token}`
				break
			case 'oauth2':
				req.header[resolveAuthHeaderName('Authorization')] = `Bearer ${rightsInfo.token}`
				break
			case 'apiKey':
				if (definition.in === 'header') {
					req.header[resolveAuthHeaderName(definition.name)] = rightsInfo.apiKey
				} else if (definition.in === 'query') {
					req.query[definition.name] = rightsInfo.apiKey
				} else {
					throw new Error(`Api key must be in header or query not '${definition.in}'`)
				}
				break
			default:
				throw new Error(`Security definition type '${definition.type}' not supported`)
		}
	})
	return req
}

function processResponse(req, response, attempt, options) {
	const format = response.ok ? formatResponse : formatServiceError
	const contentType = response.headers.get('content-type') || ''

	let parse
	if (response.status === 204) {
		parse = Promise.resolve()
	} else if (~contentType.indexOf('json')) {
		parse = response.json()
	} else if (~contentType.indexOf('octet-stream')) {
		parse = response.blob()
	} else if (~contentType.indexOf('text')) {
		parse = response.text()
	} else {
		parse = Promise.resolve()
	}

	return parse
		.then(data => format(response, data, options))
		.then(res => {
			if (options.processResponse) return options.processResponse(req, res, attempt)
			else return Promise.resolve({ res })
		})
}

function formatResponse(response, data, options) {
	return { raw: response, data }
}

function formatServiceError(response, data, options) {
	if (options.formatServiceError) {
		data = options.formatServiceError(response, data)
	} else {
		const serviceError = new ServiceError()
		if (data) {
			if (typeof data === 'string') serviceError.message = data
			else {
				if (data.message) serviceError.message = data.message
				if (data.body) serviceError.body = data.body
				else serviceError.body = data
			}

			if (data.code) serviceError.code = data.code
		} else {
			serviceError.message = response.statusText
		}
		serviceError.status = response.status
		data = serviceError
	}
	return { raw: response, data, error: true }
}

function processError(req, error) {
	const { processError } = options
	const res = { res: { raw: {}, data: error, error: true } }

	return Promise.resolve(processError ? processError(req, res) : res)
}

const COLLECTION_DELIM = { csv: ',', multi: '&', pipes: '|', ssv: ' ', tsv: '\t' }

export function formatArrayParam(array, format, name) {
	if (!array) return
	if (format === 'multi') return array.map(value => createQueryParam(name, value))
	const delim = COLLECTION_DELIM[format]
	if (!delim) throw new Error(`Invalid collection format '${format}'`)
	return array.map(formatParam).join(delim)
}

export function formatDate(date, format) {
	if (!date) return
	const str = date.toISOString()
	return (format === 'date') ? str.split('T')[0] : str
}
