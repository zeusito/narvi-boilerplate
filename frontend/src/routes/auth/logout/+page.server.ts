import type { PageServerLoad } from './$types';
import { signOut } from '$lib/server/gateway/auth';

export const load: PageServerLoad = async ({ cookies }) => {
	const session = cookies.get('session');
	if (session) {
		await signOut(session);
		cookies.delete('session', { path: '/' });
	}

	return {};
};
