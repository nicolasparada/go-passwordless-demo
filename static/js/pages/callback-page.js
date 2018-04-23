import http from '../http.js'

const template = document.createElement('template')
template.innerHTML = `
    <h1>Authenticating you...</h1>
`

export default function CallbackPage() {
    const page = template.content.cloneNode(true)

    if (typeof jwt === 'string' && isDate(expiresAt)) {
        fetchAuthUser(jwt).then(stringify).then(authUser => {
            localStorage.setItem('jwt', jwt)
            localStorage.setItem('auth_user', authUser)
            localStorage.setItem('expires_at', expiresAt)
        }).catch(err => {
            alert(err.message)
        }).then(() => {
            location.replace('/')
        })
    } else {
        alert('Invalid URL')
        location.replace('/')
    }

    return page
}

const f = new URLSearchParams(decodeURIComponent(location.hash.substr(1)))
const jwt = f.get('jwt')
const expiresAt = f.get('expires_at')

/**
 * @param {string} token
 * @returns {Promise<AuthUser>}
 */
function fetchAuthUser(token) {
    return http.get('/api/auth_user', { Authorization: `Bearer ${token}` })
}

/**
 * @param {string} str
 */
function isDate(str) {
    return typeof str === 'string' && !isNaN(new Date(str).valueOf())
}

function stringify(x) {
    try {
        return Promise.resolve(JSON.stringify(x))
    } catch (err) {
        return Promise.reject(err)
    }
}

/**
 * @typedef AuthUser
 * @property {string} id
 * @property {string} username
 * @property {string} email
 */
