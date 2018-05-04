import { getAuthUser } from '../auth.js';
import http from '../http.js';

export default function homePageHandler() {
    const authUser = getAuthUser()
    const template = document.createElement('template')
    template.innerHTML = `
    <div class="container">
        <h1>Passwordless Demo</h1>

        <p>Welcome back, ${authUser.username} 👋</p>
        <button id="logout-button">Logout</button>
    </div>
    `

    const page = template.content
    const logoutButton = /** @type {HTMLButtonElement} */ (page.getElementById('logout-button'))

    /**
     * @param {MouseEvent} ev
     */
    const onLogoutButtonClick = ev => {
        logoutButton.disabled = true
        http.post("/api/logout").then(() => {
            localStorage.clear()
            location.reload()
        }).catch(err => {
            alert(err.body.message || err.body || err.message)
            logoutButton.disabled = false
        })
    }

    logoutButton.addEventListener('click', onLogoutButtonClick)

    return page
}
