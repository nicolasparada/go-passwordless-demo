import { isAuthenticated } from './auth.js';
import { importWithCache } from './dynamic-import.js';
import Router from './router.js';

const router = new Router()

router.handle('/', guard(view('home'), view('welcome')))
router.handle('/callback', view('callback'))
router.handle(/^\//, view('not-found'))

router.install(async resultPromise => {
    document.body.innerHTML = ''
    document.body.appendChild(await resultPromise)
})


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
