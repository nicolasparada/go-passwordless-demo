const template = document.createElement('template')
template.innerHTML = `
<h1>Not Found Page</h1>
`

export default function HomePage() {
    return template.content.cloneNode(true)
}
