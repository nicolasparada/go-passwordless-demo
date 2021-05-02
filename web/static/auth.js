export function loginCallback() {
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
        setLocalAuth(data)
        location.replace("/")
        return
    }

    location.assign("/")
}

/**
 * @param {URLSearchParams} data
 */
export function setLocalAuth(data) {
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
export function getLocalAuth() {
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
    } catch (_) { }

    return null
}
