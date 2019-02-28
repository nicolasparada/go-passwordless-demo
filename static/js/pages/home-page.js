import { getAuthUser } from '../auth.js';

export default function homePage() {
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
    page.getElementById('logout-button').onclick = onLogoutClick
    return page
}

function onLogoutClick() {
    localStorage.clear()
    location.reload()
}
