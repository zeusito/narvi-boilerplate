<script lang="ts">
	import Logo from '$lib/assets/narvi-logo.png';
</script>

<div class="flex min-h-full flex-col items-center justify-center gap-2 bg-muted p-6 md:p-10">
	<div class="flex w-full max-w-sm flex-col gap-6">
		<img class="mx-auto h-15 w-auto" src={Logo} alt="Narvi" />
	</div>

	<div class="grid min-h-full place-items-center px-6 py-6">
		<div class="text-center">
			<h1 class="mt-4 text-lg font-semibold tracking-tight text-balance text-red-800">
				You have been logged out successfully.
			</h1>
			<p class="mt-6 text-lg font-medium text-pretty text-gray-500 sm:text-xl/8">
				You can now close this window or return to the home page.
			</p>
		</div>
	</div>
</div>
