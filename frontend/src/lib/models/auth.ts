import { z } from 'zod/v4';

export const LoginWithOTP = z.object({
	email: z.email({ error: 'must be a valid email' }).max(255).toLowerCase().trim()
});

export const VerifyOTP = z.object({
	email: z.email({ error: 'must be a valid email' }).max(255).toLowerCase().trim(),
	otp: z.string().min(6).max(6)
});

export interface AuthResponse {
	token: string;
	expiresIn: number;
}
