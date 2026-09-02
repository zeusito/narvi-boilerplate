# Narvi Boilerplate

This project provides a highly opinionated monorepo boilerplate for building solid, maintainable, and scalable micro SaaS applications. 

It is designed to facilitate collaboration between human developers and AI agents, enabling a seamless development experience.

## Project Structure

The project is organized into two main components:

### Backend API

A Go powered API that serves as the backbone. It is responsible for handling the business logic of the application and providing data to the frontend.

### Frontend

A SvelteKit powered web application that serves as the user interface. It acts as a BFF (Backend For Frontend) and it is the primary consumer of the backend API.


## Running the project locally

In order to run the development server, you can use the provided `Makefile`:

```bash
make dev
```

This will start both the backend and frontend development servers concurrently.

## Development Workflow (OpenSpec)

This project strictly follows a **Spec-Driven Development** workflow powered by **OpenSpec**. All non-trivial changes, feature implementations, and architectural modifications must be planned, specified, and tracked through OpenSpec before code is written.

### Specification & Planning Structure

- **Main Specifications (`openspec/specs/<capability>/spec.md`):** Authoritative, source-of-truth capability requirements and behavior scenarios.
- **Active Changes (`openspec/changes/<change-name>/`):** Self-contained change directories holding:
  - `proposal.md`: Context, motivation, and cross-codebase impact.
  - `specs/<capability>/spec.md`: Delta specifications defining added, modified, or removed requirements.
  - `design.md`: Technical decisions, sequence diagrams, and architecture trade-offs.
  - `tasks.md`: Ordered, testable implementation tasks.
- **Archive (`openspec/changes/archive/`):** Historical record of applied and synced changes.

### Standard Workflow Commands

When collaborating with AI agents or developing features:

1. **Explore (`/opsx-explore`):** Brainstorm ideas, investigate the codebase, and evaluate technical trade-offs without writing code.
2. **Propose (`/opsx-propose <change-name>`):** Scaffold a change and generate proposal, delta spec, design document, and task breakdown in one step.
3. **Apply (`/opsx-apply <change-name>`):** Work through `tasks.md` sequentially, keeping changes minimal, focused, and tested.
4. **Archive (`/opsx-archive <change-name>`):** Merge delta specs into the main capability specs and archive the completed change.

## Contributing

Please follow the guidelines present in each folder's `AGENTS.md` file to ensure code quality and consistency

