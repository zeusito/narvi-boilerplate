export interface PrincipalClaims {
	isAuthenticated: boolean;
	token: string;
	identityId: string;
	fullName: string;
	Email: string;
	activeOrganizationId: string;
	organizationSlug: string;
	organizationName: string;
	organizationKind: string;
	organizationRole: string;
}

export const ANONYMOUS_CLAIMS: PrincipalClaims = {
	isAuthenticated: false,
	token: '',
	identityId: '',
	fullName: '',
	Email: '',
	activeOrganizationId: '',
	organizationSlug: '',
	organizationName: '',
	organizationKind: '',
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

	public isManagementAdmin(): boolean {
		return (
			this.claims.isAuthenticated &&
			this.claims.organizationKind === 'management' &&
			this.claims.organizationRole === 'admin'
		);
	}
}
