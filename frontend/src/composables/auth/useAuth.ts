import { useRouter } from "vue-router";
import { apiClient } from "~/api/client";
import {
    authStatus,
    isAuthenticated,
    markAuthenticated,
    markUnauthorized,
    username,
} from "~/composables/auth/session";

interface LoginResponse {
    token: string;
    username: string;
    user: {
        ID: number;
        username: string;
        email?: string;
    };
}

interface RegisterResponse {
    token: string;
    username: string;
    user: {
        ID: number;
        username: string;
        email?: string;
    };
}

let resolveInFlight: Promise<void> | null = null;

export async function resolveAuthSession(): Promise<void> {
    if (authStatus.value !== "unknown") {
        return;
    }
    if (!resolveInFlight) {
        resolveInFlight = (async () => {
            try {
                const { data } = await apiClient.get<{ user?: { username: string }; username?: string }>(
                    "/auth/me",
                );
                const name = data.username ?? data.user?.username;
                if (name) {
                    markAuthenticated(name);
                } else {
                    markUnauthorized();
                }
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
            const response = await apiClient.post<LoginResponse>("/auth/login", {
                username: usernameInput,
                password,
            });

            const name = response.data.username ?? response.data.user?.username;
            if (name) {
                markAuthenticated(name);
            } else {
                throw new Error("No user in login response");
            }
        } catch (error: any) {
            const message = error.response?.data?.error || error.message || "Login failed";
            throw new Error(message);
        }
    }

    async function register(usernameInput: string, password: string, email?: string): Promise<void> {
        try {
            const response = await apiClient.post<RegisterResponse>("/auth/register", {
                username: usernameInput,
                password,
                email,
            });

            const name = response.data.username ?? response.data.user?.username;
            if (name) {
                markAuthenticated(name);
            } else {
                throw new Error("No user in registration response");
            }
        } catch (error: any) {
            const message = error.response?.data?.error || error.message || "Registration failed";
            throw new Error(message);
        }
    }

    async function logout() {
        try {
            await apiClient.post("/auth/logout");
        } catch (error) {
            console.error("Logout error:", error);
        } finally {
            markUnauthorized();
            router.push("/signin");
        }
    }

    function getUsername(): string | null {
        return username.value;
    }

    return {
        username,
        isAuthenticated,
        login,
        register,
        logout,
        getUsername,
    };
}
