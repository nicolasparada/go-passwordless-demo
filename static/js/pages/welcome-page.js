import http from '../http.js';

const template = document.createElement('template')
template.innerHTML = `
<div class="container">
    <h1>Passwordless Demo</h1>

    <h2>Access</h2>

    <form id="access-form">
        <input type="email" id="email-input" placeholder="Email" required>
        <button type="submit">Send Magic Link</button>
    </form>
</div>
`

export default function welcomePageHandler() {
    const page = /** @type {DocumentFragment} */ (template.content.cloneNode(true))

    const accessForm = /** @type {HTMLFormElement} */ (page.getElementById('access-form'))
    const emailInput = /** @type {HTMLInputElement} */ (page.getElementById('email-input'))
    const accessButton = /** @type {HTMLButtonElement} */ (accessForm.querySelector('[type=submit]'))

    /**
     * @param {Event} ev
     */
    const onAccessFormSubmit = ev => {
        ev.preventDefault()

        const email = emailInput.value

        emailInput.disabled = true
        accessButton.disabled = true

        sendMagicLink(email).catch(err => {
            if (err.statusCode === 404) {
                if (wantToCreateAccount())
                    runCreateUserProgram(email)
            } else if ('email' in err.body) {
                emailInput.setCustomValidity(err.body.email)
                setTimeout(() => {
                    if ('reportValidity' in emailInput)
                        emailInput['reportValidity']()
                }, 0)
            } else {
                alert(err.body.message || err.body || err.message)
            }
        }).then(() => {
            emailInput.disabled = false
            accessButton.disabled = false
        })
    }

    accessForm.addEventListener('submit', onAccessFormSubmit)
    emailInput.addEventListener('input', cleanInputError)

    return page
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
                alert(err.body.message || err.body || err.message)
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
