import type { ClaimsService } from '$lib/models/claims';

// See https://svelte.dev/docs/kit/types#app.d.ts
// for information about these interfaces
declare global {
	namespace App {
		// interface Error {}
		interface Locals {
			claims: ClaimsService;
		}
		// interface PageData {}
		// interface PageState {}
		// interface Platform {}
	}
}

export {};
