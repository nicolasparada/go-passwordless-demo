const template = document.createElement('template')
template.innerHTML = `
<div class="container">
    <h1>Not Found Page</h1>
</div>
`

export default function HomePage() {
    return template.content.cloneNode(true)
}
