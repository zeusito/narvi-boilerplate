<script lang="ts">
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import { Separator } from '$lib/components/ui/separator/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { BellIcon, LogOutIcon } from '@lucide/svelte';
	import { goto } from '$app/navigation';

	type Props = {
		orgName: string;
	};

	let props: Props = $props();

	const logout = () => {
		goto('/auth/logout');
	};
</script>

<header class="flex h-16 shrink-0 items-center border-b">
	<div class="flex w-full items-center justify-between px-4 lg:px-6">
		<div class="flex items-center gap-1 lg:gap-2">
			<Sidebar.Trigger class="-ms-1" />
			<Separator orientation="vertical" class="mx-2 data-[orientation=vertical]:h-4" />
			<div class="flex flex-col justify-baseline">
				<span class="text-base font-semibold">{props.orgName}</span>
			</div>
		</div>

		<div class="flex items-center gap-2">
			<Button variant="ghost" size="icon" aria-label="Notifications" class="relative">
				<BellIcon class="size-5" />
			</Button>
			<Button variant="ghost" size="icon" aria-label="Log out" onclick={logout}>
				<LogOutIcon class="size-5" />
			</Button>
		</div>
	</div>
</header>
