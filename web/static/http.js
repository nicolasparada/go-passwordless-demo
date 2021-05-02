/**
 * @param {Response} resp
 */
export function parseResponse(resp) {
    return resp.clone().json().catch(() => resp.text()).then(body => {
        if (!resp.ok) {
            const msg = typeof body === "string" && body !== "" ? body : resp.statusText
            const err = new Error(msg)
            return Promise.reject(err)
        }

        return body
    })
}
