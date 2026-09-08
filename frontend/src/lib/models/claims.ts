// --- Org Roles ---
export const ORG_ROLE_OWNER = 'owner';
export const ORG_ROLE_ADMIN = 'admin';
export const ORG_ROLE_MEMBER = 'member';

export type OrgRole = typeof ORG_ROLE_OWNER | typeof ORG_ROLE_ADMIN | typeof ORG_ROLE_MEMBER;

const ORG_ROLE_HIERARCHY: Record<OrgRole, number> = {
	[ORG_ROLE_OWNER]: 300,
	[ORG_ROLE_ADMIN]: 200,
	[ORG_ROLE_MEMBER]: 100
};

export interface PrincipalClaims {
	isAuthenticated: boolean;
	token: string;
	subject: string;
	subjectName: string;
	subjectEmail: string;
	organizationId: string;
	organizationSlug: string;
	organizationName: string;
	organizationRole: string;
}

export const ANONYMOUS_CLAIMS: PrincipalClaims = {
	isAuthenticated: false,
	token: '',
	subject: '',
	subjectName: '',
	subjectEmail: '',
	organizationId: '',
	organizationSlug: '',
	organizationName: '',
	organizationRole: ''
};

export class ClaimsService {
	private readonly claims: PrincipalClaims;

	constructor(claims: PrincipalClaims) {
		this.claims = claims;
	}

	public getClaims(): PrincipalClaims {
		return this.claims;
	}

	public isAuthenticated(): boolean {
		return this.claims.isAuthenticated;
	}
}
