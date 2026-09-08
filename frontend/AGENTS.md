# Agent Rules

These rules apply to all tasks and code changes made within this project.

## Tech Stack

This project is built using the following technologies:

- **Framework**: SvelteKit 2 + Svelte 5 (using runes)
- **Language**: TypeScript
- **Styling**: Tailwind CSS v4
- **UI Components**: [shadcn-svelte](https://www.shadcn-svelte.com/)
- **Validation**: Zod
- **Testing**: Vitest + Vitest Browser Svelte

## Skills

This project uses skills installed from [skills.sh](https://skills.sh).
In order to install skills, use the following command:

```bash
pnpm dlx skills add <skill-name>
```

## Component library rules

This project uses shadcn-svelte for UI components. There is a skill installed for this. In order to add a new component, use the following command:

```bash
pnpm dlx shadcn-svelte@latest add <component-name>
```

Adhere to the following guidelines when working with components:

- Before adding a new component, make sure it's not already installed. Use the skill command to check for existing components.
- Always try to use shadcn-svelte components when possible, if you can't find a component that suits your needs, then it's ok to implement it yourself, but make sure it is well documented and reusable.
- Never modify the appearance of a component in a way that breaks its intended purpose or accessibility.

## Code Quality & Best Practices

1. **State Management (Svelte 5 Runes)**:
   - Use `$state()`, `$derived()`, `$props()`, and `$effect()` runes exclusively for state and reactivity. Avoid legacy store APIs (`writable`, `derived`) unless interacting with external libraries that require them.
   - Use `$derived()` for computed state. Limit `$effect()` usage to synchronization with external systems or side-effects, keeping them minimal and well-documented.

2. **Communication with backend API**:
   - All communication with the backend API must reside in the `src/lib/server/gateway/` directory. All calls to the backend API must be done through this directory.
   - Do not call the backend API directly from the client side. If you need to make a call to the backend API, then you must do it through the gateway.
   - In the gateway, all functions must be async and return a `Promise<Result<T, E = Error>>`. The interface is defined at `src/lib/server/gateway/common.ts`. All responses from the backend API must be parsed and returned through this interface.

3. **Error Handling & User Feedback**:
   - SvelteKit Actions should return typed failure structures using `fail(...)` from `@sveltejs/kit` on validation or runtime errors.
   - Implement try/catch blocks in Actions/endpoints and log errors on the server side before returning a sanitized fail state to the client.
   - Never show raw internal exceptions or database error messages directly in the client UI.

4. **TypeScript and Type Safety**:
   - Write fully typed code. Avoid using `any` or loose types (unless absolutely necessary).
   - Ensure SvelteKit endpoints and load functions utilize typing constructs correctly (e.g., `satisfies Actions` to preserve type safety).
   - Use explicit validation schema parsing output types where applicable.

5. **Data Serialization and Protection**:
   - Ensure only safe, serialized data is returned from the server `load` functions or API endpoints to the client.
   - Never leak database connection strings, credentials, raw database models/records, or sensitive user fields (e.g., password hashes) to the client.

6. **Validation**:
   - Validate and sanitize all user input on the server side before performing database operations, use Zod schemas to validate and sanitize user input. you can find some examples in `$lib/models/`.

7. **Multi-tenancy & Routing Structure**:
   - We use path-based multi-tenancy with a decoupled B2B2B structure under the `src/routes/(private)/` route group:
     - **Organization Routes (`/org/[slug]/...`)**: Administrative tenant governance, billing, and team rosters (e.g., `/org/[slug]/team`, `/org/[slug]/workspaces`).
     - **Workspace Routes (`/workspace/[slug]/...`)**: Isolated operational project containers for campaigns and content (e.g., `/workspace/[slug]/library`).
     - **Triage & Standalone Routes**: `/home` for workspace-less / multi-tenant triage, and `/invitations` for reviewing pending invites.
     - **Public Auth Routes (`src/routes/auth/`)**: `/auth/login`, `/auth/magic`, `/auth/logout`.
   - Server `load` functions and form actions should extract and validate `params.slug` corresponding to the active route segment.

8. **Linting, Formatting, Checking & Tests**:
   - Always run `pnpm format` before committing.
   - Always run `pnpm lint` before committing.
   - Always run `pnpm check` before committing.
   - Always run `pnpm test` before committing.

You are able to use the Svelte MCP server, where you have access to comprehensive Svelte 5 and SvelteKit documentation. Here's how to use the available tools effectively:

## Available Svelte MCP Tools:

### 1. list-sections

Use this FIRST to discover all available documentation sections. Returns a structured list with titles, use_cases, and paths.
When asked about Svelte or SvelteKit topics, ALWAYS use this tool at the start of the chat to find relevant sections.

### 2. get-documentation

Retrieves full documentation content for specific sections. Accepts single or multiple sections.
After calling the list-sections tool, you MUST analyze the returned documentation sections (especially the use_cases field) and then use the get-documentation tool to fetch ALL documentation sections that are relevant for the user's task.

### 3. svelte-autofixer

Analyzes Svelte code and returns issues and suggestions.
You MUST use this tool whenever writing Svelte code before sending it to the user. Keep calling it until no issues or suggestions are returned.

### 4. playground-link

Generates a Svelte Playground link with the provided code.
After completing the code, ask the user if they want a playground link. Only call this tool after user confirmation and NEVER if code was written to files in their project.
