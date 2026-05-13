export interface User {
    email: string;
    token: string;
}

const USER_STORAGE_KEY = "user";

const isStoredUser = (value: unknown): value is User => {
    if (!value || typeof value !== "object") {
        return false;
    }

    const candidate = value as Record<string, unknown>;
    return typeof candidate.email === "string" && typeof candidate.token === "string";
};

export const loadStoredUser = (): User | null => {
    const raw = localStorage.getItem(USER_STORAGE_KEY);
    if (!raw) {
        return null;
    }

    try {
        const parsed = JSON.parse(raw);
        if (isStoredUser(parsed)) {
            return parsed;
        }
    } catch {
        // Invalid persisted session data should not crash bootstrap.
    }

    localStorage.removeItem(USER_STORAGE_KEY);
    return null;
};

export const saveStoredUser = (user: User) => {
    localStorage.setItem(USER_STORAGE_KEY, JSON.stringify(user));
};

export const clearStoredUser = () => {
    localStorage.removeItem(USER_STORAGE_KEY);
};
