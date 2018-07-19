import Router from 'https://unpkg.com/@nicolasparada/router@0.6.0/router.js';
import { guard } from './auth.js';
import { importWithCache } from './dynamic-import.js';

const router = new Router()

router.handle('/', guard(view('home'), view('access')))
router.handle('/callback', view('callback'))
router.handle(/^\//, view('not-found'))
router.install(async result => {
    document.body.innerHTML = ''
    document.body.appendChild(await result)
})

function view(name) {
    return (...args) => importWithCache(`/js/pages/${name}-page.js`)
        .then(m => m.default)
        .then(h => h(...args))
}
