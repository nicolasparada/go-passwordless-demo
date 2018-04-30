import http from '../http.js';

const template = document.createElement('template')
template.innerHTML = `
<div class="container">
    <h1>Authenticating you...</h1>
</div>
`

export default function callbackPageHandler() {
    const page = template.content.cloneNode(true)

    const f = new URLSearchParams(decodeURIComponent(location.hash.substr(1)))
    const jwt = f.get('jwt')
    const expiresAt = f.get('expires_at')

    if (typeof jwt === 'string' && isDate(expiresAt)) {
        http.get('/api/auth_user', { authorization: `Bearer ${jwt}` })
            .then(res => res.body)
            .then(authUser => {
                localStorage.setItem('jwt', jwt)
                localStorage.setItem('auth_user', JSON.stringify(authUser))
                localStorage.setItem('expires_at', expiresAt)
            })
            .catch(err => {
                alert(err.body.message || err.body || err.message)
            })
            .then(() => {
                location.replace('/')
            })
    } else {
        alert('Invalid URL')
        location.replace('/')
    }

    return page
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
