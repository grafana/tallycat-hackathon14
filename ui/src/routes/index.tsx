import { createFileRoute } from '@tanstack/react-router'
import { ModeToggle } from '@/components/mode-toggle'

export const Route = createFileRoute('/')({
  component: App,
})

function App() {
  return (
    <div className="text-center">
      <ModeToggle />
    </div>
  )
}
