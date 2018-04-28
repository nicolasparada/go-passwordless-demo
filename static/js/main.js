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

const disconnectEvent = new CustomEvent('disconnect')
const loadingHTML = document.body.innerHTML
let currentPage

async function render() {
    const rendered = currentPage instanceof Node
    if (rendered) {
        document.body.innerHTML = loadingHTML
        currentPage.dispatchEvent(disconnectEvent)
    }
    currentPage = await router.exec(location.pathname)
    document.body.innerHTML = ''
    document.body.appendChild(currentPage)
}
render()

addEventListener('click', dispatchPopStateOnClick)
addEventListener('popstate', render)
