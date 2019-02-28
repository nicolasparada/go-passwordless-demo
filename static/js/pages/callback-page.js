import http from '../http.js';

export default async function callbackPage() {
    const f = new URLSearchParams(location.hash.substr(1))
    for (const [k, v] of f.entries()) {
        f.set(decodeURIComponent(k), decodeURIComponent(v))
    }
    const token = f.get('token')
    const expiresAt = f.get('expires_at')

    try {
        if (token === null || expiresAt === null) {
            throw new Error('Invalid URL')
        }

        const authUser = await getAuthUser(token)

        localStorage.setItem('auth_user', JSON.stringify(authUser))
        localStorage.setItem('token', token)
        localStorage.setItem('expires_at', expiresAt)
    } catch (err) {
        alert(err.message)
    } finally {
        location.replace('/')
    }
}

/**
 * @param {string} token
 */
function getAuthUser(token) {
    return http.get('/api/auth_user', { authorization: `Bearer ${token}` })
}
