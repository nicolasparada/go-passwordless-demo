import { createRouter } from 'https://unpkg.com/@nicolasparada/router@0.8.0/router.js';
import { guard } from './auth.js';
import { importWithCache } from './dynamic-import.js';

const r = createRouter()

r.route('/', guard(view('home'), view('access')))
r.route('/callback', view('callback'))
r.route(/^\//, view('not-found'))
r.subscribe(async result => {
    document.body.innerHTML = ''
    result = await result
    if (typeof result === 'string') {
        document.body.innerHTML = result
    } else if (result instanceof Node) {
        document.body.appendChild(result)
    } else {
        throw new Error('cannot render page')
    }
})
r.install()

function view(name) {
    return (...args) => importWithCache(`/js/pages/${name}-page.js`)
        .then(m => m.default(...args))
}
