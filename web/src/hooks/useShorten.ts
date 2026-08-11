import { useMutation } from '@tanstack/react-query'
import { shortenUrl } from '@/lib/api'

export function useShorten() {
  return useMutation({
    mutationFn: shortenUrl,
  })
}
