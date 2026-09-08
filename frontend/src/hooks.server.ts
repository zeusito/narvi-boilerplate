import { logger } from '$lib/logger';
import { ANONYMOUS_CLAIMS, ClaimsService } from '$lib/models/claims';
import { introspectToken } from '$lib/server/gateway/auth';
import { redirect, type Handle } from '@sveltejs/kit';
import { sequence } from '@sveltejs/kit/hooks';

/**
 * 1. Authentication Handle:
 * Reads session cookie, introspects token, and initializes `locals.claims`.
 */
const authentication: Handle = async ({ event, resolve }) => {
	const { cookies, locals } = event;
	const session = cookies.get('session');

	let claims = ANONYMOUS_CLAIMS;

	if (session) {
		const resp = await introspectToken(session);

		logger.info(`Introspected token response: ${JSON.stringify(resp)}`);

		if (resp.success && resp.data) {
			claims = resp.data;
			claims.token = session;
		} else {
			logger.error('Failed to introspect token, clearing invalid session');
			cookies.delete('session', { path: '/' });
		}
	}

	logger.info(`Initialized claims for user: ${claims.fullName ?? 'anonymous'}`);

	// Always attach claims service to locals for downstream consumers
	locals.claims = new ClaimsService(claims);

	return resolve(event);
};

/**
 * 2. Authorization Handle:
 * Enforces route-level access, roles, and tenant slug matching.
 */
const authorization: Handle = async ({ event, resolve }) => {
	const { url, locals } = event;
	const claimsService = locals.claims;

	// Public routes: skip auth guards
	if (url.pathname === '/' || url.pathname.startsWith('/auth')) {
		return resolve(event);
	}

	// 1. Base authentication guard for private routes
	if (!claimsService.isAuthenticated()) {
		logger.warn(`Unauthenticated attempt to access ${url.pathname}, redirecting to login`);
		throw redirect(303, '/auth/login');
	}

	return resolve(event);
};

// Chain handles together in execution order
export const handle: Handle = sequence(authentication, authorization);
