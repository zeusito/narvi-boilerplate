import { error, fail } from '@sveltejs/kit';
import {
	createOrganization,
	listOrganizations,
	updateOrganization
} from '$lib/server/gateway/organizations';
import { CreateOrganizationSchema, UpdateOrganizationSchema } from '$lib/models/organization';
import { logger } from '$lib/logger';
import type { Actions, PageServerLoad } from './$types';

export const load: PageServerLoad = async ({ locals }) => {
	const claims = locals.claims;

	if (!claims.isManagementAdmin()) {
		throw error(403, 'Only management organization members can access this page.');
	}

	const resp = await listOrganizations(claims.getClaims().token);

	if (!resp.success || !resp.data) {
		logger.error('Failed to list organizations');
		return {
			organizations: [],
			count: 0,
			loadError: 'Could not load organizations. Please try again.'
		};
	}

	return {
		organizations: resp.data.data ?? [],
		count: resp.data.count ?? 0,
		loadError: null as string | null
	};
};

export const actions: Actions = {
	create: async ({ request, locals }) => {
		const claims = locals.claims;

		if (!claims.isManagementAdmin()) {
			throw error(403, 'Only management organization members can perform this action.');
		}

		const formData = await request.formData();
		const formEntries = Object.fromEntries(formData.entries());

		const validation = CreateOrganizationSchema.safeParse(formEntries);
		if (!validation.success) {
			return fail(400, {
				success: false,
				error: 'Please provide a valid name (min 3 chars) and kind.'
			});
		}

		const resp = await createOrganization(claims.getClaims().token, validation.data);
		if (!resp.success) {
			logger.error('Failed to create organization');
			return fail(400, {
				success: false,
				error: resp.error?.message || 'Could not create organization. Try again.'
			});
		}

		return { success: true };
	},

	update: async ({ request, locals }) => {
		const claims = locals.claims;

		if (!claims.isManagementAdmin()) {
			throw error(403, 'Only management organization members can perform this action.');
		}

		const formData = await request.formData();
		const raw = Object.fromEntries(formData.entries());
		// Normalize empty logo to undefined so Zod optional handling stays clean
		if (typeof raw.logo === 'string' && raw.logo.trim() === '') {
			delete raw.logo;
		}

		const validation = UpdateOrganizationSchema.safeParse(raw);
		if (!validation.success) {
			return fail(400, {
				success: false,
				error: 'Please provide a valid name (min 3 chars) and state.'
			});
		}

		const resp = await updateOrganization(claims.getClaims().token, validation.data);
		if (!resp.success) {
			logger.error('Failed to update organization');
			return fail(400, {
				success: false,
				error: 'Could not update organization. Try again.'
			});
		}

		return { success: true };
	}
};
