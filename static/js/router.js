export class Router {
    constructor() {
        this.routes = []

        this.handle = this.handle.bind(this)
        this.exec = this.exec.bind(this)
    }

    /**
     * @param {string|RegExp} pattern
     * @param {handler} handler
     */
    handle(pattern, handler) {
        this.routes.push({ pattern, handler })
    }

    /**
     * @param {string} pathname
     * @returns {Node|Promise<Node>}
     */
    exec(pathname) {
        pathname = decodeURI(pathname)
        for (const route of this.routes) {
            if (typeof route.pattern === 'string') {
                if (route.pattern !== pathname) continue
                return route.handler()
            }
            const match = route.pattern.exec(pathname)
            if (match === null) continue
            return route.handler(...match.slice(1))
        }
    }
}

/**
 * @param {MouseEvent} ev
 */
export function dispatchPopStateOnClick(ev) {
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

/** @typedef {function(...string): Node|Promise<Node>} handler */
