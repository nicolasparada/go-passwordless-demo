/**
 * @returns {AuthUser=}
 */
export function getAuthUser() {
    const authUserItem = localStorage.getItem('auth_user')
    const expiresAtItem = localStorage.getItem('expires_at')

    if (authUserItem !== null && expiresAtItem !== null) {
        const expiresAt = new Date(expiresAtItem)
        if (!isNaN(expiresAt.valueOf()) && expiresAt > new Date()) {
            try {
                return JSON.parse(authUserItem)
            } catch (_) { }
        }
    }

    return null
}

export function isAuthenticated() {
    return localStorage.getItem('jwt') !== null && getAuthUser() !== null
}

/**
 * @typedef AuthUser
 * @property {string} id
 * @property {string} username
 * @property {string} email
 */
