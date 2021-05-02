const guestViewTmpl = document.createElement("template")
guestViewTmpl.innerHTML = `
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

const authenticatedViewTmpl = document.createElement("template")
authenticatedViewTmpl.innerHTML = `
    <main class="container">
        <h1>Welcome</h1>
        <p>Logged-in as <span data-ref="username"></span> 😉</p>
        <br>
        <button id="logout-btn">Logout</button>
    </main>
`

void async function main() {
    const data = new URLSearchParams(location.hash.substring(1))
    if (data.has("error")) {
        const errMsg = decodeURIComponent(data.get("error"))
        alert(errMsg)

        if (!data.has("retry_uri")) {
            location.assign("/")
            return
        }

        if (errMsg === "user not found") {
            const ok = confirm("do you want to create a new account?")
            if (!ok) {
                location.assign("/")
                return
            }
        }


        const username = prompt("Username")
        if (username === null) {
            location.assign("/")
            return
        }

        const retryURI = new URL(decodeURIComponent(data.get("retry_uri")), location.origin)
        retryURI.searchParams.set("username", username)
        location.replace(retryURI.toString())
        return
    }

    if (["token", "expires_at", "user.id", "user.email", "user.username"].every(k => data.has(k))) {
        saveLocalAuth(data)
        location.replace("/")
        return
    }

    const auth = localAuth()
    if (auth === null) {
        render(guestView())
        return
    }

    render(authenticatedView(auth))
}()

function guestView() {
    const view = /** @type {DocumentFragment} */ (guestViewTmpl.content.cloneNode(true))
    view.querySelector("[name=login-form]").addEventListener("submit", onLoginFormSubmit)
    return view
}

/**
 * @param {Auth} auth
 */
function authenticatedView(auth) {
    const view = /** @type {DocumentFragment} */ (authenticatedViewTmpl.content.cloneNode(true))
    view.querySelector("[data-ref=username]").textContent = auth.user.username
    view.querySelector("#logout-btn").addEventListener("click", onLogoutBtnClick)
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
 * @param {Event} ev
 */
function onLogoutBtnClick(ev) {
    const btn = /** @type {HTMLButtonElement} */ (ev.currentTarget)
    btn.disabled = true
    localStorage.removeItem("auth")
    location.replace("/")
}

/**
 * @param {URLSearchParams} data
 */
function saveLocalAuth(data) {
    localStorage.setItem("auth", JSON.stringify({
        user: {
            id: decodeURIComponent(data.get("user.id")),
            email: decodeURIComponent(data.get("user.email")),
            username: decodeURIComponent(data.get("user.username")),
        },
        token: decodeURIComponent(data.get("token")),
        expiresAt: decodeURIComponent(data.get("expires_at")),
    }))
}

/**
 * @typedef {object} User
 * @prop {string} id
 * @prop {string} email
 * @prop {string} username
 *
 * @typedef {object} Auth
 * @prop {User} user
 * @prop {string} token
 * @prop {Date} expiresAt
 *
 * @returns {Auth|null}
 */
function localAuth() {
    const authItem = localStorage.getItem("auth")
    if (authItem === null) {
        null
    }

    try {
        const auth = JSON.parse(authItem)
        if (typeof auth !== "object"
            || auth === null
            || typeof auth.token !== "string"
            || typeof auth.expiresAt !== "string") {
            return null
        }

        auth.expiresAt = new Date(auth.expiresAt)
        if (isNaN(auth.expiresAt.valueOf()) || auth.expiresAt < new Date()) {
            return null
        }

        const user = auth["user"]
        if (typeof user !== "object"
            || user === null
            || typeof user.id !== "string"
            || typeof user.email !== "string"
            || typeof user.username !== "string") {
            return null
        }

        return auth
    } catch (_) {
        return null
    }
}

/**
 * @param {string} email
 * @param {string=} redirectURI
 * @returns {Promise<void>}
 */
function sendMagicLink(email, redirectURI = location.origin) {
    return fetch("/api/send-magic-link", {
        method: "POST",
        headers: {
            "content-type": "application/json; charset=utf-8",
        },
        body: JSON.stringify({ email, redirectURI }),
    }).then(parseResponse)
}

/**
 * @param {Response} resp
 */
function parseResponse(resp) {
    return resp.clone().json().catch(() => resp.text()).then(body => {
        if (!resp.ok) {
            const msg = typeof body === "string" && body !== "" ? body : resp.statusText
            const err = new Error(msg)
            return Promise.reject(err)
        }

        return body
    })
}

/**
 * @param {Node} node
 * @param {Node} target
 */
function render(node, target = document.body) {
    cleanupNode(target)
    target.appendChild(node)
}

/**
 * @param {Node} node
 */
function cleanupNode(node) {
    while (node.firstChild !== null) {
        node.removeChild(node.lastChild)
    }
}
