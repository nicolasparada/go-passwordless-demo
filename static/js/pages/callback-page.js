import http from '../http.js';

const template = document.createElement('template')
template.innerHTML = `
<div class="container">
    <h1>Authenticating you...</h1>
</div>
`

export default function CallbackPage() {
    const page = template.content.cloneNode(true)

    if (typeof jwt === 'string' && isDate(expiresAt)) {
        fetchAuthUser(jwt).then(JSON.stringify).then(authUser => {
            localStorage.setItem('jwt', jwt)
            localStorage.setItem('auth_user', authUser)
            localStorage.setItem('expires_at', expiresAt)
        }).catch(err => {
            alert(err.body.message || err.body || err.message)
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
    return http
        .get('/api/auth_user', { authorization: `Bearer ${token}` })
        .then(res => res.body)
}

/**
 * @param {string} str
 */
function isDate(str) {
    return typeof str === 'string'
        && !isNaN(new Date(str).valueOf())
}

/**
 * @typedef AuthUser
 * @property {string} id
 * @property {string} username
 * @property {string} email
 */
