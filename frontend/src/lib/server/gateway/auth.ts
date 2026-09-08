import {
	APIError,
	fetchWithTimeout,
	handleFetchResponse,
	logAndHandleFetchError,
	type GenericResponse,
	type Result
} from './common';
import type { AuthResponse } from '$lib/models/auth';
import type { PrincipalClaims } from '$lib/models/claims';

export const signOut = async (
	principalToken: string
): Promise<Result<GenericResponse, APIError>> => {
	try {
		const response = await fetchWithTimeout('/v1/auth/logout', {
			method: 'DELETE',
			headers: {
				Authorization: `Bearer ${principalToken}`
			}
		});

		const data = await handleFetchResponse<GenericResponse>(response);

		return {
			success: true,
			data
		};
	} catch (error) {
		const apiError = logAndHandleFetchError('AuthGateway.signOut', error);

		return {
			success: false,
			error: apiError
		};
	}
};

export const signInWithOneTimePassword = async (
	email: string
): Promise<Result<GenericResponse, APIError>> => {
	try {
		const response = await fetchWithTimeout('/v1/auth/otp/send', {
			method: 'POST',
			body: JSON.stringify({
				email
			})
		});

		await handleFetchResponse(response);

		return {
			success: true,
			data: { success: true }
		};
	} catch (error) {
		const apiError = logAndHandleFetchError('AuthGateway.signInWithOneTimePassword', error);

		return {
			success: false,
			error: apiError
		};
	}
};

export const verifyOneTimePassword = async (
	email: string,
	code: string,
	ipAddress: string,
	userAgent: string
): Promise<Result<AuthResponse, APIError>> => {
	try {
		const response = await fetchWithTimeout('/v1/auth/otp/verify', {
			method: 'POST',
			body: JSON.stringify({
				email,
				code,
				ipAddress,
				userAgent
			})
		});

		const data = await handleFetchResponse<AuthResponse>(response);

		return {
			success: true,
			data
		};
	} catch (error) {
		const apiError = logAndHandleFetchError('AuthGateway.verifyOneTimePassword', error);

		return {
			success: false,
			error: apiError
		};
	}
};

export const introspectToken = async (
	principalToken: string
): Promise<Result<PrincipalClaims, APIError>> => {
	try {
		const response = await fetchWithTimeout('/v1/auth/introspect', {
			method: 'GET',
			headers: {
				Authorization: `Bearer ${principalToken}`
			}
		});

		const data = await handleFetchResponse<PrincipalClaims>(response);

		return {
			success: true,
			data
		};
	} catch (error) {
		const apiError = logAndHandleFetchError('AuthGateway.introspectToken', error);

		return {
			success: false,
			error: apiError
		};
	}
};
