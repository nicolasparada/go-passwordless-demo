/**
 * @returns {AuthUser=}
 */
export function getAuthUser() {
    if (!isAuthenticated()) {
        return null
    }

    const authUser = localStorage.getItem('auth_user')
    if (authUser === null) {
        return null
    }

    try {
        return JSON.parse(authUser)
    } catch (_) {
        return null
    }
}

export function isAuthenticated() {
    const token = localStorage.getItem('token')
    const expiresAtItem = localStorage.getItem('expires_at')
    if (token === null || expiresAtItem === null) {
        return false
    }

    const expiresAt = new Date(expiresAtItem)
    if (isNaN(expiresAt.valueOf()) || expiresAt <= new Date()) {
        return false
    }

    return true
}

/**
 * @param {function} fn1
 * @param {function} fn2
 */
export function guard(fn1, fn2) {
    return (...args) => isAuthenticated()
        ? fn1(...args)
        : fn2(...args)
}

/**
 * @typedef AuthUser
 * @property {string} id
 * @property {string} username
 * @property {string} email
 */
