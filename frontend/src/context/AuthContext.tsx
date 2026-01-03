'use client';

import { createContext, useContext, useEffect, useState, ReactNode } from 'react';
import { useRouter, usePathname } from 'next/navigation';

interface User {
    id: string;
    email: string;
    full_name: string;
    role: string;
    cabinet_id: string;
    onboarding_completed?: boolean;
}

interface AuthContextType {
    user: User | null;
    token: string | null;
    isLoading: boolean;
    login: (token: string, user: User) => void;
    logout: () => void;
    isAuthenticated: boolean;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function AuthProvider({ children }: { children: ReactNode }) {
    const [user, setUser] = useState<User | null>(null);
    const [token, setToken] = useState<string | null>(null);
    const [isLoading, setIsLoading] = useState(true);
    const router = useRouter();
    const pathname = usePathname();

    // Init: Check localStorage and Validate with API
    useEffect(() => {
        const initAuth = async () => {
            const storedToken = localStorage.getItem('fiducia_token');
            const storedUser = localStorage.getItem('fiducia_user');

            if (storedToken) {
                setToken(storedToken);
                // Optimistic set
                if (storedUser) setUser(JSON.parse(storedUser));

                // Verify and Refresh User Data
                try {
                    const res = await fetch('http://localhost:8080/api/v1/auth/me', {
                        headers: { 'Authorization': `Bearer ${storedToken}` }
                    });
                    if (res.ok) {
                        const data = await res.json();
                        setUser(data);
                        localStorage.setItem('fiducia_user', JSON.stringify(data));
                    }
                } catch (e) {
                    console.error("Auth validation failed", e);
                }
            }
            setIsLoading(false);
        };
        initAuth();
    }, []);

    // Protected Route Guard
    useEffect(() => {
        if (isLoading) return;

        const publicRoutes = ['/login', '/register', '/'];
        const isPublicRoute = publicRoutes.includes(pathname);

        if (!token && !isPublicRoute) {
            // Redirect to login if trying to access protected route
            router.push('/login');
        } else if (token) {
            if (isPublicRoute && pathname !== '/') {
                // Redirect from login/register to inside
                if (user?.onboarding_completed === false) {
                    router.push('/onboarding');
                } else {
                    router.push('/dashboard');
                }
            } else if (pathname !== '/onboarding' && user?.onboarding_completed === false) {
                // Force onboarding if not completed
                router.push('/onboarding');
            }
        }
    }, [isLoading, token, pathname, router, user]);

    const login = (newToken: string, newUser: User) => {
        setToken(newToken);
        setUser(newUser);
        localStorage.setItem('fiducia_token', newToken);
        localStorage.setItem('fiducia_user', JSON.stringify(newUser));

        if (newUser.onboarding_completed === false) {
            router.push('/onboarding');
        } else {
            router.push('/dashboard');
        }
    };

    const logout = () => {
        setToken(null);
        setUser(null);
        localStorage.removeItem('fiducia_token');
        localStorage.removeItem('fiducia_user');
        router.push('/login');
    };

    return (
        <AuthContext.Provider value={{
            user,
            token,
            isLoading,
            login,
            logout,
            isAuthenticated: !!token
        }}>
            {children}
        </AuthContext.Provider>
    );
}

export function useAuth() {
    const context = useContext(AuthContext);
    if (context === undefined) {
        throw new Error('useAuth must be used within an AuthProvider');
    }
    return context;
}
