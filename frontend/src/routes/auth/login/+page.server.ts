import { fail, redirect } from '@sveltejs/kit';
import { signInWithOneTimePassword, verifyOneTimePassword } from '$lib/server/gateway/auth';
import { LoginWithOTP, VerifyOTP } from '$lib/models/auth';
import type { Actions, PageServerLoad } from './$types';
import { env } from '$env/dynamic/private';
import { logger } from '$lib/logger';

export const load: PageServerLoad = async () => {
	return {};
};

export const actions: Actions = {
	sendOTP: async ({ request }) => {
		const formData = await request.formData();
		const formEntries = Object.fromEntries(formData.entries());

		logger.info('sending otp...');

		const validation = LoginWithOTP.safeParse(formEntries);
		if (!validation.success) {
			return fail(400, {
				success: false,
				error: 'Please enter a valid email address.'
			});
		}

		const resp = await signInWithOneTimePassword(validation.data.email);
		if (!resp.success) {
			logger.error('Failed to send OTP: %s', resp.error);
			return fail(400, {
				success: false,
				error: 'Could not send code. Try again'
			});
		}

		return {
			success: true,
			email: validation.data.email,
			step: 2
		};
	},
	validateOTP: async ({ request, cookies }) => {
		const formData = await request.formData();
		const formEntries = Object.fromEntries(formData.entries());

		//console.log('Form entries received for OTP validation:', formEntries);

		logger.info('validating otp...');

		const validation = VerifyOTP.safeParse(formEntries);
		if (!validation.success) {
			//console.log('Invalid OTP validation:', validation.error);
			return fail(400, {
				success: false,
				error: 'Invalid Credentials.'
			});
		}

		const ipAddress =
			request.headers.get('cf-connecting-ip') ||
			request.headers.get('x-real-ip') ||
			request.headers.get('x-forwarded-for')?.split(',')[0].trim() ||
			'127.0.0.1';

		const userAgent = request.headers.get('user-agent') || 'unknown';

		const resp = await verifyOneTimePassword(
			validation.data.email,
			validation.data.otp,
			ipAddress,
			userAgent
		);

		if (!resp.success || !resp.data?.token) {
			return fail(400, {
				success: false,
				error: 'Invalid Credentials.'
			});
		}

		// We are good to go
		const isProduction = env.NODE_ENV === 'production';

		cookies.set('session', resp.data!.token, {
			path: '/',
			maxAge: 60 * 60 * 24, // 1 day
			httpOnly: true,
			sameSite: 'lax',
			secure: isProduction
		});

		redirect(303, '/home');
	}
};
