// VITE_API_URL is set in Vercel's environment variables (e.g. https://your-api.onrender.com).
// In local dev it is empty and Vite's proxy forwards /api → localhost:8080.
const API_BASE = import.meta.env.VITE_API_URL ?? ''

interface ShortenResponse {
  success: boolean
  short_url?: string
  error?: string
}

export async function shortenUrl(longUrl: string): Promise<string> {
  const res = await fetch(`${API_BASE}/api/url_shorter`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ long_url: longUrl }),
  })

  const data: ShortenResponse = await res.json().catch(() => ({ success: false }))

  if (!res.ok || !data.success) {
    throw new Error(data.error ?? 'Something went wrong. Please try again.')
  }

  if (!data.short_url) {
    throw new Error('Unexpected response from the server.')
  }

  return data.short_url
}

export function shortUrlHref(shortUrl: string): string {
  if (/^https?:\/\//i.test(shortUrl)) return shortUrl
  return `http://${shortUrl}`
}
