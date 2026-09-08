<script lang="ts">
	import { fade } from 'svelte/transition';
	import { enhance } from '$app/forms';
	import { Button } from '$lib/components/ui/button';
	import { Field, FieldGroup, FieldLabel, FieldDescription } from '$lib/components/ui/field';
	import { Input } from '$lib/components/ui/input';
	import { Spinner } from '$lib/components/ui/spinner';

	interface Props {
		email: string;
		onCancel: () => void;
	}

	let props: Props = $props();
	let working = $state(false);
</script>

<div in:fade={{ duration: 200 }}>
	<form
		class="space-y-6"
		method="POST"
		action="?/validateOTP"
		use:enhance={() => {
			working = true;

			return async ({ update }) => {
				await update();
				working = false;
			};
		}}
	>
		<input type="hidden" name="email" value={props.email} />
		<FieldGroup>
			<Field>
				<FieldLabel for="otp">Verification Code</FieldLabel>
				<Input
					id="otp"
					name="otp"
					type="text"
					placeholder="123456"
					required
					maxlength={6}
					class="text-center text-2xl tracking-[0.5em]"
				/>
				<FieldDescription class="text-center">
					Enter the 6-digit code sent to
					<span class="font-medium text-foreground">
						{props.email}
					</span>
				</FieldDescription>
			</Field>
			<Field class="flex flex-col gap-2">
				<Button type="submit" class="w-full" disabled={working}>
					{#if working}
						<Spinner />
					{/if}
					Verify Code
				</Button>
				<Button variant="ghost" type="button" class="w-full" onclick={props.onCancel}>Back</Button>
			</Field>
		</FieldGroup>
	</form>
</div>
