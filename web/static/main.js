import { getLocalAuth, loginCallback } from "./auth.js"

void async function main() {
    if (location.pathname === "/login-callback") {
        loginCallback()
        return
    }

    const auth = getLocalAuth()
    if (auth === null) {
        import("./guest-view.js").then(m => {
            update(m.guestView())
        })
        return
    }

    import("./authenticated-view.js").then(m => {
        update(m.authenticatedView(auth))
    })
}()

/**
 * @param {Node} node
 * @param {Node} target
 */
function update(node, target = document.body) {
    while (target.firstChild !== null) {
        target.removeChild(target.lastChild)
    }
    target.appendChild(node)
}
