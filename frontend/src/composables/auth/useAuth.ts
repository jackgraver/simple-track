import { ref, computed } from 'vue';
import { useRouter } from 'vue-router';
import { apiClient } from '~/utils/axios';

const TOKEN_KEY = 'auth_token';
const USERNAME_KEY = 'auth_username';

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

const token = ref<string | null>(null);
const username = ref<string | null>(null);
const isAuthenticated = computed(() => !!token.value);

function loadTokenFromStorage() {
    if (typeof window === 'undefined') return;
    
    const storedToken = sessionStorage.getItem(TOKEN_KEY);
    const storedUsername = sessionStorage.getItem(USERNAME_KEY);
    
    if (storedToken && storedUsername) {
        token.value = storedToken;
        username.value = storedUsername;
    }
}

function saveTokenToStorage(t: string, u: string) {
    if (typeof window === 'undefined') return;
    
    sessionStorage.setItem(TOKEN_KEY, t);
    sessionStorage.setItem(USERNAME_KEY, u);
    token.value = t;
    username.value = u;
}

function clearTokenFromStorage() {
    if (typeof window === 'undefined') return;
    
    sessionStorage.removeItem(TOKEN_KEY);
    sessionStorage.removeItem(USERNAME_KEY);
    token.value = null;
    username.value = null;
}

export function useAuth() {
    const router = useRouter();
    
    loadTokenFromStorage();
    
    async function login(usernameInput: string, password: string): Promise<void> {
        try {
            const response = await apiClient.post<LoginResponse>('/auth/login', {
                username: usernameInput,
                password,
            });
            
            if (response.data.token) {
                saveTokenToStorage(response.data.token, response.data.username);
            } else {
                throw new Error('No token received from server');
            }
        } catch (error: any) {
            const message = error.response?.data?.error || error.message || 'Login failed';
            throw new Error(message);
        }
    }
    
    async function register(usernameInput: string, password: string, email?: string): Promise<void> {
        try {
            const response = await apiClient.post<RegisterResponse>('/auth/register', {
                username: usernameInput,
                password,
                email,
            });
            
            if (response.data.token) {
                saveTokenToStorage(response.data.token, response.data.username);
            } else {
                throw new Error('No token received from server');
            }
        } catch (error: any) {
            const message = error.response?.data?.error || error.message || 'Registration failed';
            throw new Error(message);
        }
    }
    
    async function logout() {
        try {
            await apiClient.post('/auth/logout');
        } catch (error) {
            console.error('Logout error:', error);
        } finally {
            clearTokenFromStorage();
            router.push('/signin');
        }
    }
    
    function getToken(): string | null {
        return token.value;
    }
    
    function getUsername(): string | null {
        return username.value;
    }
    
    return {
        token: computed(() => token.value),
        username: computed(() => username.value),
        isAuthenticated,
        login,
        register,
        logout,
        getToken,
        getUsername,
    };
}

