import Cookies from 'universal-cookie';

const BACKEND_URL = import.meta.env.VITE_JURY_URL;

export async function getRequest<T>(path: string): Promise<FetchResponse<T>> {
    try {
        const options: RequestInit = {
            method: 'GET',
            headers: createHeaders(true),
            credentials: 'include',
        };
        const response = await fetch(`${BACKEND_URL}${path}`, options);
        if (!response.ok) throw new Error(response.statusText);

        const data = await response.json();
        return { status: response.status, error: data.error ? data.error : '', data };
        // eslint-disable-next-line
    } catch (error: any) {
        console.error(error);
        return { status: 404, error: error, data: null };
    }
}

export async function postRequest<T>(
    path: string,
    body: any
): Promise<FetchResponse<T>> {
    try {
        const options: RequestInit = {
            method: 'POST',
            headers: createHeaders(true),
            credentials: 'include',
            body: body ? JSON.stringify(body) : null,
        };
        const response = await fetch(`${BACKEND_URL}${path}`, options);
        if (!response.ok) throw new Error(response.statusText);

        const data = await response.json();
        return { status: response.status, error: data.error ? data.error : '', data };
        // eslint-disable-next-line
    } catch (error: any) {
        console.error(error);
        return { status: 404, error: error, data: null };
    }
}

export async function putRequest<T>(
    path: string,
    body: any
): Promise<FetchResponse<T>> {
    try {
        const options: RequestInit = {
            method: 'PUT',
            headers: createHeaders(true),
            credentials: 'include',
            body: body ? JSON.stringify(body) : null,
        };
        const response = await fetch(`${BACKEND_URL}${path}`, options);
        if (!response.ok) throw new Error(response.statusText);

        const data = await response.json();
        return { status: response.status, error: data.error ? data.error : '', data };
        // eslint-disable-next-line
    } catch (error: any) {
        console.error(error);
        return { status: 404, error: error, data: null };
    }
}

export async function deleteRequest(
    path: string
): Promise<FetchResponse<OkResponse>> {
    try {
        const options: RequestInit = {
            method: 'DELETE',
            headers: createHeaders(true),
            credentials: 'include',
        };
        const response = await fetch(`${BACKEND_URL}${path}`, options);
        if (!response.ok) throw new Error(response.statusText);

        const data = await response.json();
        return { status: response.status, error: data.error ? data.error : '', data };
        // eslint-disable-next-line
    } catch (error: any) {
        console.error(error);
        return { status: 404, error: error, data: null };
    }
}

export function createHeaders(json: boolean): Headers {
    const headers = new Headers();
    if (json) {
        headers.append('Content-Type', 'application/json');
    }
    return headers;
}
