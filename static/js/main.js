import { Router, dispatchPopStateOnClick } from './router.js'
import dynamicImport from './dynamic-import.js'

const pageFnsCache = new Map()
async function loadPageFn(name) {
    if (pageFnsCache.has(name))
        return pageFnsCache.get(name)
    const pageFn = await dynamicImport(`/js/pages/${name}-page.js`).then(m => m.default)
    pageFnsCache.set(name, pageFn)
    return pageFn
}

function view(name) {
    return (...args) => loadPageFn(name).then(page => page(...args))
}

const router = new Router()

router.handle('/', view('home'))
router.handle('/callback', view('callback'))
router.handle(/^\//, view('not-found'))

const pageOutlet = document.getElementById('page-outlet')
const disconnectEvent = new CustomEvent('disconnect')

let currentPage

async function render() {
    if (currentPage instanceof Node) {
        pageOutlet.innerHTML = ''
        currentPage.dispatchEvent(disconnectEvent)
    }
    currentPage = await router.exec(location.pathname)
    pageOutlet.appendChild(currentPage)
}
render()

addEventListener('click', dispatchPopStateOnClick)
addEventListener('popstate', render)
