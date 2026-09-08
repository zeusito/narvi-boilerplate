<script lang="ts">
	import { fade } from 'svelte/transition';
	import { enhance } from '$app/forms';
	import { Button } from '$lib/components/ui/button';
	import {
		Field,
		FieldGroup,
		FieldLabel,
		FieldDescription,
		FieldSeparator
	} from '$lib/components/ui/field';
	import { Input } from '$lib/components/ui/input';
	import { Spinner } from '$lib/components/ui/spinner';

	interface Props {
		email?: string;
	}

	let props: Props = $props();
	let working = $state(false);
</script>

<div in:fade={{ duration: 200 }}>
	<form
		class="space-y-6"
		method="POST"
		action="?/sendOTP"
		use:enhance={() => {
			working = true;

			return async ({ update }) => {
				await update();
				working = false;
			};
		}}
	>
		<FieldGroup>
			<Field>
				<Button
					variant="outline"
					class="w-full"
					href="/oauth/google"
					data-sveltekit-preload-data="off"
				>
					<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">
						<path
							d="M12.48 10.92v3.28h7.84c-.24 1.84-.853 3.187-1.787 4.133-1.147 1.147-2.933 2.4-6.053 2.4-4.827 0-8.6-3.893-8.6-8.72s3.773-8.72 8.6-8.72c2.6 0 4.507 1.027 5.907 2.347l2.307-2.307C18.747 1.44 16.133 0 12.48 0 5.867 0 .307 5.387.307 12s5.56 12 12.173 12c3.573 0 6.267-1.173 8.373-3.36 2.16-2.16 2.84-5.213 2.84-7.667 0-.76-.053-1.467-.173-2.053H12.48z"
							fill="currentColor"
						/>
					</svg>
					Login with Google
				</Button>
			</Field>
			<FieldSeparator class="*:data-[slot=field-separator-content]:bg-card">
				Or continue with
			</FieldSeparator>
			<Field>
				<FieldLabel for="email">Email</FieldLabel>
				<Input
					id="email"
					name="email"
					type="email"
					placeholder="m@example.com"
					required
					value={props.email || ''}
				/>
			</Field>
			<Field>
				<Button type="submit" class="w-full" disabled={working}>
					{#if working}
						<Spinner />
					{/if}
					Continue
				</Button>
				<FieldDescription class="text-center text-xs">
					Don't have an account? Contact an administrator.
				</FieldDescription>
			</Field>
		</FieldGroup>
	</form>
</div>
