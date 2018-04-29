import { getAuthUser } from '../auth.js';


export default function HomePage() {
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

    logoutButton.addEventListener('click', logout)

    return page
}

function logout() {
    localStorage.clear()
    location.reload()
}
