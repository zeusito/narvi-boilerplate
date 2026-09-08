<script lang="ts">
	import * as Card from '$lib/components/ui/card/index.js';
	import type { PageProps } from './$types';
	import { toast } from 'svelte-sonner';
	import FormEmail from '$lib/components/login/form-email.svelte';
	import FormCode from '$lib/components/login/form-code.svelte';
	import Logo from '$lib/assets/narvi-logo.png';

	let props: PageProps = $props();
	let currentStep: number = $state(1);

	$effect(() => {
		if (props.data?.error) {
			toast.error(props.data.error);
		}
	});

	$effect(() => {
		if (props.form) {
			if (!props.form.success) {
				toast.error(props.form.error || 'Invalid Request');
			} else if (props.form.step) {
				currentStep = props.form.step;
			}
		}
	});
</script>

<div class="flex min-h-full flex-col items-center justify-center gap-6 bg-muted p-6 md:p-10">
	<div class="flex w-full max-w-sm flex-col gap-6">
		<img class="mx-auto h-15 w-auto" src={Logo} alt="Narvi" />
	</div>

	<div class="flex w-full max-w-sm flex-col gap-6">
		<Card.Root>
			<Card.Header class="text-center">
				<Card.Title class="text-xl">Welcome to Narvi</Card.Title>
				<Card.Description>Sign in to your account</Card.Description>
			</Card.Header>
			<Card.Content>
				{#if currentStep === 1}
					<FormEmail email={props.form?.email} />
				{:else if currentStep === 2}
					<FormCode email={props.form?.email || ''} onCancel={() => (currentStep = 1)} />
				{/if}
			</Card.Content>
		</Card.Root>
	</div>
</div>
