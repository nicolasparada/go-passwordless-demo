export default class Router {
    constructor() {
        this.routes = /** @type {Route[]} */ ([])
    }

    /**
     * Adds a route handler.
     *
     * @param {string|RegExp} pattern
     * @param {Handler} handler
     */
    handle(pattern, handler) {
        this.routes.push({ pattern, handler })
    }

    /**
     * Executes the handler for the given pathname.
     *
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
     * Register a callback for every time an anchor link is clicked or a "popstate" event accurs.
     * It's also called initialy too.
     *
     * @param {function} callback
     */
    install(callback) {
        const execCallback = () => {
            callback(this.exec(decodeURI(location.pathname)))
        }

        document.body.addEventListener('click', ev => {
            if (ev.defaultPrevented
                || ev.button !== 0
                || ev.ctrlKey
                || ev.shiftKey
                || ev.altKey
                || ev.metaKey) {
                return
            }

            const a = /** @type {Element} */ (ev.target).closest('a')

            if (a === null
                || (a.target !== '' && a.target !== '_self')
                || a.hostname !== location.hostname) {
                return
            }

            ev.preventDefault()

            if (a.href !== location.href) {
                history.pushState(history.state, document.title, a.href)
                execCallback()
            }
        })

        addEventListener('popstate', execCallback)

        execCallback()
    }
}

/**
 * @typedef Route
 * @property {string|RegExp} pattern
 * @property {Handler} handler
 */

/** @typedef {function(...string): any} Handler */
