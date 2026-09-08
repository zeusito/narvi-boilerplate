import { z } from 'zod/v4';

export const OrganizationKindSchema = z.enum(['standard', 'management']);
export type OrganizationKind = z.infer<typeof OrganizationKindSchema>;

export const OrganizationStateSchema = z.enum(['active', 'suspended', 'deleted']);
export type OrganizationState = z.infer<typeof OrganizationStateSchema>;

export const CreateOrganizationSchema = z.object({
	name: z.string().trim().min(3).max(255),
	kind: OrganizationKindSchema
});

export const UpdateOrganizationSchema = z.object({
	id: z.string().min(1).max(255),
	name: z.string().trim().min(3).max(255),
	logo: z.string().trim().max(500).optional(),
	state: z.enum(['active', 'suspended'])
});

export type CreateOrganizationInput = z.infer<typeof CreateOrganizationSchema>;
export type UpdateOrganizationInput = z.infer<typeof UpdateOrganizationSchema>;

export interface Organization {
	id: string;
	name: string;
	slug: string;
	kind: OrganizationKind;
	logo: string;
	state: OrganizationState;
	createdAt: string;
	updatedAt: string;
}

export interface OrganizationListResponse {
	count: number;
	data: Organization[];
}
