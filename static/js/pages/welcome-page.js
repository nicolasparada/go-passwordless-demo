import http from '../http.js';

const template = document.createElement('template')
template.innerHTML = `
    <div class="container">
        <h1>Passwordless Demo</h1>

        <h2>Access</h2>

        <form id="access-form">
            <input type="email" id="email-input" name="email" placeholder="Email" autofocus required>
            <button>Send Magic Link</button>
        </form>
    </div>
`

export default function welcomePage() {
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
    const input = form.querySelector('input')
    const submitButton = form.querySelector('button')

    const email = input.value

    input.disabled = true
    submitButton.disabled = true
    sendMagicLink(email).catch(err => {
        if (err.statusCode === 404) {
            if (wantToCreateAccount()) {
                runCreateUserProgram(email)
            }
        } else if (err.body.errors && 'email' in err.body.errors) {
            input.setCustomValidity(err.body.errors.email)
            setTimeout(() => {
                if ('reportValidity' in input) {
                    input['reportValidity']()
                }
            }, 0)
        } else {
            alert(err.message)
        }
    }).then(() => {
        input.disabled = false
        submitButton.disabled = false
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
    if (username === null) {
        return
    }

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
