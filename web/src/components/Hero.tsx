import { useState, type FormEvent } from 'react'
import { ArrowRight, Check, Copy, Link2, Loader2 } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { useShorten } from '@/hooks/useShorten'
import { shortUrlHref } from '@/lib/api'

function isValidHttpUrl(value: string): boolean {
  try {
    const parsed = new URL(value)
    return parsed.protocol === 'http:' || parsed.protocol === 'https:'
  } catch {
    return false
  }
}

export function Hero() {
  const [url, setUrl] = useState('')
  const [copied, setCopied] = useState(false)
  const { mutate, isPending, error, data } = useShorten()

  const handleSubmit = (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    const value = url.trim()
    if (!value || isPending) return

    if (!isValidHttpUrl(value)) {
      toast.error('Please enter a valid http(s) URL.')
      return
    }

    mutate(value)
  }

  const handleCopy = async () => {
    if (!data) return
    try {
      await navigator.clipboard.writeText(shortUrlHref(data))
      setCopied(true)
      toast.success('Copied to clipboard')
      setTimeout(() => setCopied(false), 2000)
    } catch {
      toast.error('Failed to copy. Copy it manually.')
    }
  }

  return (
    <section className="relative w-full overflow-hidden bg-background">
      <div className="pointer-events-none absolute top-0 left-1/2 h-[800px] w-[800px] -translate-x-1/2 rounded-full bg-primary/5 blur-[100px]" />
      <div className="pointer-events-none absolute top-1/4 -right-32 h-[400px] w-[400px] rounded-full bg-orange-400/10 blur-[80px]" />

      <div className="relative z-10 mx-auto flex w-full max-w-[800px] flex-col items-center px-4 pt-28 pb-20 text-center sm:px-6 lg:px-12 lg:pt-36 lg:pb-32">
        <h1 className="text-[30px] leading-tight font-extrabold tracking-tight text-foreground sm:text-4xl lg:text-5xl lg:leading-[1.15]">
          Shorten your links, <span className="text-primary">broaden your reach.</span>
        </h1>
        <p className="mt-6 max-w-[600px] text-lg leading-7 text-muted-foreground">
          Simple, powerful URL shortening to help you share clean, trackable links.
        </p>

        <form
          className="mt-10 flex w-full max-w-[640px] flex-col gap-3 rounded-2xl bg-card p-2 shadow-lg transition-shadow focus-within:shadow-xl hover:shadow-xl sm:flex-row"
          onSubmit={handleSubmit}
        >
          <div className="flex h-14 items-center rounded-xl bg-muted/60 px-4 sm:h-16 sm:flex-1">
            <Link2 className="mr-3 size-5 shrink-0 text-muted-foreground" />
            <Input
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              placeholder="Paste a long URL here..."
              type="url"
              inputMode="url"
              autoComplete="url"
              aria-label="Long URL to shorten"
              className="h-full flex-1 border-0 bg-transparent p-0 text-base shadow-none focus-visible:ring-0"
            />
          </div>
          <Button
            type="submit"
            disabled={isPending || !url.trim()}
            className="h-14 px-8 text-sm font-semibold sm:h-16"
          >
            {isPending ? (
              <>
                <Loader2 className="animate-spin" />
                Shortening…
              </>
            ) : (
              <>
                Shorten
                <ArrowRight />
              </>
            )}
          </Button>
        </form>

        {error && (
          <p className="mt-4 rounded-lg bg-destructive/10 px-4 py-2 text-sm font-medium text-destructive">
            {error.message}
          </p>
        )}

        {data && (
          <div className="mt-8 flex w-full max-w-[640px] flex-col items-stretch justify-between gap-3 rounded-2xl border border-primary/20 bg-primary/5 p-4 sm:flex-row sm:items-center sm:gap-4">
            <div className="flex min-w-0 flex-col items-start px-2 text-left">
              <span className="text-xs font-semibold text-primary">Link Shortened!</span>
              <a
                href={shortUrlHref(data)}
                target="_blank"
                rel="noreferrer"
                className="truncate font-medium text-primary hover:underline"
              >
                {data}
              </a>
            </div>
            <Button variant="outline" onClick={handleCopy} className="w-full shrink-0 sm:w-auto">
              {copied ? <Check /> : <Copy />}
              {copied ? 'Copied' : 'Copy'}
            </Button>
          </div>
        )}
      </div>
    </section>
  )
}
