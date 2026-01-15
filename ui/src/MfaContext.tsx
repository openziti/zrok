import { createContext, useContext, useState, useCallback, ReactNode } from "react";
import { User } from "./model/user.ts";
import { getAccountApi } from "./model/api.ts";
import MfaChallengeModal from "./MfaChallengeModal.tsx";

interface MfaContextValue {
    /**
     * Check if MFA is enabled for the current user
     */
    isMfaEnabled: () => Promise<boolean>;

    /**
     * Require MFA verification before proceeding.
     * Shows the MFA challenge modal and returns a challenge token on success.
     *
     * @returns Challenge token if verified, null if cancelled or MFA not enabled
     */
    requireMfaChallenge: () => Promise<string | null>;
}

const MfaContext = createContext<MfaContextValue | null>(null);

interface MfaProviderProps {
    user: User;
    children: ReactNode;
}

interface MfaState {
    showModal: boolean;
    resolve: ((token: string | null) => void) | null;
}

/**
 * Provider component for MFA functionality.
 *
 * Wrap your app or a section of your app with this provider to enable
 * MFA step-up authentication.
 *
 * Usage:
 * ```tsx
 * <MfaProvider user={user}>
 *     <YourApp />
 * </MfaProvider>
 * ```
 */
export function MfaProvider({ user, children }: MfaProviderProps) {
    const [mfaState, setMfaState] = useState<MfaState>({
        showModal: false,
        resolve: null
    });

    const isMfaEnabled = useCallback(async (): Promise<boolean> => {
        try {
            const status = await getAccountApi(user).mfaStatus();
            return status.enabled || false;
        } catch (e) {
            console.error("Failed to check MFA status:", e);
            return false;
        }
    }, [user]);

    const requireMfaChallenge = useCallback((): Promise<string | null> => {
        return new Promise((resolve) => {
            isMfaEnabled().then((enabled) => {
                if (!enabled) {
                    resolve(null);
                    return;
                }

                setMfaState({
                    showModal: true,
                    resolve: resolve
                });
            });
        });
    }, [isMfaEnabled]);

    const handleSuccess = useCallback((challengeToken: string) => {
        if (mfaState.resolve) {
            mfaState.resolve(challengeToken);
        }
        setMfaState({ showModal: false, resolve: null });
    }, [mfaState.resolve]);

    const handleCancel = useCallback(() => {
        if (mfaState.resolve) {
            mfaState.resolve(null);
        }
        setMfaState({ showModal: false, resolve: null });
    }, [mfaState.resolve]);

    const contextValue: MfaContextValue = {
        isMfaEnabled,
        requireMfaChallenge
    };

    return (
        <MfaContext.Provider value={contextValue}>
            {children}
            <MfaChallengeModal
                isOpen={mfaState.showModal}
                user={user}
                onSuccess={handleSuccess}
                onCancel={handleCancel}
            />
        </MfaContext.Provider>
    );
}

/**
 * Hook to access MFA functionality from the context.
 *
 * Usage:
 * ```tsx
 * const { isMfaEnabled, requireMfaChallenge } = useMfaContext();
 *
 * const handleSensitiveAction = async () => {
 *     const token = await requireMfaChallenge();
 *     if (token) {
 *         // User verified - proceed with sensitive action
 *         await doSensitiveAction({ mfaChallengeToken: token });
 *     }
 * };
 * ```
 */
export function useMfaContext(): MfaContextValue {
    const context = useContext(MfaContext);
    if (!context) {
        throw new Error("useMfaContext must be used within an MfaProvider");
    }
    return context;
}

export default MfaContext;
