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
    return typeof localStorage.getItem('jwt') === 'string' && getAuthUser() !== null
}

export function logout() {
    localStorage.clear()
    location.reload()
}

/**
 * @typedef AuthUser
 * @property {string} id
 * @property {string} username
 * @property {string} email
 */
