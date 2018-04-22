import http from './http.js'
import { getAuthUser } from './auth.js'

const authUser = getAuthUser()
const authenticated = authUser !== null
const authenticatedDiv = document.getElementById('authenticated')
const guestDiv = document.getElementById('guest')
const logoutButton = /** @type {HTMLButtonElement} */ (document.getElementById('logout-button'))
const accessForm = /** @type {HTMLFormElement} */ (document.getElementById('access-form'))
const emailInput = /** @type {HTMLInputElement} */ (document.getElementById('email-input'))
const accessButton = /** @type {HTMLButtonElement} */ (accessForm.querySelector('[type=submit]'))

logoutButton.addEventListener('click', logout)
accessForm.addEventListener('submit', onSendMagicLinkSubmit)
emailInput.addEventListener('input', cleanInputError)

if (authenticated) {
    const greeting = document.createTextNode(`Welcome back, ${authUser.username} 👋`)
    authenticatedDiv.insertBefore(greeting, authenticatedDiv.firstChild)
    authenticatedDiv.hidden = false
} else {
    guestDiv.hidden = false
}

function logout() {
    localStorage.clear()
    location.reload()
}

/**
 * @param {Event} ev
 */
function onSendMagicLinkSubmit(ev) {
    ev.preventDefault()
    const email = emailInput.value
    emailInput.disabled = true
    accessButton.disabled = true
    sendMagicLink(email).then(onMagicLinkSent).catch(err => {
        if (err.statusCode === 404) {
            runCreateUserProgram(email)
        } else if ('email' in err) {
            emailInput.setCustomValidity(err.email)
            setTimeout(() => {
                if ('reportValidity' in emailInput)
                    emailInput['reportValidity']()
            }, 0)
        } else {
            alert(err.message)
        }
    }).then(() => {
        emailInput.disabled = false
        accessButton.disabled = false
    })
}

/**
 * @param {string} email
 * @returns {Promise<void>}
 */
function sendMagicLink(email) {
    return http.post('/api/passwordless/start', {
        email,
        redirectUri: location.origin + '/callback',
    })
}

function onMagicLinkSent() {
    alert('Magic link sent. Go check your email.')
}

/**
 * @param {string} email
 */
function runCreateUserProgram(email) {
    if (!confirm("No user found with that email. Do you want to create an account?"))
        return

    const username = prompt("Enter username")
    if (username === null) return

    createUser(email, username).then(onUserCreated).catch(err => {
        if ('email' in err) {
            alert(err.email)
        } else if ('username' in err) {
            alert(err.username)
        } else {
            alert(err.message)
        }
    })
}

/**
 * @param {string} email
 * @param {string} username
 * @returns {Promise<User>}
 */
function createUser(email, username) {
    return http.post('/api/users', { email, username })
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
