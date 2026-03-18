import {ResponseError, FetchError} from "../api/runtime";

export function isAbortError(e: unknown): boolean {
    if (e instanceof DOMException && e.name === "AbortError") return true;
    if (e instanceof FetchError && e.cause?.name === "AbortError") return true;
    return false;
}

export async function extractErrorMessage(e: unknown, fallback = "an error occurred"): Promise<string> {
    if (e instanceof ResponseError) {
        try {
            const body = await e.response.json();
            if (body?.message) return body.message;
        } catch {
            // response wasn't JSON
        }
        return `request failed (${e.response.status})`;
    }
    if (e instanceof FetchError) {
        return "unable to contact the server";
    }
    if (e instanceof Error) {
        return e.message;
    }
    return fallback;
}
