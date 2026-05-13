import {Root} from "./environment";
import {
    Share,
    ShareRequest,
    AccessRequest,
    Access,
    PROXY_BACKEND_MODE,
    PUBLIC_SHARE_MODE,
    ShareMode,
} from "./model";
import {createShare, deleteShare, releaseReservedShare} from "./share";
import {createAccess, deleteAccess} from "./access";
import {getOverview} from "./overview";

export async function withShare<T>(
    root: Root,
    request: ShareRequest,
    fn: (share: Share) => Promise<T>,
): Promise<T> {
    const shr = await createShare(root, request);
    try {
        return await fn(shr);
    } finally {
        if (!request.reserved) {
            await deleteShare(root, shr);
        }
    }
}

export async function withAccess<T>(
    root: Root,
    request: AccessRequest,
    fn: (access: Access) => Promise<T>,
): Promise<T> {
    const acc = await createAccess(root, request);
    try {
        return await fn(acc);
    } finally {
        await deleteAccess(root, acc);
    }
}

export interface ProxyShareOptions {
    shareMode?: ShareMode;
    uniqueName?: string;
    frontends?: string[];
    verifySsl?: boolean;
}

export class ProxyShare {
    root: Root;
    share: Share;
    private _cleanupRegistered: boolean = false;

    private constructor(root: Root, share: Share) {
        this.root = root;
        this.share = share;
    }

    static async create(root: Root, target: string, options?: ProxyShareOptions): Promise<ProxyShare> {
        const shareMode = options?.shareMode || PUBLIC_SHARE_MODE;
        const uniqueName = options?.uniqueName;

        if (uniqueName) {
            const overview = await getOverview(root);
            if (overview.environments) {
                for (const envRes of overview.environments) {
                    if (envRes.environment?.zId === root.environment?.zId) {
                        const shares = envRes.shares || [];
                        for (const s of shares) {
                            if (s.shareToken === uniqueName) {
                                const existingShare = new Share(
                                    s.shareToken!,
                                    s.frontendEndpoints || [],
                                );
                                return new ProxyShare(root, existingShare);
                            }
                        }
                    }
                }
            }
        }

        const request = new ShareRequest(shareMode, PROXY_BACKEND_MODE, target);
        request.reserved = true;
        if (uniqueName) {
            request.uniqueName = uniqueName;
        }
        if (options?.frontends) {
            request.frontends = options.frontends;
        }

        const shr = await createShare(root, request);
        const instance = new ProxyShare(root, shr);

        if (!uniqueName) {
            instance.registerCleanup();
        }

        return instance;
    }

    get token(): string {
        return this.share.shareToken;
    }

    get endpoints(): string[] | undefined {
        return this.share.frontendEndpoints;
    }

    registerCleanup(): void {
        if (!this._cleanupRegistered) {
            process.on("beforeExit", async () => {
                try {
                    await this.cleanup();
                } catch {
                    // ignore errors during cleanup
                }
            });
            this._cleanupRegistered = true;
        }
    }

    async cleanup(): Promise<void> {
        await releaseReservedShare(this.root, this.share);
    }
}
