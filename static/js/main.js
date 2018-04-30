import { isAuthenticated } from './auth.js';
import { importWithCache } from './dynamic-import.js';
import Router from './router.js';

function view(name) {
    return (...args) => importWithCache(`/js/pages/${name}-page.js`)
        .then(m => m.default)
        .then(h => h(...args))
}

function guard(fn1, fn2 = view('not-found')) {
    return (...args) => isAuthenticated()
        ? fn1(...args)
        : fn2(...args)
}

const router = new Router()

router.handle('/', guard(view('home'), view('welcome')))
router.handle('/callback', view('callback'))
router.handle(/^\//, view('not-found'))

const disconnectEvent = new CustomEvent('disconnect')
const pageOutlet = document.body
const loadingHTML = pageOutlet.innerHTML
let currentPage

async function render() {
    if (currentPage instanceof Node) {
        pageOutlet.innerHTML = loadingHTML
        currentPage.dispatchEvent(disconnectEvent)
    }
    currentPage = await router.exec(decodeURI(location.pathname))
    pageOutlet.innerHTML = ''
    pageOutlet.appendChild(currentPage)
}

render()
addEventListener('click', Router.delegateClicks)
addEventListener('popstate', render)
