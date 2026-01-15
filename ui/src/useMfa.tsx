import { useState, useCallback, ReactNode } from "react";
import { User } from "./model/user.ts";
import { getAccountApi } from "./model/api.ts";
import MfaChallengeModal from "./MfaChallengeModal.tsx";

interface UseMfaResult {
    isMfaEnabled: () => Promise<boolean>;
    requireMfaChallenge: () => Promise<string | null>;
    MfaChallengeRenderer: () => ReactNode;
}

interface MfaState {
    showModal: boolean;
    resolve: ((token: string | null) => void) | null;
}

/**
 * Hook for MFA step-up authentication.
 *
 * This hook provides methods for extensions or other parts of the app
 * to check MFA status and require MFA re-assertion before performing
 * sensitive operations.
 *
 * Usage:
 * ```tsx
 * const { isMfaEnabled, requireMfaChallenge, MfaChallengeRenderer } = useMfa(user);
 *
 * // Check if MFA is enabled
 * const enabled = await isMfaEnabled();
 *
 * // Require MFA verification before a sensitive operation
 * const challengeToken = await requireMfaChallenge();
 * if (challengeToken) {
 *     // User verified - proceed with operation
 *     // Pass challengeToken to the backend for verification
 * } else {
 *     // User cancelled or MFA not enabled
 * }
 *
 * // Include the renderer in your component
 * return (
 *     <>
 *         {MfaChallengeRenderer()}
 *         ... rest of your UI
 *     </>
 * );
 * ```
 */
export function useMfa(user: User): UseMfaResult {
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
            // First check if MFA is enabled
            isMfaEnabled().then((enabled) => {
                if (!enabled) {
                    // MFA not enabled - resolve with null immediately
                    resolve(null);
                    return;
                }

                // Show the challenge modal
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

    const MfaChallengeRenderer = useCallback((): ReactNode => {
        return (
            <MfaChallengeModal
                isOpen={mfaState.showModal}
                user={user}
                onSuccess={handleSuccess}
                onCancel={handleCancel}
            />
        );
    }, [mfaState.showModal, user, handleSuccess, handleCancel]);

    return {
        isMfaEnabled,
        requireMfaChallenge,
        MfaChallengeRenderer
    };
}

export default useMfa;
