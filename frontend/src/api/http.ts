const API_URL: string = import.meta.env.VITE_API_URL ?? ''

export class ApiError extends Error {
    status: number

    constructor(status: number, message: string) {
        super(message)
        this.status = status
    }
}

type Query = Record<string, string | number | boolean | undefined | null>

interface RequestOptions {
    method?: string
    body?: unknown
    query?: Query
    token?: string | null
}

let unauthorizedHandler: (() => void) | null = null

export function onUnauthorized(handler: () => void): void {
    unauthorizedHandler = handler
}

function buildUrl(path: string, query?: Query): string {
    const params = new URLSearchParams()
    for (const [key, value] of Object.entries(query ?? {})) {
        if (value !== undefined && value !== null && value !== '') {
            params.set(key, String(value))
        }
    }
    const qs = params.toString()
    return `${API_URL}/api${path}${qs ? `?${qs}` : ''}`
}

export async function request<T>(
    path: string,
    options: RequestOptions = {}
): Promise<T> {
    const headers: Record<string, string> = {}
    let body: BodyInit | undefined

    if (options.body instanceof FormData) {
        body = options.body
    } else if (options.body !== undefined) {
        headers['Content-Type'] = 'application/json'
        body = JSON.stringify(options.body)
    }
    if (options.token) {
        headers.Authorization = `Bearer ${options.token}`
    }

    const response = await fetch(buildUrl(path, options.query), {
        method: options.method ?? 'GET',
        headers,
        body,
    })

    if (response.status === 401 && options.token) {
        unauthorizedHandler?.()
    }

    if (!response.ok) {
        let message = response.statusText
        try {
            const data = await response.json()
            message = data.error ?? message
        } catch {
            // non-JSON error body
        }
        throw new ApiError(response.status, message)
    }

    if (response.status === 204) {
        return undefined as T
    }
    return (await response.json()) as T
}

/** Resolves uploaded image paths against the API origin when it differs. */
export function assetUrl(url: string): string {
    if (!url || /^(https?:)?\/\//.test(url) || !url.startsWith('/uploads/')) {
        return url
    }
    return `${API_URL}${url}`
}
