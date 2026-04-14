import { greet } from 'shared'

function App() {
  return (
    <div className="app">
      <header>
        <h1>Feishu Pipeline</h1>
      </header>
      <main>
        <p>{greet('Monorepo')}</p>
      </main>
    </div>
  )
}

export default App
