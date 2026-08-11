import { Link2 } from 'lucide-react'

export function Header() {
  return (
    <header className="fixed top-0 z-50 w-full bg-white/80 backdrop-blur-xl shadow-[0_1px_8px_rgba(0,0,0,0.04)]">
      <div className="mx-auto flex h-20 max-w-6xl items-center justify-between px-4 sm:px-6 lg:px-12">
        <div className="flex items-center gap-2">
          <span className="text-primary">
            <Link2 className="size-7 sm:size-8" />
          </span>
          <span className="text-xl font-semibold tracking-tight sm:text-2xl">ShortenIt</span>
        </div>
      </div>
    </header>
  )
}
