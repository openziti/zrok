import {BrowserRouter, Route, Routes, useNavigate} from "react-router";
import ApiConsole from "./ApiConsole.tsx";
import Login from "./Login.tsx";
import {useEffect, useState, useCallback} from "react";
import {User} from "./model/user.ts";
import useApiConsoleStore from "./model/store.ts";
import ForgotPassword from "./ForgotPassword.tsx";
import Register from "./Register.tsx";
import ResetPassword from "./ResetPassword.tsx";
import {extensionRegistry} from "./extensions/registry.ts";
import {loadExtensions} from "./extensions.config.ts";
import {Snackbar, Alert} from "@mui/material";

// Notification state for extensions
interface Notification {
    message: string;
    severity: 'info' | 'success' | 'warning' | 'error';
    open: boolean;
}

// Inner app component that has access to router context
const AppContent = () => {
    const navigate = useNavigate();
    const user = useApiConsoleStore((state) => state.user);
    const updateUser = useApiConsoleStore((state) => state.updateUser);
    const initializeExtensionStates = useApiConsoleStore((state) => state.initializeExtensionStates);
    const [extensionsLoaded, setExtensionsLoaded] = useState(false);
    const [notification, setNotification] = useState<Notification>({
        message: '',
        severity: 'info',
        open: false
    });

    // Notification function for extensions
    const notify = useCallback((message: string, severity: 'info' | 'success' | 'warning' | 'error' = 'info') => {
        setNotification({ message, severity, open: true });
    }, []);

    const handleCloseNotification = () => {
        setNotification(prev => ({ ...prev, open: false }));
    };

    // Load and initialize extensions
    useEffect(() => {
        const initExtensions = async () => {
            // Load extension manifests
            loadExtensions();

            // Initialize extension states in store
            const initialStates = extensionRegistry.getInitialStates();
            initializeExtensionStates(initialStates);

            // Initialize all extensions
            await extensionRegistry.initializeAll(navigate, notify);

            setExtensionsLoaded(true);
        };

        initExtensions();
    }, [navigate, notify, initializeExtensionStates]);

    // Check for stored user on mount
    useEffect(() => {
        const checkUser = () => {
            const storedUser = localStorage.getItem("user");
            if (storedUser) {
                updateUser(JSON.parse(storedUser));
            }
        }
        checkUser();

        document.addEventListener("userUpdated", checkUser);

        return () => {
            document.removeEventListener("userUpdated", checkUser);
        }
    }, [updateUser]);

    // Notify extensions of user changes
    useEffect(() => {
        if (!extensionsLoaded) return;

        if (user) {
            extensionRegistry.notifyUserLogin(user);
        } else {
            extensionRegistry.notifyUserLogout();
        }
    }, [user, extensionsLoaded]);

    const login = (user: User) => {
        updateUser(user);
        localStorage.setItem("user", JSON.stringify(user));
    }

    const logout = () => {
        updateUser(null as User);
        localStorage.clear();
    }

    // Get extension routes
    const extensionRoutes = extensionRegistry.getRoutes();

    const consoleRoot = user ? <ApiConsole logout={logout}/> : <Login onLogin={login}/>

    return (
        <>
            <Routes>
                <Route index element={consoleRoot}/>
                <Route path="/forgotPassword" element={<ForgotPassword />} />
                <Route path="/register/:regToken" element={<Register />} />
                <Route path="/resetPassword/:resetToken" element={<ResetPassword />} />

                {/* Extension routes */}
                {extensionRoutes.map((route) => {
                    const context = extensionRegistry.getContext(route.extensionId);
                    const RouteComponent = route.component;

                    // Handle authentication requirement (default true)
                    const requiresAuth = route.requiresAuth !== false;

                    if (requiresAuth && !user) {
                        // Redirect to login for protected routes
                        return (
                            <Route
                                key={`${route.extensionId}-${route.path}`}
                                path={route.path}
                                element={<Login onLogin={login} />}
                            />
                        );
                    }

                    return (
                        <Route
                            key={`${route.extensionId}-${route.path}`}
                            path={route.path}
                            element={
                                <RouteComponent
                                    user={user}
                                    context={context!}
                                    logout={logout}
                                />
                            }
                        />
                    );
                })}
            </Routes>

            {/* Notification snackbar for extensions */}
            <Snackbar
                open={notification.open}
                autoHideDuration={6000}
                onClose={handleCloseNotification}
                anchorOrigin={{ vertical: 'bottom', horizontal: 'right' }}
            >
                <Alert
                    onClose={handleCloseNotification}
                    severity={notification.severity}
                    variant="filled"
                >
                    {notification.message}
                </Alert>
            </Snackbar>
        </>
    );
}

const App = () => {
    return (
        <BrowserRouter>
            <AppContent />
        </BrowserRouter>
    );
}

export default App;
