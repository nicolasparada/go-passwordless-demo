export default class Router {
    constructor() {
        this.routes = /** @type {Route[]} */ ([])

        this.handle = this.handle.bind(this)
        this.exec = this.exec.bind(this)
    }

    /**
     * @param {string|RegExp} pattern
     * @param {Handler} handler
     */
    handle(pattern, handler) {
        this.routes.push({ pattern, handler })
    }

    /**
     * @param {string} pathname
     */
    exec(pathname) {
        for (const route of this.routes) {
            if (typeof route.pattern === 'string') {
                if (route.pattern !== pathname) continue
                // @ts-ignore
                return route.handler()
            }
            const match = route.pattern.exec(pathname)
            if (match === null) continue
            return route.handler(...match.slice(1))
        }
    }

    /**
     * @param {MouseEvent} ev
     */
    static delegateClicks(ev) {
        if (ev.defaultPrevented
            || ev.altKey
            || ev.ctrlKey
            || ev.metaKey
            || ev.shiftKey
            || ev.button !== 0)
            return

        const a = /** @type {Element} */ (ev.target).closest('a')

        if (a === null
            || (a.target !== '' && a.target !== '_self')
            || a.hostname !== location.hostname)
            return

        ev.preventDefault()
        Router.updateHistory(a.href)
    }

    /**
     * @param {string} href
     * @param {boolean=} redirect
     */
    static updateHistory(href, redirect = false) {
        const { state } = history
        history[redirect ? 'replaceState' : 'pushState'](state, document.title, href)
        dispatchEvent(new PopStateEvent('popstate', { state }))
    }
}

function* walkParents(node) {
    do {
        yield node
    } while ((node = node.parentNode) instanceof Node)
}

/**
 * @typedef Route
 * @property {string|RegExp} pattern
 * @property {Handler} handler
 */

/** @typedef {function(...string): any} Handler */
