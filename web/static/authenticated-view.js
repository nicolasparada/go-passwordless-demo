import { parseResponse } from "./http.js"

const tmpl = document.createElement("template")
tmpl.innerHTML = `
    <main class="container">
        <h1>Welcome</h1>
        <p>Logged-in as <span data-ref="username"></span> 😉</p>
        <br>
        <button id="logout-btn">Logout</button>
    </main>
`

/**
 * @param {import("./auth.js").Auth} auth
 */
export function authenticatedView(auth) {
    const view = /** @type {DocumentFragment} */ (tmpl.content.cloneNode(true))
    view.querySelector("[data-ref=username]").textContent = auth.user.username
    view.querySelector("#logout-btn").addEventListener("click", onLogoutBtnClick)

    setTimeout(() => {
        fetchAuthUser(auth.token).then(authUser => {
            console.log(authUser)
        }).catch(err => {
            console.error(err)
        })
    })

    return view
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
 * @param {string} token
 * @returns {Promise<import("./auth.js").User>}
 */
function fetchAuthUser(token) {
    return fetch("/api/auth-user", {
        method: "GET",
        headers: {
            "authorization": "Bearer " + token,
        },
    }).then(parseResponse)
}
