import {
	APIError,
	fetchWithTimeout,
	handleFetchResponse,
	logAndHandleFetchError,
	type Result
} from './common';
import type {
	CreateOrganizationInput,
	Organization,
	OrganizationListResponse,
	UpdateOrganizationInput
} from '$lib/models/organization';

export const listOrganizations = async (
	principalToken: string
): Promise<Result<OrganizationListResponse, APIError>> => {
	try {
		const response = await fetchWithTimeout('/v1/admin/organizations', {
			method: 'GET',
			headers: {
				Authorization: `Bearer ${principalToken}`
			}
		});

		const data = await handleFetchResponse<OrganizationListResponse>(response);

		return { success: true, data };
	} catch (error) {
		const apiError = logAndHandleFetchError('OrganizationsGateway.listOrganizations', error);
		return { success: false, error: apiError };
	}
};

export const createOrganization = async (
	principalToken: string,
	payload: CreateOrganizationInput
): Promise<Result<Organization, APIError>> => {
	try {
		const response = await fetchWithTimeout('/v1/admin/organizations', {
			method: 'POST',
			headers: {
				Authorization: `Bearer ${principalToken}`
			},
			body: JSON.stringify(payload)
		});

		const data = await handleFetchResponse<Organization>(response);

		return { success: true, data };
	} catch (error) {
		const apiError = logAndHandleFetchError('OrganizationsGateway.createOrganization', error);
		return { success: false, error: apiError };
	}
};

export const updateOrganization = async (
	principalToken: string,
	payload: UpdateOrganizationInput
): Promise<Result<Organization, APIError>> => {
	try {
		const { id, ...body } = payload;
		const response = await fetchWithTimeout(`/v1/admin/organizations/${id}`, {
			method: 'PATCH',
			headers: {
				Authorization: `Bearer ${principalToken}`
			},
			body: JSON.stringify(body)
		});

		const data = await handleFetchResponse<Organization>(response);

		return { success: true, data };
	} catch (error) {
		const apiError = logAndHandleFetchError('OrganizationsGateway.updateOrganization', error);
		return { success: false, error: apiError };
	}
};
