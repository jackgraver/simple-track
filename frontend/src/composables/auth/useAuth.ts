import { useRouter } from "vue-router";
import { apiClient } from "~/api/client";
import { queryClient } from "~/api/queryClient";
import { getApiErrorMessage } from "~/api/httpError";
import {
    authStatus,
    isAdmin,
    isAuthenticated,
    markAuthenticated,
    markUnauthorized,
    setEnvironment,
    username,
} from "~/composables/auth/session";
import { clearUserStorage } from "~/utils/userStorage";

export interface AuthResponse {
    username?: string;
    environment?: string;
    user?: {
        ID: number;
        username: string;
        email?: string;
        is_admin?: boolean;
    };
}

/** Applies a successful /auth/me, /auth/login or /auth/setup response to the session. */
export function applyAuthResponse(data: AuthResponse): boolean {
    const name = data.username ?? data.user?.username;
    if (!name) return false;
    setEnvironment(data.environment);
    markAuthenticated(name, data.user?.ID, data.user?.is_admin === true);
    return true;
}

let resolveInFlight: Promise<void> | null = null;

export async function resolveAuthSession(): Promise<void> {
    if (authStatus.value !== "unknown") {
        return;
    }
    if (!resolveInFlight) {
        resolveInFlight = (async () => {
            try {
                const { data } = await apiClient.get<AuthResponse>("/auth/me");
                if (!applyAuthResponse(data)) markUnauthorized();
            } catch {
                markUnauthorized();
            }
        })();
    }
    await resolveInFlight;
}

export function useAuth() {
    const router = useRouter();

    async function login(usernameInput: string, password: string): Promise<void> {
        try {
            const response = await apiClient.post<AuthResponse>("/auth/login", {
                username: usernameInput,
                password,
            });

            if (!applyAuthResponse(response.data)) throw new Error("No user in login response");
        } catch (error) {
            throw new Error(getApiErrorMessage(error, "Login failed"));
        }
    }

    async function register(usernameInput: string, password: string, email?: string): Promise<void> {
        try {
            const response = await apiClient.post<AuthResponse>("/auth/register", {
                username: usernameInput,
                password,
                email,
            });

            if (!applyAuthResponse(response.data)) throw new Error("No user in registration response");
        } catch (error) {
            throw new Error(getApiErrorMessage(error, "Registration failed"));
        }
    }

    /** Resolves false when the server didn't confirm; the session is cleared locally either way. */
    async function logout(): Promise<boolean> {
        try {
            await apiClient.post("/auth/logout");
            return true;
        } catch (error) {
            console.error("Logout error:", error);
            return false;
        } finally {
            queryClient.clear();
            clearUserStorage();
            markUnauthorized();
            void router.push({ name: "signin" });
        }
    }

    function getUsername(): string | null {
        return username.value;
    }

    return {
        username,
        isAdmin,
        isAuthenticated,
        login,
        register,
        logout,
        getUsername,
    };
}
