function createRouter(routes) {
    return function (pathname) {
        for (const [pattern, handler] of routes) {
            if (typeof pattern === 'string') {
                if (pattern !== pathname) continue
                return handler()
            }
            const match = pattern.exec(pathname)
            if (match === null) continue
            return handler(...match.slice(1))
        }
    }
}

const pagesCache = new Map()
async function loadPage(name) {
    if (pagesCache.has(name))
        return pagesCache.get(name)
    const page = await import(`/js/pages/${name}-page.js`).then(m => m.default)
    pagesCache.set(name, page)
    return page
}

function view(name) {
    return (...args) => loadPage(name).then(page => page(...args))
}

const route = createRouter([
    ['/', view('home')],
    ['/callback', view('callback')],
    [/^\//, view('not-found')],
])

const pageOutlet = document.getElementById('page-outlet')
let currentPage
async function render() {
    if (currentPage instanceof Node) {
        pageOutlet.innerHTML = ''
    }
    currentPage = await route(decodeURI(location.pathname))
    pageOutlet.appendChild(currentPage)
}
render()

addEventListener('popstate', render)
addEventListener('click', hijackClicks)

/**
 * @param {MouseEvent} ev
 */
function hijackClicks(ev) {
    if (ev.defaultPrevented
        || ev.altKey
        || ev.ctrlKey
        || ev.metaKey
        || ev.shiftKey
        || ev.button !== 0) return

    const a = Array
        .from(walkParents(ev.target))
        .find(n => n instanceof HTMLAnchorElement)

    if (!(a instanceof HTMLAnchorElement)
        || (a.target !== '' && a.target !== '_self')
        || a.hostname !== location.hostname)
        return

    ev.stopImmediatePropagation()
    ev.stopPropagation()
    ev.preventDefault()

    const { state } = history
    history.pushState(state, document.title, a.href)
    dispatchEvent(new PopStateEvent('popstate', { state }))
}

function* walkParents(node) {
    do {
        yield node
    } while ((node = node.parentNode) instanceof Node)
}
