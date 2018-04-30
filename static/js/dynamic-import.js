Object.defineProperty(window, Symbol.for('registry'), { value: new Map() })
export default function dynamicImport(src) {
    const registry = window[Symbol.for('registry')]
    if (registry.has(src))
        return registry.get(src).promise

    const script = document.createElement('script')
    script.type = 'module'
    script.textContent = `import * as x from '${src}'\nwindow[Symbol.for('registry')].get('${src}').resolve(x)`
    script.onload = () => {
        script.remove()
    }
    script.onerror = err => {
        registry.get(src).reject(err)
        script.remove()
    }

    const record = { src, script }
    record.promise = new Promise((resolve, reject) => {
        record.resolve = resolve
        record.reject = reject
    })

    document.head.appendChild(script)
    registry.set(src, record)

    return record.promise
}

const modulesCache = new Map()
export async function importWithCache(specifier) {
    if (modulesCache.has(specifier))
        return modulesCache.get(specifier)
    const m = await dynamicImport(specifier)
    modulesCache.set(specifier, m)
    return m
}
