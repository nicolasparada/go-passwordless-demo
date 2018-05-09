import http from '../http.js';

const template = document.createElement('template')
template.innerHTML = `
<div class="container">
    <h1>Passwordless Demo</h1>

    <h2>Access</h2>

    <form id="access-form">
        <input type="email" id="email-input" name="email" placeholder="Email" required>
        <button type="submit">Send Magic Link</button>
    </form>
</div>
`

export default function welcomePageHandler() {
    const page = /** @type {DocumentFragment} */ (template.content.cloneNode(true))

    page.getElementById('access-form')
        .addEventListener('submit', onAccessFormSubmit)

    page.getElementById('email-input')
        .addEventListener('input', cleanInputError)

    return page
}

/**
 * @param {Event} ev
 */
function onAccessFormSubmit(ev) {
    ev.preventDefault()

    const form = /** @type {HTMLFormElement} */ (ev.currentTarget)
    const input = /** @type {HTMLInputElement} */ (form['email'])
    const button = /** @type {HTMLButtonElement} */ (form.querySelector('[type=submit]'))

    const email = input.value

    input.disabled = true
    button.disabled = true
    sendMagicLink(email).catch(err => {
        if (err.statusCode === 404) {
            if (wantToCreateAccount())
                runCreateUserProgram(email)
        } else if ('email' in err.body) {
            input.setCustomValidity(err.body.email)
            setTimeout(() => {
                if ('reportValidity' in input)
                    input['reportValidity']()
            }, 0)
        } else {
            alert(err.message)
        }
    }).then(() => {
        input.disabled = false
        button.disabled = false
    })
}

/**
 * @param {string} email
 */
function sendMagicLink(email) {
    return http.post('/api/passwordless/start', {
        email,
        redirectUri: location.origin + '/callback',
    }).then(() => {
        alert('Magic link sent. Go check your email.')
    })
}

function wantToCreateAccount() {
    return confirm("No user found with that email. Do you want to create an account?")
}

/**
 * @param {string} email
 * @param {string=} username
 */
function runCreateUserProgram(email, username) {
    username = prompt("Enter username", username)
    if (username === null)
        return

    http.post('/api/users', { email, username })
        .then(res => res.body)
        .then(user => sendMagicLink(user.email))
        .catch(err => {
            if ('email' in err.body) {
                alert(err.body.email)
            } else if ('username' in err.body) {
                alert(err.body.username)
                runCreateUserProgram(email, username)
            } else {
                alert(err.message)
            }
        })
}

/**
 * @param {Event} ev
 */
function cleanInputError(ev) {
    const input = /** @type {HTMLInputElement} */ (ev.currentTarget)
    input.setCustomValidity('')
}

/**
 * @typedef User
 * @property {string} id
 * @property {string} username
 * @property {string} email
 */
