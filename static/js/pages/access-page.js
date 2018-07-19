import http from '../http.js';

const template = document.createElement('template')
template.innerHTML = /*html*/`
    <div class="container">
        <h1>Passwordless Demo</h1>

        <h2>Access</h2>

        <form id="access-form">
            <input type="email" id="email-input" name="email" placeholder="Email" autofocus required>
            <button>Send Magic Link</button>
        </form>
    </div>
`

export default function accessPage() {
    const page = /** @type {DocumentFragment} */ (template.content.cloneNode(true))
    page.getElementById('access-form').onsubmit = onAccessFormSubmit
    page.getElementById('email-input').oninput = clearValidity
    return page
}

/**
 * @param {Event} ev
 */
async function onAccessFormSubmit(ev) {
    ev.preventDefault()

    const form = /** @type {HTMLFormElement} */ (ev.currentTarget)
    const input = form.querySelector('input')
    const submitButton = form.querySelector('button')

    const email = input.value

    input.disabled = true
    submitButton.disabled = true

    try {
        await sendMagicLink(email)
        input.value = ''
    } catch (err) {
        if (err.statusCode === 404) {
            if (confirm('User not found. Want to create an account?')) {
                runRegistrationProgram(email)
            }
            return
        }

        if (err.statusCode === 422 && 'email' in err.body.errors) {
            input.setCustomValidity(err.body.errors.email)
        } else {
            alert(err.message)
        }

        setTimeout(() => {
            input.focus()
        }, 0)
    } finally {
        input.disabled = false
        submitButton.disabled = false
    }
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

/**
 * @param {string} email
 * @param {string=} username
 */
function runRegistrationProgram(email, username) {
    username = promptUsername()
    if (username === null) {
        return
    }

    http.post('/api/users', { email, username })
        .then(user => sendMagicLink(user.email))
        .catch(err => {
            if (err.statusCode === 422 && 'email' in err.body.errors) {
                alert(err.body.errors.email)
            } else if (err.statusCode === 422 && 'username' in err.body.errors) {
                alert(err.body.errors.username)
                runRegistrationProgram(email, username)
            } else {
                alert(err.message)
            }
        })
}

/**
 * @param {string=} oldValue
 */
function promptUsername(oldValue) {
    let result = prompt('Username:', oldValue)
    if (result === null) {
        return null
    }
    result = result.trim()
    if (result === '') {
        return promptUsername()
    }
    if (!/^[a-zA-Z][a-zA-Z0-9-_]{0,17}$/.test(result)) {
        alert('Invalid username. Alpha numeric and dashes allowed. Must start with a letter. 18 characters max.')
        return promptUsername(result)
    }
    return result
}

/**
 * @param {Event} ev
 */
function clearValidity(ev) {
    const input = /** @type {HTMLInputElement} */ (ev.currentTarget)
    input.setCustomValidity('')
}
