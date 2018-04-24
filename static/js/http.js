import { isAuthenticated } from './auth.js'

/**
 * @param {Response} res
 */
async function handleResponse(res) {
    const ct = res.headers.get('content-type')
    const isJSON = typeof ct === 'string' && ct.startsWith('application/json')

    let payload = await res[isJSON ? 'json' : 'text']()
    if (!isJSON) {
        try {
            payload = JSON.parse(payload)
        } catch (_) {
            payload = { message: payload }
        }
    }

    if (!res.ok) {
        const err = new Error(res.statusText)
        err['statusCode'] = res.status
        Object.assign(err, payload)
        throw err
    }

    return payload
}

/**
 * @param {string} url
 * @param {{[key: string]: string}=} headers
 */
function get(url, headers) {
    return fetch(url, { headers: Object.assign(getDefaultHeaders(), headers) })
        .then(handleResponse)
}

/**
 * @param {string} url
 * @param {(FormData|File|{[key: string]: any})=} payload
 * @param {{[key: string]: string}=} headers
 */
function post(url, payload, headers) {
    const options = {
        method: 'POST',
        headers: getDefaultHeaders(),
    }

    if (payload instanceof FormData) {
        options['body'] = payload
        options.headers['Content-Type'] = 'multipart/form-data'
    } else if (payload instanceof File) {
        options['body'] = payload
        options.headers['Content-Type'] = payload.type
    } else if (typeof payload === 'object' && payload !== null) {
        options['body'] = JSON.stringify(payload)
        options.headers['Content-Type'] = 'application/json; charset=utf-8'
    }

    if (typeof headers !== 'undefined') {
        Object.assign(options.headers, headers)
    }

    return fetch(url, options).then(handleResponse)
}

function getDefaultHeaders() {
    return isAuthenticated()
        ? { Authorization: `Bearer ${localStorage.getItem('jwt')}` }
        : {}
}

export default {
    handleResponse,
    get,
    post,
}
