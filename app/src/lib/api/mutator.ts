export type MutatorOptions = RequestInit & { baseUrl?: string };

export async function customFetch<T>(
	url: string,
	options?: MutatorOptions,
): Promise<T> {
	const { baseUrl = "", ...fetchOptions } = options ?? {};
	const res = await fetch(`${baseUrl}${url}`, fetchOptions);
	const body = [204, 205, 304].includes(res.status) ? null : await res.text();
	const data = body ? JSON.parse(body) : {};
	return { data, status: res.status, headers: res.headers } as T;
}
