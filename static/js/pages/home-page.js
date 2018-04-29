import { getAuthUser } from '../auth.js';
import http from '../http.js';

const template = document.createElement('template')
template.innerHTML = `
<div class="container">
    <h1>Passwordless Demo</h1>

    <div id="authenticated" hidden>
        <button id="logout-button">Logout</button>
    </div>

    <div id="guest" hidden>
        <h2>Access</h2>

        <form id="access-form">
            <input type="email" id="email-input" placeholder="Email" required>
            <button type="submit">Send Magic Link</button>
        </form>
    </div>
</div>
`

export default function HomePage() {
    const page = /** @type {DocumentFragment} */ (template.content.cloneNode(true))

    const authUser = getAuthUser()
    const authenticated = authUser !== null
    const authenticatedDiv = page.getElementById('authenticated')
    const guestDiv = page.getElementById('guest')
    const logoutButton = /** @type {HTMLButtonElement} */ (page.getElementById('logout-button'))
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

        sendMagicLink(email).then(onMagicLinkSent).catch(err => {
            if (err.statusCode === 404) {
                if (confirm("No user found with that email. Do you want to create an account?"))
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

    if (authenticated) {
        const greeting = document.createTextNode(`Welcome back, ${authUser.username} 👋`)
        authenticatedDiv.insertBefore(greeting, authenticatedDiv.firstChild)
        authenticatedDiv.hidden = false
    } else {
        guestDiv.hidden = false
    }

    logoutButton.addEventListener('click', logout)
    accessForm.addEventListener('submit', onAccessFormSubmit)
    emailInput.addEventListener('input', cleanInputError)

    return page
}

function logout() {
    localStorage.clear()
    location.reload()
}

/**
 * @param {string} email
 */
function sendMagicLink(email) {
    return http.post('/api/passwordless/start', {
        email,
        redirectUri: location.origin + '/callback',
    }).then(() => undefined)
}

function onMagicLinkSent() {
    alert('Magic link sent. Go check your email.')
}

/**
 * @param {string} email
 * @param {string=} username
 */
function runCreateUserProgram(email, username) {
    username = prompt("Enter username", username)
    if (username === null)
        return

    createUser(email, username).then(onUserCreated).catch(err => {
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
 * @param {string} email
 * @param {string} username
 */
function createUser(email, username) {
    return http.post('/api/users', { email, username }).then(res => res.body)
}

/**
 * @param {User} user
 */
function onUserCreated(user) {
    return sendMagicLink(user.email).then(onMagicLinkSent)
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
