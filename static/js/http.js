import { isAuthenticated } from './auth.js';

/**
 * @param {string} url
 * @param {{[key: string]: string}=} headers
 */
function get(url, headers) {
    return fetch(url, {
        headers: Object.assign(getAuthHeader(), headers),
    }).then(handleResponse)
}

/**
 * @param {string} url
 * @param {{[key: string]: any}=} body
 * @param {{[key: string]: string}=} headers
 */
function post(url, body, headers) {
    return fetch(url, {
        method: 'POST',
        headers: Object.assign(getAuthHeader(), { 'content-type': 'application/json' }, headers),
        body: JSON.stringify(body),
    }).then(handleResponse)
}

function getAuthHeader() {
    return isAuthenticated()
        ? { authorization: `Bearer ${localStorage.getItem('jwt')}` }
        : {}
}

/**
 * @param {Response} res
 */
export async function handleResponse(res) {
    const body = await res.clone().json().catch(() => res.text())
    const response = {
        statusCode: res.status,
        statusText: res.statusText,
        headers: res.headers,
        body,
    }
    if (!res.ok) {
        const message = typeof body === 'object' && body !== null && 'message' in body
            ? body.message
            : typeof body === 'string' && body !== ''
                ? body
                : res.statusText
        const err = new Error(message)
        throw Object.assign(err, response)
    }
    return response
}

export default {
    get,
    post,
}
