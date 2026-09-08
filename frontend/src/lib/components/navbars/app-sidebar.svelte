<script lang="ts">
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import { UsersRoundIcon, Library, Building, CogIcon, House } from '@lucide/svelte';
	import AppUser from './app-user.svelte';
	import { page } from '$app/state';

	type Props = {
		userName: string;
		orgSlug: string;
		orgName: string;
	};

	let props: Props = $props();

	const buildOrgUrl = (base: string) => `/org/${props.orgSlug}${base}`;

	const items = {
		orgActions: [{ title: 'Team', url: buildOrgUrl('/team'), icon: UsersRoundIcon }]
	};
</script>

<Sidebar.Root collapsible="offcanvas" variant="inset">
	<Sidebar.Header>
		Org: {props.orgName}
	</Sidebar.Header>
	<Sidebar.Content>
		<Sidebar.Group>
			<Sidebar.GroupLabel>Your Actions</Sidebar.GroupLabel>
			<Sidebar.GroupContent>
				<Sidebar.Menu>
					<Sidebar.MenuItem>
						<Sidebar.MenuButton isActive={page.url.pathname.startsWith('/home')}>
							{#snippet child({ props })}
								<a href="/home" {...props}>
									<House />
									<span>Home</span>
								</a>
							{/snippet}
						</Sidebar.MenuButton>
					</Sidebar.MenuItem>
				</Sidebar.Menu>
			</Sidebar.GroupContent>
		</Sidebar.Group>
		<Sidebar.Group>
			<Sidebar.GroupLabel>Org Actions</Sidebar.GroupLabel>
			<Sidebar.GroupContent>
				{#if props.orgSlug.length > 0}
					<Sidebar.Menu>
						{#each items.orgActions as item (item.title)}
							<Sidebar.MenuItem>
								<Sidebar.MenuButton isActive={item.url.includes(page.url.pathname)}>
									{#snippet child({ props })}
										<a href={item.url} {...props}>
											<item.icon />
											<span>{item.title}</span>
										</a>
									{/snippet}
								</Sidebar.MenuButton>
							</Sidebar.MenuItem>
						{/each}
					</Sidebar.Menu>
				{/if}
			</Sidebar.GroupContent>
		</Sidebar.Group>
	</Sidebar.Content>
	<Sidebar.Footer>
		<AppUser userName={props.userName} orgName={props.orgName} />
	</Sidebar.Footer>
</Sidebar.Root>
