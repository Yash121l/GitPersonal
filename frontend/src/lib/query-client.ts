import { QueryClient, VueQueryPlugin, type VueQueryPluginOptions } from '@tanstack/vue-query'

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 15_000,
      gcTime: 5 * 60_000,
      refetchOnWindowFocus: false,
      retry(failureCount, error) {
        if (error instanceof Error && 'status' in error && (error as { status?: number }).status === 401) {
          return false
        }
        return failureCount < 2
      },
    },
  },
})

export { VueQueryPlugin }
export type { VueQueryPluginOptions }
