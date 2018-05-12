import { isAuthenticated } from './auth.js';

/**
 * @param {string} url
 * @param {{[key: string]: string}=} headers
 */
function get(url, headers) {
    return fetch(url, {
        headers: Object.assign(getAuthHeader(), headers),
        credentials: 'include',
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
        credentials: 'include',
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
        url: res.url,
        statusCode: res.status,
        statusText: res.statusText,
        headers: res.headers,
        body,
    }
    if (!res.ok) throw Object.assign(
        new Error(body.message || body || res.statusText),
        response
    )
    return response
}

export default {
    get,
    post,
}
