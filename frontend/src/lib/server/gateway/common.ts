import { env } from '$env/dynamic/private';
import { logger } from '$lib/logger';

export class APIError extends Error {
	code: string;

	constructor(code: string, message: string) {
		super(message);
		this.code = code;
	}

	static fromFetchError(status: number, data: unknown): APIError {
		let code: string;
		let message: string;

		// Try to extract code and message from response data
		if (data && typeof data === 'object' && 'code' in data && 'message' in data) {
			code = String(data.code);
			message = String(data.message);
		} else {
			code = 'UNKNOWN_ERROR';
			message = 'An unknown error occurred';
		}

		// Override specific status codes
		if (status === 401) {
			code = 'UNAUTHORIZED';
		} else if (status === 403) {
			code = 'FORBIDDEN';
		} else if (status === 404) {
			code = 'NOT_FOUND';
		} else if (status === 500) {
			code = 'INTERNAL_SERVER_ERROR';
		}

		return new APIError(code, message);
	}
}

export interface Result<T, E = Error> {
	success: boolean;
	data?: T;
	error?: E;
}

export interface GenericResponse {
	success: boolean;
}

export interface GenericStringErrorResponse {
	message: string;
}

export const API_URL = env.BACKEND_API || 'http://localhost:3000';
export const MEDIA_URL = env.MEDIA_URL || 'https://media.myapp.com';
export const CLOUDFLARE_ACCOUNT_ID = env.CLOUDFLARE_CLIENT_ID || '';
export const CLOUDFLARE_TOKEN = env.CLOUDFLARE_API_TOKEN || '';

const REQUEST_TIMEOUT_MS = 10000;

/**
 * Create a fetch request with timeout and standard headers
 */
export async function fetchWithTimeout(
	endpoint: string,
	options: RequestInit = {}
): Promise<Response> {
	const controller = new AbortController();
	const timeoutId = setTimeout(() => controller.abort(), REQUEST_TIMEOUT_MS);

	// Build headers
	const headers = new Headers(options.headers);

	// Only set Content-Type for requests with a body (and not FormData)
	if (!headers.has('Content-Type') && options.body && !(options.body instanceof FormData)) {
		headers.set('Content-Type', 'application/json');
	}

	try {
		// Construct full URL safely
		const url = new URL(endpoint, API_URL).toString();
		//console.log(`[Fetch] Requesting URL: ${url}`);
		const response = await fetch(url, {
			...options,
			headers,
			signal: controller.signal
		});

		//console.log(`[Fetch] ${response.status} ${url}`);

		return response;
	} finally {
		clearTimeout(timeoutId);
	}
}

/**
 * Handle fetch response and convert to typed data
 */
export async function handleFetchResponse<T>(response: Response): Promise<T> {
	const text = await response.text();
	let data: unknown;

	try {
		data = text ? JSON.parse(text) : {};
	} catch {
		data = text;
	}

	if (!response.ok) {
		const apiError = APIError.fromFetchError(response.status, data);
		throw apiError;
	}

	return data as T;
}

/**
 * Log and handle fetch errors
 */
export const logAndHandleFetchError = (action: string, error: unknown): APIError => {
	if (error instanceof APIError) {
		const errorDataString = JSON.stringify(
			{
				code: error.code,
				message: error.message
			},
			null,
			2
		);
		logger.error(`${action} call failed (${error.code}). Message: ${errorDataString}`);
		return error;
	} else if (error instanceof TypeError && error.message.includes('AbortError')) {
		logger.error(`${action} call failed: Request timeout`);
		return new APIError('TIMEOUT', 'Request timeout');
	} else if (error instanceof TypeError) {
		logger.error(`${action} call failed with network error: ${error.message}`);
		return new APIError('NETWORK_ERROR', 'Network error occurred');
	} else {
		const errorString = error instanceof Error ? error.message : String(error);
		logger.error(`${action} call failed with unknown error: ${errorString}`);
		return new APIError('UNKNOWN_ERROR', 'An unknown error occurred');
	}
};
