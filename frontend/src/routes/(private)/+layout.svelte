<script lang="ts">
	import type { Snippet } from 'svelte';
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import AppHeader from '$lib/components/navbars/app-topbar.svelte';
	import AppSidebar from '$lib/components/navbars/app-sidebar.svelte';
	import type { LayoutData } from './$types';

	let { data, children }: { data: LayoutData; children: Snippet } = $props();
</script>

<Sidebar.Provider>
	<AppSidebar
		userName={data.claims.fullName}
		orgName={data.claims.organizationName ?? ''}
		orgSlug={data.claims.organizationSlug ?? ''}
	/>
	<Sidebar.Inset>
		<AppHeader orgName={data.claims.organizationName} />
		<main class="min-h-full bg-muted">
			<div class="mx-auto flex w-full max-w-7xl flex-col p-4">
				{@render children?.()}
			</div>
		</main>
	</Sidebar.Inset>
</Sidebar.Provider>
