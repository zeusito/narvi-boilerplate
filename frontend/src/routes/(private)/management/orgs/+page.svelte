<script lang="ts">
	import { enhance } from '$app/forms';
	import { toast } from 'svelte-sonner';
	import * as Card from '$lib/components/ui/card/index.js';
	import * as Table from '$lib/components/ui/table/index.js';
	import * as Dialog from '$lib/components/ui/dialog/index.js';
	import * as Select from '$lib/components/ui/select/index.js';
	import { Badge } from '$lib/components/ui/badge/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import { Input } from '$lib/components/ui/input/index.js';
	import { Field, FieldGroup, FieldLabel } from '$lib/components/ui/field/index.js';
	import { Spinner } from '$lib/components/ui/spinner/index.js';
	import type { PageProps } from './$types';
	import type { Organization } from '$lib/models/organization';

	let props: PageProps = $props();

	let dialogOpen = $state(false);
	let mode: 'create' | 'edit' = $state('create');
	let working = $state(false);
	let editing: Organization | null = $state(null);

	let kindValue = $state('standard');
	let stateValue = $state('active');
	let nameValue = $state('');
	let logoValue = $state('');

	const organizations = $derived(props.data.organizations ?? []);

	function openCreate() {
		mode = 'create';
		editing = null;
		nameValue = '';
		logoValue = '';
		kindValue = 'standard';
		stateValue = 'active';
		dialogOpen = true;
	}

	function openEdit(org: Organization) {
		mode = 'edit';
		editing = org;
		nameValue = org.name;
		logoValue = org.logo ?? '';
		stateValue = org.state === 'deleted' ? 'active' : org.state;
		dialogOpen = true;
	}

	$effect(() => {
		if (props.data.loadError) {
			toast.error(props.data.loadError);
		}
	});
</script>

<div class="flex flex-col gap-4">
	<Card.Root>
		<Card.Header class="flex flex-row items-center justify-between">
			<div>
				<Card.Title>Organizations</Card.Title>
				<Card.Description>
					{props.data.count} organization{props.data.count === 1 ? '' : 's'} registered.
				</Card.Description>
			</div>
			<Button onclick={openCreate}>New organization</Button>
		</Card.Header>
		<Card.Content>
			{#if organizations.length === 0}
				<p class="py-8 text-center text-sm text-muted-foreground">
					No organizations yet. Create the first one.
				</p>
			{:else}
				<Table.Root>
					<Table.Header>
						<Table.Row>
							<Table.Head>Name</Table.Head>
							<Table.Head>Slug</Table.Head>
							<Table.Head>Kind</Table.Head>
							<Table.Head>State</Table.Head>
							<Table.Head>Created</Table.Head>
							<Table.Head class="text-right">Actions</Table.Head>
						</Table.Row>
					</Table.Header>
					<Table.Body>
						{#each organizations as org (org.id)}
							<Table.Row>
								<Table.Cell class="font-medium">{org.name}</Table.Cell>
								<Table.Cell class="font-mono text-xs">{org.slug}</Table.Cell>
								<Table.Cell>
									<Badge variant={org.kind === 'management' ? 'default' : 'secondary'}>
										{org.kind}
									</Badge>
								</Table.Cell>
								<Table.Cell>
									<Badge
										variant={org.state === 'active'
											? 'secondary'
											: org.state === 'suspended'
												? 'destructive'
												: 'outline'}
									>
										{org.state}
									</Badge>
								</Table.Cell>
								<Table.Cell class="text-xs text-muted-foreground">
									{new Date(org.createdAt).toLocaleDateString()}
								</Table.Cell>
								<Table.Cell class="text-right">
									{#if org.state !== 'deleted'}
										<Button variant="outline" size="sm" onclick={() => openEdit(org)}>Edit</Button>
									{/if}
								</Table.Cell>
							</Table.Row>
						{/each}
					</Table.Body>
				</Table.Root>
			{/if}
		</Card.Content>
	</Card.Root>
</div>

<Dialog.Root bind:open={dialogOpen}>
	<Dialog.Content>
		<Dialog.Header>
			<Dialog.Title>{mode === 'create' ? 'New organization' : 'Edit organization'}</Dialog.Title>
			<Dialog.Description>
				{mode === 'create'
					? 'Create a new tenant organization. Slug is generated automatically.'
					: `Editing ${editing?.name ?? ''}. Slug cannot be changed.`}
			</Dialog.Description>
		</Dialog.Header>
		<form
			method="POST"
			action={mode === 'create' ? '?/create' : '?/update'}
			use:enhance={() => {
				working = true;
				const currentMode = mode;
				return async ({ result, update }) => {
					await update({ reset: false });
					working = false;
					if (result.type === 'success') {
						toast.success(
							currentMode === 'create' ? 'Organization created.' : 'Organization updated.'
						);
						dialogOpen = false;
					} else if (result.type === 'failure') {
						const err = result.data?.error;
						if (err) {
							toast.error(String(err));
						}
					} else if (result.type === 'error') {
						toast.error('An unexpected error occurred.');
					}
				};
			}}
		>
			<FieldGroup>
				{#if mode === 'edit' && editing}
					<input type="hidden" name="id" value={editing.id} />
				{/if}
				<Field>
					<FieldLabel for="org-name">Name</FieldLabel>
					<Input
						id="org-name"
						name="name"
						required
						minlength={3}
						maxlength={255}
						placeholder="Acme Inc"
						bind:value={nameValue}
					/>
				</Field>
				{#if mode === 'create'}
					<Field>
						<FieldLabel>Kind</FieldLabel>
						<Select.Root type="single" bind:value={kindValue}>
							<Select.Trigger class="w-full">{kindValue}</Select.Trigger>
							<Select.Content>
								<Select.Item value="standard">standard</Select.Item>
								<Select.Item value="management">management</Select.Item>
							</Select.Content>
						</Select.Root>
						<input type="hidden" name="kind" value={kindValue} />
					</Field>
				{:else}
					<Field>
						<FieldLabel for="org-logo">Logo URL</FieldLabel>
						<Input
							id="org-logo"
							name="logo"
							type="url"
							maxlength={500}
							placeholder="https://…"
							bind:value={logoValue}
						/>
					</Field>
					<Field>
						<FieldLabel>State</FieldLabel>
						<Select.Root type="single" bind:value={stateValue}>
							<Select.Trigger class="w-full">{stateValue}</Select.Trigger>
							<Select.Content>
								<Select.Item value="active">active</Select.Item>
								<Select.Item value="suspended">suspended</Select.Item>
							</Select.Content>
						</Select.Root>
						<input type="hidden" name="state" value={stateValue} />
					</Field>
				{/if}
			</FieldGroup>
			<Dialog.Footer class="mt-6">
				<Dialog.Close>
					{#snippet child({ props: closeProps })}
						<Button variant="outline" {...closeProps}>Cancel</Button>
					{/snippet}
				</Dialog.Close>
				<Button type="submit" disabled={working}>
					{#if working}
						<Spinner />
					{/if}
					{mode === 'create' ? 'Create' : 'Save changes'}
				</Button>
			</Dialog.Footer>
		</form>
	</Dialog.Content>
</Dialog.Root>
