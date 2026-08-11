import { Link2 } from 'lucide-react'

export function Footer() {
  return (
    <footer className="w-full border-t border-border/60 bg-card">
      <div className="mx-auto flex max-w-6xl flex-col items-center justify-between gap-4 px-4 py-8 text-center sm:flex-row sm:px-6 sm:text-left lg:px-12">
        <div className="flex items-center gap-2">
          <span className="text-primary">
            <Link2 className="size-5" />
          </span>
          <span className="text-xl font-semibold tracking-tight">ShortenIt</span>
        </div>
        <p className="text-sm text-muted-foreground">
          © {new Date().getFullYear()} ShortenIt. All rights reserved.
        </p>
      </div>
    </footer>
  )
}
