import { parseResponse } from "./http.js"

const tmpl = document.createElement("template")
tmpl.innerHTML = `
    <main class="container">
        <h1>Login</h1>
        <form name="login-form">
            <div class="btn-grp">
                <label for="email-input">Email:</label>
                <input id="email-input" name="email" autocomplete="email" placeholder="Email" required>
            </div>
            <button>Login</button>
        </form>
    </main>
`

export function guestView() {
    const view = /** @type {DocumentFragment} */ (tmpl.content.cloneNode(true))
    view.querySelector("[name=login-form]").addEventListener("submit", onLoginFormSubmit)
    return view
}
/**
 * @param {Event} ev
 */
function onLoginFormSubmit(ev) {
    ev.preventDefault()

    const form = /** @type {HTMLFormElement} */ (ev.currentTarget)
    const input = form.querySelector("input")
    const button = form.querySelector("button")

    const email = input.value

    input.disabled = true
    button.disabled = true

    sendMagicLink(email).then(() => {
        alert("Magic link sent. Go check your inbox to login")
    }).catch(err => {
        console.error(err)
        alert(err.message)
    }).finally(() => {
        input.disabled = false
        button.disabled = false
    })
}

/**
 * @param {string} email
 * @param {string=} redirectURI
 * @returns {Promise<void>}
 */
function sendMagicLink(email, redirectURI = location.origin + "/login-callback") {
    return fetch("/api/send-magic-link", {
        method: "POST",
        headers: {
            "content-type": "application/json; charset=utf-8",
        },
        body: JSON.stringify({ email, redirectURI }),
    }).then(parseResponse)
}
